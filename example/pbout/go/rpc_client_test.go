package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/netpoll"
	"github.com/cloudwego/netpoll/mux"
	"github.com/fixkme/gokit/rpc"
	g "github.com/fixkme/gokit/util/go"
	"github.com/fixkme/protoc-gen-gom/example/pbout/go/gate"
	"github.com/panjf2000/gnet/v2"
	"google.golang.org/protobuf/proto"
)

func TestClient(t *testing.T) {
	netpollTest()
}

func gnetClient() {
	h := &rpc.ClientHander{}
	gc, err := gnet.NewClient(h, gnet.WithMulticore(false))
	if err != nil {
		log.Fatalf("new client error: %v", err)
	}
	if err = gc.Start(); err != nil {
		log.Fatalf("start client error: %v", err)
	}
	c1, err := gc.Dial("tcp", "127.0.0.1:2333")
	if err != nil {
		log.Fatalf("dial error: %v", err)
	}
	cs1 := c1.Context().(*rpc.ConnState)
	c2, err := gc.Dial("tcp", "127.0.0.1:2333")
	if err != nil {
		log.Fatalf("dial error: %v", err)
	}
	cs2 := c2.Context().(*rpc.ConnState)

	logicFn := func(cs *rpc.ConnState, sync bool) {
		opt := &rpc.CallOption{Sync: sync}
		for i := 0; i < 5; i++ {
			rsp := &gate.SNoticePlayer{}
			if err := rpc.Invoke(cs, context.Background(), "Gate/NoticePlayer", &gate.CNoticePlayer{PlayerId: int64(i)}, rsp, opt); err != nil {
				log.Printf("invoke error: %v\n", err)
			}
			fmt.Printf("call rsp:%v\n", rsp)
			time.Sleep(time.Microsecond * time.Duration(rand.Intn(1000)))
		}
	}
	go logicFn(cs1, true)
	go logicFn(cs2, true)
	go logicFn(cs1, true)

	select {
	case <-time.After(20 * time.Second):
	}
}

func netpollTest() {
	opt := &rpc.ClientOpt{DailTimeout: time.Second}
	cliConn, err := rpc.NewClientConn("tcp", "127.0.0.1:2333", opt)
	if err != nil {
		log.Fatalf("new client conn error: %v\n", err)
	}

	logicFn := func(id int, cs *rpc.ClientConn, sync bool) {
		opt := &rpc.CallOption{Sync: sync}
		if !sync {
			asyncRetCh := make(chan *rpc.AsyncCallResult, 10)
			opt.AsyncRetChan = asyncRetCh
			go func() {
				for ret := range asyncRetCh {
					fmt.Printf("%d async call ret:%v\n", id, ret)
				}
			}()
		}

		for i := 0; i < 5; i++ {
			rsp := &gate.SNoticePlayer{}
			if err := cs.Invoke(context.Background(), "Gate", "NoticePlayer", &gate.CNoticePlayer{PlayerId: int64(i + 1)}, rsp, opt); err != nil {
				log.Printf("invoke error: %v\n", err)
			} else if sync {
				fmt.Printf("%d sync call rsp:%v\n", id, rsp)
			}
			time.Sleep(time.Microsecond * time.Duration(rand.Intn(1000)))
		}
	}
	go logicFn(1, cliConn, false)
	//go logicFn(2, cliConn, false)
	// go logicFn(cliConn, true)

	select {
	case <-time.After(time.Second * 10):
	}
}
func (cli *CliConn) Call(path string, data proto.Message) (err error) {
	v2 := strings.SplitN(path, "/", 2)
	if len(v2) != 2 {
		log.Fatalf("invalid path: %s", path)
	}
	payload, err := proto.Marshal(data)
	if err != nil {
		log.Fatalf("failed to marshal request: %v", err)
	}
	rpcReq := &rpc.RpcRequestMessage{
		ServiceName: v2[0],
		MethodName:  v2[1],
		Payload:     payload,
	}
	// encode
	writer := netpoll.NewLinkBuffer()
	err = Encode(writer, rpcReq)
	if err != nil {
		log.Printf("Failed to encode: %v\n", err)
		return err
	}
	cli.wqueue.Add(func() (buf netpoll.Writer, isNil bool) {
		fmt.Printf("GoroutineID %d worker add send data\n", g.GoroutineID())
		return writer, false
	})
	fmt.Printf("GoroutineID %d send ok %d\n", g.GoroutineID(), data.(*gate.CNoticePlayer).PlayerId)
	// decode
	reader := <-cli.rch
	resp := &rpc.RpcResponseMessage{}
	err = Decode(reader, resp)
	if err != nil {
		log.Printf("Failed to decode: %v\n", err)
		return err
	}
	fmt.Printf("GoroutineID %d logic recv rpcResp:%v\n", g.GoroutineID(), resp)
	return nil
}

func newCliConn(conn netpoll.Connection) *CliConn {
	mc := &CliConn{}
	mc.conn = conn
	mc.rch = make(chan netpoll.Reader, 1)
	// loop read
	conn.SetOnRequest(func(ctx context.Context, connection netpoll.Connection) error {
		fmt.Printf("start recv data\n")
		reader := connection.Reader()
		// decode
		bLen, err := reader.Peek(4)
		if err != nil {
			return err
		}
		l := int(binary.LittleEndian.Uint32(bLen))
		r, err := reader.Slice(l + 4)
		if err != nil {
			log.Printf("Failed to slice: %v\n", err)
			return err
		}
		fmt.Printf("GoroutineID %d io recv data:%v\n", g.GoroutineID(), l+4)
		mc.rch <- r
		return nil
	})
	conn.AddCloseCallback(func(connection netpoll.Connection) error {
		fmt.Printf("[%v] connection closed\n", connection.RemoteAddr())
		return nil
	})
	// loop write
	mc.wqueue = mux.NewShardQueue(mux.ShardSize, conn)
	return mc
}

type CliConn struct {
	conn   netpoll.Connection
	rch    chan netpoll.Reader
	wqueue *mux.ShardQueue // use for write

}

// Encode .
func Encode(writer netpoll.Writer, msg *rpc.RpcRequestMessage) (err error) {
	buf, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal request: %v\n", err)
		return err
	}
	lenBuf, _ := writer.Malloc(4)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(buf)))
	writer.WriteBinary(buf)
	err = writer.Flush()
	return err
}

// Decode .
func Decode(reader netpoll.Reader, msg *rpc.RpcResponseMessage) (err error) {
	bLen, err := reader.Next(4)
	if err != nil {
		return err
	}
	l := int(binary.LittleEndian.Uint32(bLen))

	buf, err := reader.ReadBinary(l)
	if err != nil {
		return err
	}
	if err = proto.Unmarshal(buf, msg); err != nil {
		return err
	}
	err = reader.Release()
	return err
}
