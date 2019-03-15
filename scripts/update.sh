#/usr/bin/bash

update="cd dkg-go && git pull "
pemPath="${HOME}/.ssh/ax.pem"
fileName="server.json"

declare -i count=0

# awk -v p=$perPath update=$update '
# {
# 	cmd="ssh -o StrictHostKeyChecking=no -i ${pemPath} ${server} \"${update}\""
# }
# ' server.json

while read server;
do 	
	let count+=1
	if (( count < 25 && count > 15 ))
		then
			echo $count
	cmd="ssh -o StrictHostKeyChecking=no -i ${pemPath} ${server} \"${update}\""
	eval $cmd &
		fi
done < $fileName

wait 






