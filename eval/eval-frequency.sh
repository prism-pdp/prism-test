#!/bin/bash

set -e

PATH_CACHE=/opt/prism/cache

FILE_RATIO=$1

HARNESS_OPTS="--sim --log $PATH_CACHE/prism.log --cache $PATH_CACHE"
HARNESS="/opt/prism/bin/harness $HARNESS_OPTS"

AIDE_OPTS="--log $PATH_CACHE/prism.log"
AIDE="/opt/prism/bin/aide $AIDE_OPTS"

rm -rf $PATH_CACHE/*

$HARNESS setup $ADDRESS_0 $PRIVKEY_0 $ADDRESS_1 $PRIVKEY_1 
$HARNESS enroll auditor tpa1 $ADDRESS_2
$HARNESS enroll user    su1  $ADDRESS_4 $PRIVKEY_4

# setup evaluation
FILE_NUM=100
FILE_SIZE=10K
BLOCK_NUM=100
PATH_TESTDATA=/tmp/test.dat
for i in `seq $FILE_NUM`; do
    $AIDE write-log "Upload test data (index:$i)"
    $AIDE testdata $PATH_TESTDATA $FILE_SIZE $i
    $HARNESS upload su1 $PATH_TESTDATA $BLOCK_NUM
done

TRIAL_COUNT=100
DAMAGE_RATE=0.003
for block_ratio in `seq 0.1 0.2 1.0`; do
    $AIDE write-log "Start frequency evaluation (DataRatio:$block_ratio, FileRatio:$FILE_RATIO, DamageRate:$DAMAGE_RATE)"
    for i in `seq $TRIAL_COUNT`; do
        $AIDE write-log "Start cycle (cycle:$i)"
        # ========================================================================
        $AIDE corruption $PATH_CACHE/sp $DAMAGE_RATE $PATH_CACHE/corrupted.list
        $HARNESS challenge su1 $block_ratio $FILE_RATIO
        $HARNESS proof su1
        $HARNESS --detected-list $PATH_CACHE/detected.list audit tpa1 su1
        $AIDE repair-batch $PATH_CACHE/detected.list
        rm -f $PATH_CACHE/detected.list
        # ========================================================================
        $AIDE write-log "Finish cycle (cycle:$i)"
    done
    $AIDE write-log "Finish frequency evaluation (DataRatio:$block_ratio, FileRatio:$FILE_RATIO, DamageRate:$DAMAGE_RATE)"
    $AIDE write-log "Repair all corrupted files"
    $AIDE repair-batch $PATH_CACHE/corrupted.list
done

cp $PATH_CACHE/prism.log /opt/prism/logs/frequency-$FILE_RATIO.log
