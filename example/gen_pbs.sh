#!/bin/sh
set -e
# 进入脚本所在目录
current_dir="$(dirname "$0")"
cd "$current_dir"

# 生成目录
OUT_DIR="./pbout/go"
PROTO_DIR="./proto"
PROTO_SS_DIR="./proto_ss"

# 清理旧文件
rm -rf $OUT_DIR/*

protoc -I $PROTO_DIR -I $PROTO_SS_DIR --gom_out=$OUT_DIR \
--gom_opt=paths=source_relative,\
debug=true,\
syncKeyNoCache=false,\
data-pkgs=datas^model,\
rpc-pkgs=game^gate \
./proto/pbext/*.proto \
./proto/cerrs/*.proto \
./proto/datas/*.proto ./proto/model/*.proto ./proto/game/*.proto \
./proto_ss/gate/*.proto ./proto_ss/datas/*.proto  ./proto_ss/game/*.proto
