#!/bin/sh
export GOPATH=$(pwd);
echo "GOPATH="$GOPATH;
rm bin/push
go build -o bin/push "./src/ts/push"
chmod 0755 bin/push
bin/push
