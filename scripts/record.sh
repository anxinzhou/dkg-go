#/usr/bin/bash

for num in 4 8 16 24 32
do
	for order in 1 2 3 4
	do
		echo "recording ${num} ${order} now"
		./run.sh ${num} &>"log/log${num}_${order}"
	done
done
