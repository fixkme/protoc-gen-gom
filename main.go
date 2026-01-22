package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"os"
	"path/filepath"

	"github.com/fixkme/protoc-gen-gom/internal/gen"
	"github.com/fixkme/protoc-gen-gom/internal/mlog"
	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
)

const grpcDocURL = "https://grpc.io/docs/languages/go/quickstart/#regenerate-grpc-code"
const help = `https://github.com/fixkme/protoc-gen-gom/blob/main/README.md`

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Fprintf(os.Stdout, "%v %v\n", filepath.Base(os.Args[0]), gen.Version)
		os.Exit(0)
	}
	if len(os.Args) == 2 && os.Args[1] == "--help" {
		fmt.Fprintf(os.Stdout, help+"\n")
		os.Exit(0)
	}

	var (
		flags    flag.FlagSet
		dataPkgs = flags.String("data-pkgs", "", "data pkg names sep=^")
		rpcPkgs  = flags.String("rpc-pkgs", "", "rpc pkg names sep=^")
		plugins  = flags.String("plugins", "", "deprecated option")
	)
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(pg *protogen.Plugin) error {
		if *plugins != "" {
			return errors.New("protoc-gen-go: plugins are not supported; use 'protoc --go-grpc_out=...' to generate gRPC\n\n" + "See " + grpcDocURL + " for more information")
		}
		mlog.Info("data-pkgs=%s", *dataPkgs)
		mlog.Info("rpc-pkgs=%s", *rpcPkgs)

		dataFiles, rpcFiles := getParams(*dataPkgs, *rpcPkgs)
		// dataPkgs是 message名字必须加上PB前缀的包名
		for _, f := range pg.Files {
			pkgName := string(f.GoPackageName)
			if _, ok := dataFiles[pkgName]; !ok {
				continue
			}
			for _, m := range f.Messages {
				m.GoIdent.GoName = "PB" + m.GoIdent.GoName
			}
		}

		for _, f := range pg.Files {
			if f.Generate {
				// 生成pb文件
				gengo.GenerateFile(pg, f)
			}
		}

		// 生成rpc文件
		rpcGene := gen.NewRpcGenerator(pg)
		if err := rpcGene.GenerateFile(rpcFiles); err != nil {
			return err
		}

		// 生成model文件
		modelGene := gen.NewModelGenerator(pg)
		if err := modelGene.GenerateFile(); err != nil {
			return err
		}

		pg.SupportedFeatures = gengo.SupportedFeatures
		pg.SupportedEditionsMinimum = gengo.SupportedEditionsMinimum
		pg.SupportedEditionsMaximum = gengo.SupportedEditionsMaximum
		return nil
	})
}

func getParams(dataPkgs, rpcPkgs string) (datas, rpcs map[string]struct{}) {
	datas = map[string]struct{}{}
	for _, s := range strings.Split(dataPkgs, "^") {
		if len(s) > 0 {
			datas[s] = struct{}{}
		}
	}
	rpcs = map[string]struct{}{}
	for _, s := range strings.Split(rpcPkgs, "^") {
		if len(s) > 0 {
			rpcs[s] = struct{}{}
		}
	}
	return
}
