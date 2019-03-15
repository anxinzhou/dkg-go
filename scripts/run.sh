#/usr/bin/bash


runShell="excecute.sh"

declare -i count=0
cat server.json | while read server 
do 	
	let count++
	echo $count
	cmd="ssh -i ${HOME}/.ssh/ax.pem ${server} \"bash -s ${count} \" <  ${runShell}"
	eval $cmd &
done

wait
