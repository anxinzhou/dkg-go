#/usr/bin/bash


index=$1

run(){
	host=$(/sbin/ifconfig -a|grep inet|grep -v 127.0.0.1|grep -v inet6|awk '{print $2}'|tr -d "addr:")
	port=4000
	# echo "http://${host}:${port}"
	cd dkg-go/
	cd src/
	export GOPATH=$(dirname $(pwd))
	export CGO_ENABLED=0
	echo "dsaindex"$index
	/snap/bin/go run main.go -host=$host -port=$port -p=1 -num=4 -index=${index}
}

run


