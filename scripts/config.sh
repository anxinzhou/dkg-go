#!/usr/bin/env bash
pssh -i -x "-o StrictHostKeyChecking=no" -h server.json "if [ ! -d "dkg-go" ]; then git clone -b rpc https://github.com/xxRanger/dkg-go.git; fi\
&& sudo snap install go --classic"
