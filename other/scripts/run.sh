#!/usr/bin/env bash
lambda=$1
duration=$2
pssh -i -t 1000 -x "-o StrictHostKeyChecking=no" -h ${HOME}/dkg-go/scripts/server.json "\
cd dkg-go/other/ && \
export GOPATH=$(pwd) &&\
cd src/ && \
/snap/bin/go run hash.go -lambda=${lambda} -duration=${duration}"
