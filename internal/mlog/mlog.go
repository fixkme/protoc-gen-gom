package mlog

import (
	"fmt"
	"os"
)

var Debug bool

func Info(format string, args ...any) {
	if !Debug {
		return
	}
	// stdout 已被protoc的插件使用
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
