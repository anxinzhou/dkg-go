#/usr/bin/bash


run(){
	host=$(/sbin/ifconfig -a|grep inet|grep -v 127.0.0.1|grep -v inet6|awk '{print $2}'|tr -d "addr:")
	port=4000
	echo "http://${host}:${port}"
	export GOPATH=$(dirname $(pwd))
	export CGO_ENABLED=0
	go run main.go -host=$host -port=$port -p=1
}


update="cd dkg-go && git pull"
declare -i count=0
cat server.json | while read server 
do 	
	cmd="ssh -i ${HOME}/.ssh/ax.pem ${server} \"${update}\""
	eval $cmd
	let count++
	echo $count
done 



# ssh -i "~/.ssh/ax.pem" ubuntu@ec2-3-89-90-133.compute-1.amazonaws.com "/bin/bash" < update.sh



# run="cd dkg/ ;  \\
# 		export GOPATH=$(dirname $(pwd)) \\
# 		cd src/ \\
# 		tc qdisc add dev eth0 root netem delay 100ms \\
# 		host=$(/sbin/ifconfig -a|grep inet|grep -v 127.0.0.1|grep -v inet6|awk '{print $2}'|tr -d \"addr:\") \\
# 		port=4000 \\
# 		echo \"http://${host}:${port}\" \\
# 		export CGO_ENABLED=0 \\
# 		go run main.go -host=$host -port=$port -p=1"



# #update

# for (( i=0; i<count; i++ )) 
# do 
# 	echo $update
# 	call=${sshs[${count}]}" \" ${update}\" " 
# 	call &
# done 







