package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"os"
	"path/filepath"

	"github.com/protoc-gen-gom/internal/gen"
	"github.com/protoc-gen-gom/internal/mlog"
	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
)

const grpcDocURL = "https://grpc.io/docs/languages/go/quickstart/#regenerate-grpc-code"
const help = `go-mod=github.com/protoc-gen-gom/example/pbout/go`

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
		goMod    = flags.String("go-mod", "", "go module name")
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
		if *goMod == "" {
			return errors.New("protoc-gen-gom: go-mod is required")
		}
		mlog.Info("go-mod=%v", *goMod)
		mlog.Info("dataPkgs=%s", *dataPkgs)
		mlog.Info("rpcPkgs=%s", *rpcPkgs)

		dataFiles, rpcFiles := getParams(*dataPkgs, *rpcPkgs)
		// dataPkgs是 message名字必须加上PB前缀的包名
		for _, f := range pg.Files {
			pkgName := string(f.GoPackageName)
			if _, ok := dataFiles[pkgName]; !ok {
				continue
			} else if _, ok := rpcFiles[pkgName]; ok {
				continue
			}
			for _, m := range f.Messages {
				m.GoIdent.GoName = "PB" + m.GoIdent.GoName
			}
		}

		for _, f := range pg.Files {
			if f.Generate {
				gengo.GenerateFile(pg, f)
			}
		}

		gen.InitGoModuleName(*goMod)
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
