SHELL := /bin/bash

MAKEFLAGS += --no-print-directory

include accounts.env
-include ./cache/contract-addr.env

# chose one from dpduado-sol/src
CONTRACT = XZ21

TRIAL_COUNT = 100

ETHERNET_OPTS = --server ws://testnet:8545 --contract $(CONTRACT_ADDR)
ETHERNET_SENDER_OPTS_0 = --sender-addr $(ADDRESS_0) --sender-key $(PRIVKEY_0)
ETHERNET_SENDER_OPTS_1 = --sender-addr $(ADDRESS_1) --sender-key $(PRIVKEY_1)
ETHERNET_SENDER_OPTS_2 = --sender-addr $(ADDRESS_2) --sender-key $(PRIVKEY_2)
ETHERNET_SENDER_OPTS_3 = --sender-addr $(ADDRESS_3) --sender-key $(PRIVKEY_3)
ETHERNET_SENDER_OPTS_4 = --sender-addr $(ADDRESS_4) --sender-key $(PRIVKEY_4)
ETHERNET_SENDER_OPTS_5 = --sender-addr $(ADDRESS_5) --sender-key $(PRIVKEY_5)

shell:
	docker compose run $(SERVICE) bash

eval-all:
# build programs
	$(MAKE) harness@build
	$(MAKE) aide@build
# perform evaluation
	$(MAKE) eval-gentags
	$(MAKE) eval-auditing
	$(MAKE) eval-contract
	$(MAKE) eval-frequency

eval-gentags:
	$(MAKE) test-gentags
	$(MAKE) aide@eval-gentags

eval-auditing:
	$(MAKE) test-auditing
	$(MAKE) aide@eval-auditing

eval-contract:
	$(MAKE) test-contract
	$(MAKE) aide@eval-contract

eval-frequency:
	$(MAKE) eval-frequency-down
	rm -rf ./eval/frequency/logs/*
	docker compose -f docker-compose-eval-frequency.yaml --profile all up
	$(MAKE) aide@eval-frequency

eval-frequency-x:
	rm -rf ./eval/frequency/logs/frequency-$(X_FILE_RATIO).log
	docker compose -f docker-compose-eval-frequency.yaml --profile $(X_FILE_RATIO) up

eval-frequency-down:
	docker compose -f docker-compose-eval-frequency.yaml --profile all down

test-gentags:
	rm -f ./harness/app/eval/gentags/logs/*
	@for block_num in `seq 100 100 1000`; do \
		$(MAKE) test-gentags-main X_TRIAL_COUNT=$(TRIAL_COUNT) X_BLOCK_NUM=$$block_num; \
		mv ./harness/app/cache/dpduado.log ./harness/app/eval/gentags/logs/gentags-$$block_num.log; \
	done

test-gentags-main:
	rm -rf ./harness/app/cache/*
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim setup $(ADDRESS_0) $(PRIVKEY_0) $(ADDRESS_1) $(PRIVKEY_1)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim enroll user su $(ADDRESS_4) $(PRIVKEY_4)"
	@for i in `seq $(X_TRIAL_COUNT)`; do \
		rm -f ./harness/app/cache/test.dat; \
		$(MAKE) aide@testdata FILE_PATH=./cache/test.dat FILE_SIZE=$(X_BLOCK_NUM)M FILE_VAL=$$i; \
		$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim upload su ./cache/test.dat $(X_BLOCK_NUM)"; \
		rm -f ./harness/app/cache/sp/*.dat; \
		rm -f ./harness/app/cache/sp/*.dat.tag; \
	done

test-auditing:
	rm -f ./harness/app/eval/auditing/logs/*
	@for ratio in `seq 0.1 0.1 1.0`; do \
		$(MAKE) test-auditing-main X_TRIAL_COUNT=$(TRIAL_COUNT) X_BLOCK_RATIO=$$ratio; \
		mv ./harness/app/cache/dpduado.log ./harness/app/eval/auditing/logs/auditing-$$ratio.log; \
	done

test-auditing-main:
	rm -rf ./harness/app/cache/*
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim setup $(ADDRESS_0) $(PRIVKEY_0) $(ADDRESS_1) $(PRIVKEY_1)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim enroll auditor tpa $(ADDRESS_2)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim enroll user su $(ADDRESS_3) $(PRIVKEY_3)"
	$(MAKE) aide@testdata FILE_PATH="./cache/test.dat" FILE_SIZE=1000M FILE_VAL=255
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim upload su ./cache/test.dat 1000"
	@for i in `seq $(X_TRIAL_COUNT)`; do \
		$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim challenge su $(X_BLOCK_RATIO) 1.0"; \
		$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim proof"; \
		$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim audit tpa"; \
	done

test-contract:
	rm -f ./harness/app/eval/contract/logs/*
	$(MAKE) test-contract-main X_TRIAL_COUNT=$(TRIAL_COUNT)
	mv ./harness/app/cache/dpduado.log ./harness/app/eval/contract/logs/contract.log

test-contract-main:
	rm -rf ./harness/app/cache/*
	$(MAKE) testnet/down
	$(MAKE) testnet/up
	@$(MAKE) harness@cmd-setup
	@$(MAKE) harness@cmd-enroll-auditor X_NAME=tpa1 X_ADDR=$(ADDRESS_2)
	@$(MAKE) harness@cmd-enroll-user    X_NAME=su1  X_ADDR=$(ADDRESS_4) X_KEY=$(PRIVKEY_4)
	@$(MAKE) harness@cmd-enroll-user    X_NAME=su2  X_ADDR=$(ADDRESS_5) X_KEY=$(PRIVKEY_5)
	@for i in `seq $(X_TRIAL_COUNT)`; do \
		$(MAKE) aide@testdata FILE_PATH="./cache/test.dat" FILE_SIZE=1M FILE_VAL=$$i; \
		$(MAKE) harness@cmd-upload X_ETHERNET_SENDER_OPTS="$(ETHERNET_SENDER_OPTS_4)" X_USER_NAME=su1 X_PATH_FILE=./cache/test.dat X_SPLIT_COUNT=10; \
		$(MAKE) harness@cmd-upload X_ETHERNET_SENDER_OPTS="$(ETHERNET_SENDER_OPTS_5)" X_USER_NAME=su2 X_PATH_FILE=./cache/test.dat X_SPLIT_COUNT=10; \
	done
	@$(MAKE) harness@cmd-challenge X_ETHERNET_SENDER_OPTS="$(ETHERNET_SENDER_OPTS_4)" X_USER_NAME=su1 X_DATA_RATIO=0.6 X_FILE_RATIO=1.0
	@$(MAKE) harness@cmd-proof
	@$(MAKE) harness@cmd-audit     X_ETHERNET_SENDER_OPTS="$(ETHERNET_SENDER_OPTS_2)" X_AUDITOR_NAME=tpa1
	$(MAKE) testnet/down

aide@build:
	$(MAKE) docker-run SERVICE="harness" CMD="go build -o bin/aide ./cmd/aide"

aide@testdata:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide testdata $(FILE_PATH) $(FILE_SIZE) $(FILE_VAL)"

aide@eval-gentags:
	rm -f ./harness/app/eval/gentags/results/*
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide eval-gentags ./eval/gentags/logs ./eval/gentags/results"

aide@eval-auditing:
	rm -f ./harness/app/eval/auditing/results/*
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide eval-auditing ./eval/auditing/logs ./eval/auditing/results"

aide@eval-contract:
	rm -f ./harness/app/eval/contract/results/*
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide eval-contract ./eval/contract/logs ./eval/contract/results"

aide@eval-frequency:
	rm -f ./eval/frequency/results/*
	@docker compose -f docker-compose-eval.yaml run --rm aide aide eval-frequency /opt/dpduado/eval/frequency/logs /opt/dpduado/eval/frequency/results

aide@corruption:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide corruption $(X_DIR_TARGET) $(X_DAMAGE_RATE) $(X_PATH_RESULT)"

aide@list-corrupted-files:
	@$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide list-corrupted-files $(X_TPA_NAME)"

aide@repair:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide repair $(X_PATH_FILE)"

aide@repair-batch:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide repair-batch $(X_PATH_LIST)"

aide@write-log:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide write-log \"$(X_LOG)\""

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
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim setup $(ADDRESS_0) $(PRIVKEY_0) $(ADDRESS_1) $(PRIVKEY_1)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim enroll auditor tpa1 $(ADDRESS_2)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim enroll auditor tpa2 $(ADDRESS_3)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim enroll user    su1  $(ADDRESS_4) $(PRIVKEY_4)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim enroll user    su2  $(ADDRESS_5) $(PRIVKEY_5)"
	$(MAKE) aide@testdata FILE_PATH=./cache/dummy.data FILE_SIZE=1M FILE_VAL=1
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim upload su1 cache/dummy.data 100"
	$(MAKE) aide@testdata FILE_PATH=./cache/dummy.data FILE_SIZE=1M FILE_VAL=2
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim upload su1 cache/dummy.data 100"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim upload su2 cache/dummy.data 50"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim challenge su1 0.55 1.0"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim proof"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim --detected-list ./cache/detected.list audit tpa1"

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
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) setup $(ADDRESS_0) $(PRIVKEY_0) $(ADDRESS_1) $(PRIVKEY_1)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) enroll auditor tpa1 $(ADDRESS_2)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) enroll auditor tpa2 $(ADDRESS_3)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) enroll user    su1  $(ADDRESS_4) $(PRIVKEY_4)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) enroll user    su2  $(ADDRESS_5) $(PRIVKEY_5)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_4) upload su1 cache/dummy.data 100"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_5) upload su2 cache/dummy.data 50"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_4) challenge su1 0.55 1.0"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_1) proof"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_2) audit tpa1"

harness@cmd-setup:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) setup $(ADDRESS_0) $(PRIVKEY_0) $(ADDRESS_1) $(PRIVKEY_1)"

harness@cmd-enroll-auditor:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) enroll auditor $(X_NAME) $(X_ADDR)"

harness@cmd-enroll-user:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) enroll user    $(X_NAME) $(X_ADDR) $(X_KEY)"

harness@cmd-upload:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(X_ETHERNET_SENDER_OPTS) upload $(X_USER_NAME) $(X_PATH_FILE) $(X_SPLIT_COUNT)"

harness@cmd-challenge:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(X_ETHERNET_SENDER_OPTS) challenge $(X_USER_NAME) $(X_DATA_RATIO) $(X_FILE_RATIO)"

harness@cmd-proof:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_1) proof"

harness@cmd-audit:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness $(ETHERNET_OPTS) $(X_ETHERNET_SENDER_OPTS) audit $(X_AUDITOR_NAME)"

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
