#!/bin/sh
export GOPATH=$(pwd);
echo "GOPATH="$GOPATH;

echo "rebuild server"
echo ""
rm bin/server
go build -o bin/server "./src/ts/server" 
chmod 0755 bin/server

echo "rebuild importer"
echo ""
rm bin/importer
go build -o bin/importer "./src/ts/importer" 
chmod 0755 bin/importer

echo "rebuild publisher"
echo ""
rm bin/publisher
go build -o bin/publisher "./src/ts/publisher"
chmod 0755 bin/publisher

echo "rebuild push"
echo ""
rm bin/push
go build -o bin/push "./src/ts/push"
chmod 0755 bin/push

echo "rebuild appmv"
echo ""
rm bin/appmv
go build -o bin/appmv "./src/ts/appmv"
chmod 0755 bin/appmv

echo "rebuild translator"
echo ""
rm bin/translator
go build -o bin/translator "./src/ts/translator"
chmod 0755 bin/translator

echo "rebuild finish"
echo ""
