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
		-f /etc/prism/harness/config.json.template > ./cache/config.json
elif [ "$1" = "clean" ]; then
	rm -rf /var/lib/prism-harness/*
	mkdir /var/lib/prism-harness/cache
else
	exec "$@"
fi
