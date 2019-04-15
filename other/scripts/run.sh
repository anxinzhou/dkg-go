#!/usr/bin/env bash
lambda=$1
duration=$2
pssh -i -x "-o StrictHostKeyChecking=no" -h ${HOME}/dkg-go/scripts/server.json "\
cd ../ \
export GOPATH=$(pwd)\
go run hash.go -lambda=${lambda} -duration=${duration}"