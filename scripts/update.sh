#/usr/bin/bash

trap 'killAll' SIGINT

killAll(){
pkill -P $$
}


# update="cd dkg-go && git pull "
update="echo \"test\""
pemPath="${HOME}/.ssh/ax.pem"
fileName="server.json"

declare -i count=0

content=$( cat $fileName)
for server in $content 
do 
	let count+=1
	# if ((count<16))
	# then
	cmd='ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i ${pemPath} ${server} "${update}"'
	eval $cmd &
	# fi
done

wait 






