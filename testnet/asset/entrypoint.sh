#!/bin/sh

function get_addr()
{
    index="$1"
    cast wallet derive-private-key "$WALLET_MNEMONIC" $index | sed -n 2p | cut -d ':' -f 2 | tr -d ' ' | cut -c 3-
}

function get_key()
{
    index="$1"
    cast wallet derive-private-key "$WALLET_MNEMONIC" $index | sed -n 3p | cut -d ':' -f 2 | tr -d ' ' | cut -c 3-
}

if [ "$1" = "" ]; then
	/bin/ash
elif [ "$1" = "deploy" ]; then
	contract="$2"
	sm_key="$3"
	sp_addr="$4"
	forge create \
		--private-key ${sm_key} \
		src/${contract}.sol:${contract} \
		--constructor-args ${sp_addr} > ./cache/deploy.log
	cat ./cache/deploy.log | grep 'Deployed to:' | cut -d ':' -f 2 | tr -d ' ' | cut -c 3-
elif [ "$1" = "show-accounts" ]; then
    for i in $(seq $NUM_FIRST_ACCOUNTS)
    do
        num=$((i-1))
        address=$(get_addr $num)
        privkey=$(get_key  $num)
        echo "ADDRESS_$num=$address"
        echo "PRIVKEY_$num=$privkey"
    done
elif [ "$1" = "build" ]; then
    forge build
    for f in $(ls src/*.sol); do
        name=$(basename $f .sol)
        jq -c '.abi' ./out/${name}.sol/${name}.json > ./cache/${name}.abi
        jq -c -r '.bytecode.object' ./out/${name}.sol/${name}.json > ./cache/${name}.bin
        abigen --abi ./cache/${name}.abi --bin ./cache/${name}.bin --pkg sol --type ${name} --out ./cache/${name}.go
    done
elif [ "$1" = "start" ]; then
    anvil \
        --host $RPC_HOST \
        --port $RPC_PORT \
        --accounts $NUM_FIRST_ACCOUNTS \
        --balance $BALANCE_FIRST_ACCOUNTS \
        --mnemonic "$WALLET_MNEMONIC" \
        --block-time 5
else
    exec "$@"
fi
