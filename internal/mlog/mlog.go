package mlog

import (
	"fmt"
	"os"
)

func Info(format string, args ...any) {
	// stdout 已被protoc的插件使用
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
