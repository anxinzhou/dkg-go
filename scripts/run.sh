#/usr/bin/bash

trap 'killAll' SIGINT

killAll(){
pkill -P $$
}

num=32

runShell="excecute.sh"

declare -i count=0
while read server 
do 	
	let count++
	cmd="ssh -o StrictHostKeyChecking=no ${server} \"bash -s ${count} ${num} \" <  ${runShell}"
	eval $cmd &
	pids[${count}]=$!
	if (( count==num ))
	then 
		break;
	fi
done < server.json

for pid in ${pids[*]}; do
	wait $pid
done

