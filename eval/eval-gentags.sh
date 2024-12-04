#!/bin/bash

set -e

PATH_CACHE=/opt/dpduado/cache
PATH_LOG=$PATH_CACHE/dpduado.log
PATH_TESTDATA=/tmp/test.dat
TRIAL_COUNT=100

HARNESS_OPTS="--sim --log $PATH_LOG --cache $PATH_CACHE"
HARNESS="/opt/dpduado/bin/harness $HARNESS_OPTS"

AIDE_OPTS="--log $PATH_LOG"
AIDE="/opt/dpduado/bin/aide $AIDE_OPTS"

rm -rf $PATH_CACHE/*

$HARNESS setup $ADDRESS_0 $PRIVKEY_0 $ADDRESS_1 $PRIVKEY_1 
$HARNESS enroll user su1  $ADDRESS_4 $PRIVKEY_4

# gentags
for block_num in `seq 100 100 1000`; do
    for i in `seq $TRIAL_COUNT`; do
        $AIDE write-log "Upload test data (index:$i)"
        $AIDE testdata $PATH_TESTDATA ${block_num}M $i
        $HARNESS upload su1 $PATH_TESTDATA $block_num
        rm -f $PATH_CACHE/sp/*.dat
        rm -f $PATH_CACHE/sp/*.dat.tag
    done
    mv $PATH_LOG /opt/dpduado/logs/gentags-${block_num}.log
done