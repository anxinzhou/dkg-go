#/usr/bin/bash


runShell="excecute.sh"

declare -i count=0
cat server.json | while read server 
do 	
	cmd="ssh -i ${HOME}/.ssh/ax.pem ${server} \"bash -s\" <  ${runShell}"
	eval $cmd
	let count++
	echo $count
done
