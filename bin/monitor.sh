#!/bin/bash

PORT=$1
NAME=$3

while /bin/true; do
	echo Tailing...
	sudo ~/tools/bin/flasher --port $PORT --tail --tail-inactivity 10 >> $NAME.log
	while [ -f /tmp/flashing-$NAME ]; do
		echo Waiting...
		sleep 1
	done
done
