#### 后端协议生成插件，生成pb、model代码

##### 1、插件参数
- go-mod: go mod包名，见go.mod文件的module声明
- rpc-pkgs: rpc包名列表，用^隔开；这些包里面定义了request和response，还有push消息
- data-pkgs: data包名列表，用^隔开；这些包里面定义data，插件会给message加上PB前缀，目的是和M前缀的model区分开

##### 2、使用例子
- 安装插件
```shell
go install .
```
- 拷贝扩展internal/pbext/options_ext.proto到 example/proto/pbext 目录
- 生成代码
``` shell
protoc -I ./example/proto -I ./example/proto_ss  --gom_out=./example/pbout/go \
--gom_opt=paths=source_relative,\
go-mod=github.com/fixkme/protoc-gen-gom/example/pbout/go,\
data-pkgs=datas^model,\
rpc-pkgs=game^gate \
./example/proto/datas/*.proto ./example/proto/model/*.proto ./example/proto/game/*.proto \
./example/proto_ss/gate/*.proto ./example/proto_ss/datas/*.proto ./example/proto_ss/game/*.proto
```
- proto文件命名规范 

    proto文件夹是前后端公共协议，proto_ss是服务端内部协议；
    proto和proto_ss下的目录名字就是golang的包名，如果两边的内容需要相同包名，那么目录名字应该一样；
    为了能够通过protoc编译、区分cs、ss，不能出现重名的proto文件，并且，proto_ss里的文件名需要带上_ss之类的标识符