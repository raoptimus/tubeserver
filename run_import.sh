#!/bin/sh
export GOPATH=$(pwd);
echo "GOPATH="$GOPATH;
rm bin/importer
go build -o bin/importer "./src/ts/importer" 
chmod 0755 bin/importer 
bin/importer
