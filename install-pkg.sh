#!/bin/bash
export GOPATH=$(pwd)
echo "go path is set to  $GOPATH"

go get gopkg.in/mgo.v2
go get github.com/go-sql-driver/mysql
go get github.com/ziutek/mymysql/thrsafe
go get github.com/ziutek/mymysql/autorc
go get github.com/ziutek/mymysql/godrv
go get github.com/go-gorp/gorp

