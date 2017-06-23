#!/bin/sh
echo "rebuild server"
echo ""
rm bin/server
go build -o bin/server "./src/ts/server" 
chmod 0755 bin/server

./bin/server -config=./config/ra.json -pid=./bin/server_ra.pid