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

	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/protoc-gen-gom/example/pbout/go/gate"
	"github.com/panjf2000/gnet/v2"
	"google.golang.org/protobuf/proto"
)

func TestClient(t *testing.T) {
	conn, err := rpc.NewConnection("127.0.0.1:2333")
	if err != nil {
		log.Fatalf("new connection error: %v", err)
	}
	logicFn := func(sync bool) {
		for i := 0; i < 5; i++ {
			rsp := &gate.SNoticePlayer{}
			if err := conn.Invoke(context.Background(), "Gate/NoticePlayer", &gate.CNoticePlayer{PlayerId: int64(i)}, rsp, sync); err != nil {
				log.Fatalf("invoke error: %v", err)
			}
			fmt.Printf("call rsp:%v\n", rsp)
			time.Sleep(time.Microsecond * time.Duration(rand.Intn(1000)))
			// rsp = &gate.SNoticePlayer{}
			// if err := conn.Invoke(context.Background(), "Gate/NoticePlayerr", &gate.CNoticePlayer{PlayerId: 2}, rsp, true); err != nil {
			// 	log.Fatalf("invoke error: %v", err)
			// }
			// fmt.Printf("call rsp:%v\n", rsp)
			// return
		}
	}
	go logicFn(false)
	go logicFn(false)
	go logicFn(false)

	select {
	case <-time.After(20 * time.Second):
	}
}

func testf() {
	h := &rpc.ClientHander{}
	cs, err := gnet.NewClient(h, gnet.WithMulticore(false))
	if err != nil {
		log.Fatalf("new client error: %v", err)
	}
	if err = cs.Start(); err != nil {
		log.Fatalf("start client error: %v", err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	c, err := cs.DialContext("tcp", "127.0.0.1:2333", ctx)
	if err != nil {
		log.Fatalf("dial error: %v", err)
	}

	go func() {
		for {
			call(c, "Gate/NoticePlayer", &gate.CNoticePlayer{})
			//c.Flush()
			call(c, "Gate/NoticePlayerb", &gate.CNoticePlayer{})
			//return
			time.Sleep(5 * time.Second)
		}
	}()
	select {
	case <-time.After(20 * time.Second):
	}

}

func call(c gnet.Conn, path string, data proto.Message) {
	// 构造请求消息
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
	// 序列化请求消息
	buf, err := proto.Marshal(rpcReq)
	if err != nil {
		log.Printf("Failed to marshal request: %v\n", err)
		return
	}
	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(buf)))
	if _, err = c.Write(lenBuf); err != nil {
		log.Printf("Failed to Write lenBuf: %v\n", err)
		return
	}
	if _, err = c.Write(buf); err != nil {
		log.Printf("Failed to Write buf: %v\n", err)
		return
	}
	fmt.Println("call succeed")
}
