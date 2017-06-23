#!/bin/sh
export GOPATH=$(pwd);
echo "GOPATH="$GOPATH;
rm bin/appmv
go build -o bin/appmv "./src/ts/appmv"
chmod 0755 bin/appmv
DB_MAIN_URL=..com/tubeserver bin/appmv
