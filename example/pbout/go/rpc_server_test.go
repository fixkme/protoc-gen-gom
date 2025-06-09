package main

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/fixkme/gokit/rpc"
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
		Addr:          "127.0.0.1:2333",
		PollerNum:     4,
		ProcessorSize: 4,
	}
	server := rpc.NewServer(opt)
	gate.RegisterGateServer(server, &ServiceImp{})

	log.Printf("Server is listening on %s\n", opt.Addr)
	server.Run()
}

type ServiceImp struct {
}

func (s *ServiceImp) NoticePlayer(_ context.Context, req *gate.CNoticePlayer) (*gate.SNoticePlayer, error) {
	fmt.Printf("handler logic NoticePlayer:%v\n", req)
	return &gate.SNoticePlayer{Content: fmt.Sprintf("echoxxx_%d", req.PlayerId)}, nil
	//return nil, fmt.Errorf("handler NoticePlayer logic error")
}
