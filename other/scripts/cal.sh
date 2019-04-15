#!/usr/bin/env bash

#/usr/bin/bash

cd ../
export GOPATH=$(pwd)
cd src

for (( lambda=229; lambda>=225; lambda-=2 ))
do
    for (( t=60; t<=300; t+=60 ))
    do
        ./run.sh ${lambda} ${duration} &> "log/log${lambda}_${t}"
        echo "finish ${lambda}${t}"
    done
done