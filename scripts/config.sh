#!/usr/bin/env bash
pssh -i -x "-o StrictHostKeyChecking=no" -h server.json "git clone https://github.com/xxRanger/dkg-go.git && sudo snap install go --classic"