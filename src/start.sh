#!/usr/bin/env bash
host=$(/sbin/ifconfig -a|grep inet|grep -v 127.0.0.1|grep -v inet6|awk '{print $2}'|tr -d "addr:")
port=4000
echo "http://${host}:${port}"
export GOPATH=$(dirname $(pwd))
export CGO_ENABLED=0
go run main.go -host=$host -port=$port -p=1
