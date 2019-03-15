ins="sudo sed  '$ a PermitUserEnvironment yes' /etc/ssh/sshd_config && sudo /etc/init.d/ssh restart"


declare -i count=0
cat server.json | while read server 
do 	
	cmd="ssh -i ${HOME}/.ssh/ax.pem ${server} \"${ins}\" "
	eval $cmd &
	let count++
	echo $count
done

wait 