#/usr/bin/bash

trap 'killAll' SIGINT

killAll(){
pkill -P $$
}


runShell="excecute.sh"

declare -i count=0
cat server.json | while read server 
do 	
	let count++
	cmd="ssh -o StrictHostKeyChecking=no ${server} \"bash -s ${count} \" <  ${runShell}"
	eval $cmd &
done

sleep 10
wait
