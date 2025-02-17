##### 后端协议生成插件，生成pb、model代码

###### 1、插件参数
- go-mod: go mod包名
- rpc-pkgs: rpc包名列表，用#隔开

###### 2、使用例子
``` shell
# 安装插件
go install .

# 生成代码
protoc -I ./example/proto  --gom_out=paths=source_relative,\
go-mod=github.com/protoc-gen-gom/example/pbout/go,\
rpc-pkgs=game:\
./example/pbout/go \
./example/proto/model/*.proto ./example/proto/datas/*.proto ./example/proto/game/*.proto
```