pssh -i -x "-o StrictHostKeyChecking=no" -h server.json " ps -x |grep main | grep -v grep | awk ' { print \$1 }' | while read pid; do kill -9 \$pid; done "
