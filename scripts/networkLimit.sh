pssh -i -x "-o StrictHostKeyChecking=no" -h server.json "sudo tc qdisc add dev eth0 root netem delay 100ms"


