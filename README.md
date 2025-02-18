##### 后端协议生成插件，生成pb、model代码

###### 1、插件参数
- go-mod: go mod包名，见go.mod文件的module声明
- rpc-pkgs: rpc包名列表，用#隔开；这些包里面定义了request和response，还有push消息
- data-pkgs: data包名列表，用#隔开；这些包里面定义data，插件会给message加上PB前缀，目的是和M前缀的model区分开

###### 2、使用例子
- 安装插件
```shell
go install .
```
- 拷贝扩展internal/pbext/options_ext.proto到 example/proto/pbext 目录
- 生成代码
``` shell
protoc -I ./example/proto  --gom_out=paths=source_relative,\
go-mod=github.com/protoc-gen-gom/example/pbout/go,\
data-pkgs=datas^model,\
rpc-pkgs=game:\
./example/pbout/go \
./example/proto/datas/*.proto ./example/proto/model/*.proto ./example/proto/game/*.proto
``` 