#!/usr/bin/env bash

#/usr/bin/bash

cd ../
export GOPATH=$(pwd)
cd src

for (( num=4; num<=32; num+=4 ));
do
    go run crypto.go -num=${num} &> "log/log${num}"
done