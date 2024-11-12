SHELL := /bin/bash

MAKEFLAGS += --no-print-directory

include accounts.env
-include ./cache/contract-addr.env

# dpduado-sol/srcの中から選ぶ
CONTRACT = XZ21

TRIAL_COUNT = 500

ETHERNET_OPTS = --server ws://testnet:8545 --contract $(CONTRACT_ADDR)
ETHERNET_SENDER_OPTS_0 = --sender-addr $(ADDRESS_0) --sender-key $(PRIVKEY_0)
ETHERNET_SENDER_OPTS_1 = --sender-addr $(ADDRESS_1) --sender-key $(PRIVKEY_1)
ETHERNET_SENDER_OPTS_2 = --sender-addr $(ADDRESS_2) --sender-key $(PRIVKEY_2)
ETHERNET_SENDER_OPTS_3 = --sender-addr $(ADDRESS_3) --sender-key $(PRIVKEY_3)
ETHERNET_SENDER_OPTS_4 = --sender-addr $(ADDRESS_4) --sender-key $(PRIVKEY_4)
ETHERNET_SENDER_OPTS_5 = --sender-addr $(ADDRESS_5) --sender-key $(PRIVKEY_5)

shell:
	docker compose run $(SERVICE) bash

eval-offchain:
# build programs
	$(MAKE) harness@build
	$(MAKE) aide@build
# generate testdata
	$(MAKE) aide@testdata
# perform evaluation of generating tags
	$(MAKE) harness@test-gentags-all
	$(MAKE) aide@eval-gentags
# perform evaluation of generating proof and verifying proof
	$(MAKE) harness@test-auditing-all
	$(MAKE) aide@eval-auditing

eval-onchain:
# build programs
	$(MAKE) harness@build
	$(MAKE) aide@build
# perform evaluation of gas consumption of contracts
	$(MAKE) harness@test-contract-all
	$(MAKE) aide@eval-contract

aide@build:
	$(MAKE) docker-run SERVICE="harness" CMD="go build -o bin/aide ./cmd/aide"

aide@testdata:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide testdata $(FILE_PATH) $(FILE_SIZE) $(FILE_VAL)"

aide@genevaldata:
	@for i in `seq 10`; do \
		$(MAKE) aide@testdata FILE_PATH=./eval/testdata/100M-`printf %04X $$i`.dat FILE_SIZE=100M FILE_VAL=$$i; \
	done

aide@eval-gentags:
	rm -f ./harness/app/eval/gentags/results/*
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide eval-gentags ./eval/gentags/logs ./eval/gentags/results"

aide@eval-auditing:
	rm -f ./harness/app/eval/auditing/results/*
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide eval-auditing ./eval/auditing/logs ./eval/auditing/results"

aide@eval-contract:
	rm -f ./harness/app/eval/contract/results/*
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide eval-contract ./eval/contract/logs ./eval/contract/results"

test@sim:
	rm -rf ./harness/app/cache/*
	$(MAKE) harness@build
	$(MAKE) harness@run@sim

test:
	rm -rf ./harness/app/cache/*
	$(MAKE) testnet/down
	$(MAKE) testnet/up
	$(MAKE) harness@build
	$(MAKE) harness@run

test-clean:
	rm -rf ./cache/*
	$(MAKE) testnet/clean
	$(MAKE) harness/clean

testnet/build:
	$(MAKE) docker-run SERVICE="testnet" CMD='forge build'

testnet/clean:
	$(MAKE) docker-run SERVICE="testnet" CMD='forge clean'

testnet/up:
	docker compose up -d testnet
	$(MAKE) docker-exec SERVICE="testnet" CMD="deploy $(CONTRACT)" | tee ./cache/contract.addr
	echo CONTRACT_ADDR=`cat ./cache/contract.addr` > ./cache/contract-addr.env

testnet/down:
	docker compose down testnet

# Unused
testnet/test:
	$(MAKE) docker-run SERVICE="testnet" CMD='forge test'

testnet/shell:
	$(MAKE) docker-run SERVICE="testnet" CMD="sh"

testnet/login:
	$(MAKE) docker-exec SERVICE="testnet" CMD="sh"

harness/shell:
	$(MAKE) docker-run SERVICE="harness" CMD="bash"

setup:
	$(MAKE) show-accounts > accounts.env

harness@build:
	$(MAKE) docker-run SERVICE="harness" CMD="go build -o bin/harness ./cmd/dpduado"

harness@upgrade:
	sed -i '/github.com\/dpduado\/dpduado-go/d' harness/app/go.mod
	sed -i '/github.com\/dpduado\/dpduado-go/d' harness/app/go.sum
	$(MAKE) docker-run SERVICE="harness" CMD="go get github.com/dpduado/dpduado-go"
	$(MAKE) build-img

harness@simtest:
	rm -rf harness/app/cache/*
	fallocate -l 100M harness/app/cache/dummy.data
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log ./cache/log.txt setup $(ADDRESS_0) $(PRIVKEY_0) $(ADDRESS_1) $(PRIVKEY_1)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log ./cache/log.txt enroll auditor tpa1 $(ADDRESS_2)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log ./cache/log.txt enroll auditor tpa2 $(ADDRESS_3)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log ./cache/log.txt enroll user    su1  $(ADDRESS_4) $(PRIVKEY_4)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log ./cache/log.txt enroll user    su2  $(ADDRESS_5) $(PRIVKEY_5)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log ./cache/log.txt upload su1 cache/dummy.data 100"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log ./cache/log.txt upload su2 cache/dummy.data 50"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log ./cache/log.txt challenge su1 0.55"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log ./cache/log.txt proof"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log ./cache/log.txt audit tpa1"

harness@ethtest:
	$(MAKE) harness@ethtest-setup
	$(MAKE) testnet/down
	$(MAKE) testnet/up
	$(MAKE) harness@ethtest-main
	$(MAKE) testnet/down

harness@ethtest-setup:
	rm -rf ./harness/app/cache/*
	fallocate -l 100M harness/app/cache/dummy.data

harness@ethtest-main:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) --log ./cache/log.txt setup $(ADDRESS_0) $(PRIVKEY_0) $(ADDRESS_1) $(PRIVKEY_1)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) --log ./cache/log.txt enroll auditor tpa1 $(ADDRESS_2)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) --log ./cache/log.txt enroll auditor tpa2 $(ADDRESS_3)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) --log ./cache/log.txt enroll user    su1  $(ADDRESS_4) $(PRIVKEY_4)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) --log ./cache/log.txt enroll user    su2  $(ADDRESS_5) $(PRIVKEY_5)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_4) --log ./cache/log.txt upload su1 cache/dummy.data 100"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_5) --log ./cache/log.txt upload su2 cache/dummy.data 50"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_4) --log ./cache/log.txt challenge su1 0.55"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_1) --log ./cache/log.txt proof"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_2) --log ./cache/log.txt audit tpa1"

harness@test-gentags:
	rm -rf ./harness/app/cache/*
	$(eval BLOCK_NUM := $(SCALE)00)
	$(eval PATH_LOG := $(shell printf "./eval/gentags/logs/gentags-%04dM.log" $(BLOCK_NUM)))
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log $(PATH_LOG) setup $(ADDRESS_0) $(PRIVKEY_0) $(ADDRESS_1) $(PRIVKEY_1)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log $(PATH_LOG) enroll user su $(ADDRESS_4) $(PRIVKEY_4)"
	@for i in `seq 10`; do \
		rm -f ./harness/app/cache/test.dat; \
		$(MAKE) aide@testdata FILE_PATH=./cache/test.dat FILE_SIZE=$(BLOCK_NUM)M FILE_VAL=$$i; \
		$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log $(PATH_LOG) upload su ./cache/test.dat $(BLOCK_NUM)"; \
	done

harness@test-auditing:
	rm -rf ./harness/app/cache/*
	$(eval PATH_LOG := ./eval/auditing/logs/auditing-$(RATIO).log)
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log $(PATH_LOG) setup $(ADDRESS_0) $(PRIVKEY_0) $(ADDRESS_1) $(PRIVKEY_1)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log $(PATH_LOG) enroll auditor tpa $(ADDRESS_2)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log $(PATH_LOG) enroll user su $(ADDRESS_3) $(PRIVKEY_3)"
	$(MAKE) aide@testdata FILE_PATH="./cache/test.dat" FILE_SIZE=1000M FILE_VAL=255
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log $(PATH_LOG) upload su ./cache/test.dat 1000"
	@for i in `seq 10`; do \
		$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log $(PATH_LOG) challenge su $(RATIO)"; \
		$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log $(PATH_LOG) proof"; \
		$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --log $(PATH_LOG) audit tpa"; \
	done

harness@test-gentags-all:
	rm -f ./harness/app/eval/gentags/logs/*
	@for i in `seq 10`; do \
		$(MAKE) harness@test-gentags SCALE=$$i; \
	done

harness@test-auditing-all:
	rm -f ./harness/app/eval/gentags/logs/*
	@for i in `seq 0.1 0.1 1.0`; do \
		$(MAKE) harness@test-auditing RATIO=$$i; \
	done

harness@test-contract-all:
	$(eval PATH_LOG := ./eval/contract/logs/contract.log)
	rm -f ./harness/app/eval/contract/logs/*
	rm -rf ./harness/app/cache/*
	$(MAKE) testnet/down
	$(MAKE) testnet/up
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) --log $(PATH_LOG) setup $(ADDRESS_0) $(PRIVKEY_0) $(ADDRESS_1) $(PRIVKEY_1)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) --log $(PATH_LOG) enroll auditor tpa1 $(ADDRESS_2)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) --log $(PATH_LOG) enroll user    su1  $(ADDRESS_4) $(PRIVKEY_4)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) --log $(PATH_LOG) enroll user    su2  $(ADDRESS_5) $(PRIVKEY_5)"
	for i in `seq $(TRIAL_COUNT)`; do \
		$(MAKE) aide@testdata FILE_PATH="./cache/test.dat" FILE_SIZE=10M FILE_VAL=$$i; \
		$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_4) --log $(PATH_LOG) upload su1 ./cache/test.dat 10"; \
		$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_5) --log $(PATH_LOG) upload su2 ./cache/test.dat 10"; \
	done
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_4) --log $(PATH_LOG) challenge su1 0.6"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_1) --log $(PATH_LOG) proof"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_2) --log $(PATH_LOG) audit tpa1"
	$(MAKE) testnet/down

harness@run:
	$(MAKE) harness@run-setup
	$(MAKE) harness@run-upload
	$(MAKE) harness@run-audit
harness@run-setup:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness eth setup ws://testnet:8545 $(CONTRACT_ADDR)"
harness@run-upload:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness eth upload ws://testnet:8545 $(CONTRACT_ADDR)"
harness@run-audit:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness eth audit ws://testnet:8545 $(CONTRACT_ADDR)"

harness/clean:
	rm -rf ./harness/app/cache/*

show-accounts:
	@$(MAKE) docker-run SERVICE="testnet" CMD="show-accounts"

rpc:
	@echo METHOD:$(METHOD), PARAMS:[$(PARAMS)]
	@curl -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0", "method":"$(METHOD)", "params":[$(PARAMS)], "id":10}' 127.0.0.1:8545

rpc-test:
	@$(MAKE) --no-print-directory rpc METHOD="eth_accounts"

build-img:
	@docker compose build

docker-run:
	@docker compose run -it --rm $(SERVICE) $(CMD)

docker-exec:
	@docker compose exec $(SERVICE) /entrypoint.sh $(CMD)

docker-log:
	@docker compose logs -f
