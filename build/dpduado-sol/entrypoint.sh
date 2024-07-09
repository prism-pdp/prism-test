#!/bin/sh

if [ "$1" = "" ]; then
    /bin/ash
elif [ "$1" = "build" ]; then
    forge build
    for f in $(ls src/*.sol); do
        name=$(basename $f .sol)
        jq -c '.abi' ./out/${name}.sol/${name}.json > ./cache/${name}.abi
        jq -c -r '.bytecode.object' ./out/${name}.sol/${name}.json > ./cache/${name}.bin
        abigen --abi ./cache/${name}.abi --bin ./cache/${name}.bin --pkg sol --type ${name} --out ./go-bindings/${name}.go
    done

else
    exec "$@"
fi