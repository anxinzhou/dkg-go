#/usr/bin/bash

trap 'killAll' SIGINT

killAll(){
pkill -P $$
}

num=$1

runShell="excecute.sh"
startTime=$(date)
echo $startTime

declare -i count=0
while read server 
do 	
	let count++
	cmd="ssh -o StrictHostKeyChecking=no ${server} \"bash -s ${count} ${num} '${startTime}' \" <  ${runShell}"
	eval $cmd &
	pids[${count}]=$!
	if (( count==num ))
	then 
		break;
	fi
done < server.json

#for pid in ${pids[*]}; do
#	wait $pid
#done

sleep 15

for pid in ${pids[*]}; do
	kill -9 $pid
done

./kill.sh
