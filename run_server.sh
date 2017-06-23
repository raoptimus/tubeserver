#!/bin/sh
export GOPATH=$(pwd);
echo "GOPATH="$GOPATH;
rm bin/server
go build -o bin/server "./src/ts/server" 
chmod 0755 bin/server 
./server1.sh
