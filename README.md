#### 游戏后端protobuf协议生成go code插件

##### 1、介绍
- 通过protobuf生成pb、model的golang代码，用于[rpc服务](https://github.com/fixkme/gokit/tree/main/rpc)和model数据变化[delta收集](https://github.com/fixkme/gokit/tree/main/db/mongo/delta)
- 根据 enum+@CodeError标签 生成[CodeError](https://github.com/fixkme/gokit/blob/main/errs/error.go)代码
- 根据https://github.com/protocolbuffers/protobuf-go/tree/master 改造而来

##### 2、插件参数
- debug: 是否打印debug信息
- syncKeyNoCache: 每个Model struct是否生成fieldSyncIDs []string字段，默认生成，如果不需要则设置为true。区别：有fieldSyncIDs则可以直接获得syncKey，但是耗费更多内存；反之，syncKey通过拼接字符串获得，省内存，但是频繁申请内存、GC释放，不利于CPU性能
- rpc-pkgs: rpc包名列表，用^隔开；这些包里面定义了request和response，还有push消息
- data-pkgs: data包名列表，用^隔开；这些包里面定义data，插件会给message加上PB前缀，目的是和M前缀的model区分开

##### 3、标签
- @CodeError: 生成CodeError，效果和[option (is_code_error) = true;]一样，见例子(https://github.com/fixkme/protoc-gen-gom/tree/main/example/proto/cerrs/cerrors.proto)
- @model: 生成model代码，效果和[option (is_model) = true;]一样，见例子(https://github.com/fixkme/protoc-gen-gom/tree/main/example/proto/model/player_model.proto)
> [option]和[注释标签]的效果是一样的，目的都是识别类型用于生成代码；区别是option会生成Extension代码，可能会对前端生成pb代码造成影响

##### 4、使用例子
- 安装插件
```shell
go install .
# 或者
go install github.com/fixkme/protoc-gen-gom@latest
```
- 拷贝扩展internal/pbext/options_ext.proto到 example/proto/pbext 目录，添加option go_package = "github.com/fixkme/protoc-gen-gom/example/pbout/go/pbext";
- 生成代码
``` shell
bash example/gen_pbs.sh
```
- proto文件命名规范 

    proto文件夹是前后端公共协议，proto_ss是服务端内部协议；
    proto和proto_ss下的目录名字就是golang的包名，如果两边的内容需要相同包名，那么目录名字应该一样；
    为了能够通过protoc编译、区分cs、ss，不能出现重名的proto文件，并且，proto_ss里的文件名需要带上_ss之类的标识符