package main

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/protoc-gen-gom/example/pbout/go/gate"
)

func TestServer(t *testing.T) {
	opt := &rpc.ServerOpt{
		Addr:          "tcp4://127.0.0.1:2333",
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
	fmt.Printf("handler NoticePlayer:%v", req)
	return &gate.SNoticePlayer{Content: "echoxxx"}, nil
	//return nil, fmt.Errorf("handler NoticePlayer logic error")
}
