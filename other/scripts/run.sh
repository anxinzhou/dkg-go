#!/usr/bin/env bash
lambda=$1
duration=$2
pssh -i -x "-o StrictHostKeyChecking=no" -h server.json "\
cd ../ \
export GOPATH=$(PWD)\
go run hash.go -lambda=${lambda} -duration=${duration}"


