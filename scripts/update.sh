#/usr/bin/bash

update="cd dkg-go && git pull && tc qdisc add dev eth0 root netem delay 100ms"
# declare -i count=0
cat server.json | while read server 
do 	
	cmd="ssh -o StrictHostKeyChecking=no -i ${HOME}/.ssh/ax.pem ${server} \"${update}\""
	eval $cmd &
	# let count++
	# echo $count
done 

wait 






