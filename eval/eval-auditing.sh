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
        $HARNESS proof
        $HARNESS audit tpa1
    done

    mv $PATH_LOG /opt/dpduado/logs/auditing-${ratio}.log
done
