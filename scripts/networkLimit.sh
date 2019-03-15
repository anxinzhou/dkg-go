limit="sudo tc qdisc add dev eth0 root netem delay 100ms"


declare -i count=0
cat server.json | while read server 
do 	
	# let count++
	# echo $count
	cmd="ssh -o StrictHostKeyChecking=no -i ${HOME}/.ssh/ax.pem ${server} \"${limit}\""
	eval $cmd &
done 

wait 