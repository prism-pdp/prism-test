#!/bin/bash

set -e

. /opt/prism/contract-addr.env

PATH_CACHE=/opt/prism/cache
PATH_LOG=$PATH_CACHE/prism.log

HARNESS_OPTS="--server ws://testnet:8545 --contract $CONTRACT_ADDR --log $PATH_LOG --cache $PATH_CACHE"
HARNESS="/opt/prism/bin/harness $HARNESS_OPTS"

AIDE_OPTS="--log $PATH_LOG"
AIDE="/opt/prism/bin/aide $AIDE_OPTS"

SENDER_OPTS_SM="--sender-addr $ADDRESS_0 --sender-key $PRIVKEY_0"
SENDER_OPTS_SP="--sender-addr $ADDRESS_1 --sender-key $PRIVKEY_1"
SENDER_OPTS_TPA="--sender-addr $ADDRESS_2 --sender-key $PRIVKEY_2"
SENDER_OPTS_SU1="--sender-addr $ADDRESS_4 --sender-key $PRIVKEY_4"
SENDER_OPTS_SU2="--sender-addr $ADDRESS_5 --sender-key $PRIVKEY_5"

rm -rf $PATH_CACHE/*

$HARNESS $SENDER_OPTS_SM setup $ADDRESS_0 $PRIVKEY_0 $ADDRESS_1 $PRIVKEY_1 
$HARNESS $SENDER_OPTS_SM enroll auditor tpa1 $ADDRESS_2
$HARNESS $SENDER_OPTS_SM enroll user    su1  $ADDRESS_4 $PRIVKEY_4
$HARNESS $SENDER_OPTS_SM enroll user    su2  $ADDRESS_5 $PRIVKEY_5

# setup evaluation
BLOCK_NUM=100
FILE_SIZE=1M
TRIAL_COUNT=100
PATH_TESTDATA=/tmp/test.dat
for i in `seq $TRIAL_COUNT`; do
    $AIDE write-log "Upload test data (index:$i)"
    $AIDE testdata $PATH_TESTDATA $FILE_SIZE $i
    $HARNESS $SENDER_OPTS_SP upload su1 $PATH_TESTDATA $BLOCK_NUM
    $HARNESS $SENDER_OPTS_SP upload su2 $PATH_TESTDATA $BLOCK_NUM
done

$HARNESS $SENDER_OPTS_SU1 challenge su1 0.6 1.0
$HARNESS $SENDER_OPTS_SP proof su1
$HARNESS $SENDER_OPTS_TPA audit tpa1 su1

cp $PATH_LOG /opt/prism/logs/contract.log
