#/usr/bin/bash

for num in 12 20 24 28
do
	for order in 1 2 3 4
	do
		echo "recording ${num} ${order} now"
		./run.sh ${num} &>"log/log${num}_${order}"
	done
done
