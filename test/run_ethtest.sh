#!/bin/bash

set -euo pipefail

HARNESS_HOST_PATH="./harness/volume"
HARNESS_CONTAINER_PATH="/var/lib/prism-harness"

CONTRACT="XZ21"

source ../cache/accounts.env
source ../cache/contract-addr.env

ETHERNET_OPTS="--server ws://testnet:8545 --contract ${CONTRACT_ADDR}"
ETHERNET_SENDER_OPTS_0="--sender-addr ${ADDRESS_0} --sender-key ${PRIVKEY_0}"
ETHERNET_SENDER_OPTS_1="--sender-addr ${ADDRESS_1} --sender-key ${PRIVKEY_1}"
ETHERNET_SENDER_OPTS_2="--sender-addr ${ADDRESS_2} --sender-key ${PRIVKEY_2}"
ETHERNET_SENDER_OPTS_3="--sender-addr ${ADDRESS_3} --sender-key ${PRIVKEY_3}"
ETHERNET_SENDER_OPTS_4="--sender-addr ${ADDRESS_4} --sender-key ${PRIVKEY_4}"
ETHERNET_SENDER_OPTS_5="--sender-addr ${ADDRESS_5} --sender-key ${PRIVKEY_5}"

(
    cd ..

	make harness@run CMD="clean"
    make harness@run CMD="fallocate -l 100M /var/lib/prism-harness/cache/dummy.data"

	make harness@run CMD="harness ${ETHERNET_OPTS} ${ETHERNET_SENDER_OPTS_0} setup ${ADDRESS_0} ${PRIVKEY_0} ${ADDRESS_1} ${PRIVKEY_1}"
	make harness@run CMD="harness ${ETHERNET_OPTS} ${ETHERNET_SENDER_OPTS_0} enroll auditor tpa1 ${ADDRESS_2}"
	make harness@run CMD="harness ${ETHERNET_OPTS} ${ETHERNET_SENDER_OPTS_0} enroll auditor tpa2 ${ADDRESS_3}"
	make harness@run CMD="harness ${ETHERNET_OPTS} ${ETHERNET_SENDER_OPTS_0} enroll user    su1  ${ADDRESS_4} ${PRIVKEY_4}"
	make harness@run CMD="harness ${ETHERNET_OPTS} ${ETHERNET_SENDER_OPTS_0} enroll user    su2  ${ADDRESS_5} ${PRIVKEY_5}"
	make harness@run CMD="harness ${ETHERNET_OPTS} ${ETHERNET_SENDER_OPTS_1} upload su1 ${HARNESS_CONTAINER_PATH}/cache/dummy.data 100"
	make harness@run CMD="harness ${ETHERNET_OPTS} ${ETHERNET_SENDER_OPTS_1} upload su2 ${HARNESS_CONTAINER_PATH}/cache/dummy.data 50"
	make harness@run CMD="harness ${ETHERNET_OPTS} ${ETHERNET_SENDER_OPTS_4} challenge su1 0.55 1.0"
	make harness@run CMD="harness ${ETHERNET_OPTS} ${ETHERNET_SENDER_OPTS_1} proof su1"
	make harness@run CMD="harness ${ETHERNET_OPTS} ${ETHERNET_SENDER_OPTS_2} audit tpa1 su1"
)