#!/usr/bin/env bash

installGoPackage() {
	sudo apt install -y gcc
	echo "installing go package"
	export GOPATH=$(dirname "$PWD")
	go get -u github.com/gorilla/mux
	echo "finish installing go package"
}

installGO() {
	echo "installing go"
	if [ -x "$(command -v go)" ];
	then
		echo "have installed go"
		return
	fi

	if [ "$(uname)" = "Darwin" ];
	then
		brew update
		brew install golang
	elif [ "$(expr substr $(uname -s) 1 5)" = "Linux" ];
	then
		sudo snap install go --classic
	else
		echo "do not support this system"
		exit 1
	fi

	echo "finish installing go"
}

installGO
installGoPackage
