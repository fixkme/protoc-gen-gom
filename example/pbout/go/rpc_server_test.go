package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/cloudwego/netpoll"
	"github.com/fixkme/gokit/rpc"
	g "github.com/fixkme/gokit/util/go"
	"github.com/fixkme/protoc-gen-gom/example/pbout/go/gate"
)

func TestGnetServer(t *testing.T) {
	opt := &rpc.ServerOpt_Gnet{
		Addr:          "tcp4://127.0.0.1:2333",
		ProcessorSize: 4,
	}
	server := rpc.NewServer_Gnet(opt)
	gate.RegisterGateServer(server, &ServiceImp{})

	log.Printf("Server is listening on %s\n", opt.Addr)
	server.Run()
}

func TestNetpollServer(t *testing.T) {
	opt := &rpc.ServerOpt{
		ListenAddr:     "127.0.0.1:2333",
		PollerNum:      4,
		ProcessorSize:  4,
		DispatcherFunc: func(c netpoll.Connection, req *rpc.RpcRequestMessage) int { return int(req.Seq) },
	}
	server, err := rpc.NewServer(opt, context.Background())
	if err != nil {
		log.Fatalf("NewServer err:%v", err)
	}
	gate.RegisterGateServer(server, &ServiceImp{})

	log.Printf("Server is listening on %s\n", opt.ListenAddr)
	server.Run()
}

type ServiceImp struct {
}

func (s *ServiceImp) NoticePlayer(_ context.Context, req *gate.CNoticePlayer) (*gate.SNoticePlayer, error) {
	if rand.Float32() > 0.5 {
		fmt.Printf("%d is waiting\n", req.PlayerId)
		time.Sleep(time.Second * 2)
	}
	fmt.Printf("GoroutineID %d handler logic NoticePlayer:%v\n", g.GoroutineID(), req)
	return &gate.SNoticePlayer{Content: fmt.Sprintf("echoxxx_%d", req.PlayerId)}, nil
	//return nil, fmt.Errorf("handler NoticePlayer logic error")
}
