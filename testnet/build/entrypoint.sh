#!/bin/sh

if [ "$1" = "" ]; then
    /bin/ash
elif [ "$1" = "deploy" ]; then
    contract="$2"
	forge create \
        --mnemonic-path "$WALLET_MNEMONIC" \
        src/${contract}.sol:${contract}
elif [ "$1" = "build" ]; then
    forge build
    for f in $(ls src/*.sol); do
        name=$(basename $f .sol)
        jq -c '.abi' ./out/${name}.sol/${name}.json > ./cache/${name}.abi
        jq -c -r '.bytecode.object' ./out/${name}.sol/${name}.json > ./cache/${name}.bin
        abigen --abi ./cache/${name}.abi --bin ./cache/${name}.bin --pkg sol --type ${name} --out ./cache/${name}.go
    done
elif [ "$1" = "derive-private-key" ]; then
    cast wallet derive-private-key \
        "$WALLET_MNEMONIC" \
        0
elif [ "$1" = "start" ]; then
    anvil \
        --host $RPC_HOST \
        --port $RPC_PORT \
        --accounts $NUM_FIRST_ACCOUNTS \
        --balance $BALANCE_FIRST_ACCOUNTS \
        --mnemonic "$WALLET_MNEMONIC"
else
    exec "$@"
fi