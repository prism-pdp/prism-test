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

for block_num in `seq 100 100 1000`; do
    rm -rf $PATH_CACHE/*

    $AIDE write-log "Start eval auditing (block_num:${block_num})"

    $HARNESS setup $ADDRESS_0 $PRIVKEY_0 $ADDRESS_1 $PRIVKEY_1 
    $HARNESS enroll auditor tpa1 $ADDRESS_2
    $HARNESS enroll user    su1  $ADDRESS_4 $PRIVKEY_4

    $AIDE testdata $PATH_TESTDATA 1000M 255
    $HARNESS upload su1 $PATH_TESTDATA $block_num

    # Fixed block num
    ratio=$(perl -e "print 100/${block_num}")
    $AIDE write-log "Start fixed block num (ratio:${ratio})"
    for i in `seq $TRIAL_COUNT`; do
        $HARNESS challenge su1 $ratio 1.0
        $HARNESS proof su1
        $HARNESS audit tpa1 su1
    done
    $AIDE write-log "Finish fixed block num (ratio:${ratio})"

    # Fixed block ratio
    ratio="0.1"
    $AIDE write-log "Start fixed block ratio (ratio:${ratio})"
    for i in `seq $TRIAL_COUNT`; do
        $HARNESS challenge su1 $ratio 1.0
        $HARNESS proof su1
        $HARNESS audit tpa1 su1
    done
    $AIDE write-log "Finish fixed block ratio (ratio:${ratio})"

    $AIDE write-log "Finish eval auditing (block_num:${block_num})"

    mv $PATH_LOG /opt/dpduado/logs/auditing-${block_num}.log
done



# auditing
for ratio in `seq 0.1 0.1 1.0`; do
    rm -rf $PATH_CACHE/*

    $HARNESS setup $ADDRESS_0 $PRIVKEY_0 $ADDRESS_1 $PRIVKEY_1 
    $HARNESS enroll auditor tpa1 $ADDRESS_2
    $HARNESS enroll user    su1  $ADDRESS_4 $PRIVKEY_4

    $AIDE testdata $PATH_TESTDATA 1000M 255
    $HARNESS upload su1 $PATH_TESTDATA 1000

    for i in `seq $TRIAL_COUNT`; do
        $HARNESS challenge su1 $ratio 1.0
        $HARNESS proof su1
        $HARNESS audit tpa1 su1
    done

    mv $PATH_LOG /opt/dpduado/logs/auditing-${ratio}.log
done
