#!/bin/bash

set -euo pipefail

HARNESS_CONTAINER_PATH="/var/lib/prism-harness"

(
    cd ..
	make harness@run CMD="clean"
	make harness@run CMD="harness --sim setup 0010 PRIVKEY_0 0011 PRIVKEY_1"
	make harness@run CMD="harness --sim enroll auditor tpa1 0012"
	make harness@run CMD="harness --sim enroll auditor tpa2 0013"
	make harness@run CMD="harness --sim enroll user    su1  0014 PRIVKEY_4"
	make harness@run CMD="harness --sim enroll user    su2  0015 PRIVKEY_5"
	make harness@run CMD="aide testdata ${HARNESS_CONTAINER_PATH}/cache/dummy.data 100K 1"
	make harness@run CMD="harness --sim upload su1 ${HARNESS_CONTAINER_PATH}/cache/dummy.data 100"
	make harness@run CMD="aide testdata ${HARNESS_CONTAINER_PATH}/cache/dummy.data 100K 2"
	make harness@run CMD="harness --sim upload su1 ${HARNESS_CONTAINER_PATH}/cache/dummy.data 100"
	make harness@run CMD="harness --sim upload su2 ${HARNESS_CONTAINER_PATH}/cache/dummy.data 50"
	make harness@run CMD="harness --sim challenge su1 0.55 1.0"
	make harness@run CMD="harness --sim proof su1"
	make harness@run CMD="harness --sim --detected-list ${HARNESS_CONTAINER_PATH}/cache/detected.list audit tpa1 su1"
)
