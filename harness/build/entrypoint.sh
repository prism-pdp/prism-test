#!/bin/bash

if [ "$1" = "" ]; then
	/bin/bash
elif [ "$1" = "build" ]; then
	go build -o bin/harness
elif [ "$1" = "make-conf" ]; then
	jq -n \
		--arg server $2 \
		--arg privKey $3 \
		--arg contractAddr $4 \
		-f /etc/dpduado/harness/config.json.template > ./cache/config.json
else
	exec "$@"
fi
