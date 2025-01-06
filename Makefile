SHELL := /bin/bash

MAKEFLAGS += --no-print-directory

-include ./cache/accounts.env
-include ./cache/contract-addr.env

# chose one from prism-sol/src
CONTRACT = XZ21

TRIAL_COUNT = 100

ETHERNET_OPTS = --server ws://testnet:8545 --contract $(CONTRACT_ADDR)
ETHERNET_SENDER_OPTS_0 = --sender-addr $(ADDRESS_0) --sender-key $(PRIVKEY_0)
ETHERNET_SENDER_OPTS_1 = --sender-addr $(ADDRESS_1) --sender-key $(PRIVKEY_1)
ETHERNET_SENDER_OPTS_2 = --sender-addr $(ADDRESS_2) --sender-key $(PRIVKEY_2)
ETHERNET_SENDER_OPTS_3 = --sender-addr $(ADDRESS_3) --sender-key $(PRIVKEY_3)
ETHERNET_SENDER_OPTS_4 = --sender-addr $(ADDRESS_4) --sender-key $(PRIVKEY_4)
ETHERNET_SENDER_OPTS_5 = --sender-addr $(ADDRESS_5) --sender-key $(PRIVKEY_5)

HARNESS_HOST_PATH = ./harness/volume
HARNESS_CONTAINER_PATH = /var/lib/prism-harness

.PHONY: eval

shell:
	docker compose run $(SERVICE) bash

upgrade:
	$(MAKE) harness@upgrade

ethcheck:
	$(MAKE) setup
	$(MAKE) testnet@down
	$(MAKE) testnet@up
	$(MAKE) ethcheck-main
	$(MAKE) testnet@down

eval:
# build programs
	$(MAKE) harness@build
	$(MAKE) aide@build
# perform evaluation: gentags
	$(MAKE) test-gentags
	$(MAKE) eval-gentags
# perform evaluation: auditing
	$(MAKE) test-auditing
	$(MAKE) eval-auditing
# perform evaluation: contract
	$(MAKE) test-contract
	$(MAKE) eval-contract
# perform evaluation: frequency
	$(MAKE) test-frequency
	$(MAKE) eval-frequency

test-gentags:
	$(MAKE) test-gentags-down
	rm -rf ./eval/gentags/logs/*
	docker compose -f docker-compose-eval-gentags.yaml up

test-auditing:
	$(MAKE) test-auditing-down
	rm -rf ./eval/auditing/logs/*
	docker compose -f docker-compose-eval-auditing.yaml up

test-contract:
	$(MAKE) test-contract-down
	rm -f ./cache/contract.addr
	rm -f ./cache/contract-addr.env
	docker compose -f docker-compose-eval-contract.yaml up -d testnet
	docker compose -f docker-compose-eval-contract.yaml exec testnet /entrypoint.sh deploy $(CONTRACT) $(PRIVKEY_0) $(ADDRESS_1) | tee ./cache/contract.addr
	echo CONTRACT_ADDR=`cat ./cache/contract.addr` > ./cache/contract-addr.env
	docker compose -f docker-compose-eval-contract.yaml up harness

test-gentags-down:
	docker compose -f docker-compose-eval-gentags.yaml down

test-auditing-down:
	docker compose -f docker-compose-eval-auditing.yaml down

test-contract-down:
	docker compose -f docker-compose-eval-contract.yaml down

test-frequency:
	$(MAKE) test-frequency-down
	rm -rf ./eval/frequency/logs/*
	docker compose -f docker-compose-eval-frequency.yaml --profile all up

test-frequency-x:
	rm -rf ./eval/frequency/logs/frequency-$(X_FILE_RATIO).log
	docker compose -f docker-compose-eval-frequency.yaml --profile $(X_FILE_RATIO) up

test-frequency-down:
	docker compose -f docker-compose-eval-frequency.yaml --profile all down

eval-gentags:
	rm -f ./eval/gentags/results/*
	$(MAKE) docker-run-eval MODE="eval-gentags" TYPE="gentags"

eval-auditing:
	rm -f ./eval/auditing/results/*
	$(MAKE) docker-run-eval MODE="eval-auditing" TYPE="auditing"

eval-contract:
	rm -f ./harness/app/eval/contract/results/*
	$(MAKE) docker-run-eval MODE="eval-contract" TYPE="contract"

eval-frequency:
	rm -f ./eval/frequency/results/*
	$(MAKE) docker-run-eval MODE="eval-frequency" TYPE="frequency"

testnet@up:
	docker compose up -d testnet
	$(MAKE) testnet@exec CMD="deploy $(CONTRACT) $(PRIVKEY_0) $(ADDRESS_1)" | tee ./cache/contract.addr
	echo CONTRACT_ADDR=`cat ./cache/contract.addr` > ./cache/contract-addr.env

testnet@down:
	docker compose down testnet

testnet@shell:
	$(MAKE) testnet@run CMD="sh"

harness@shell:
	$(MAKE) harness@run CMD="bash"

setup:
	$(MAKE) show-accounts > ./cache/accounts.env

harness@build-img:
	docker build -t prism-test/harness ./harness

harness@upgrade:
	$(MAKE) harness@run CMD="go get -u github.com/prism-pdp/prism-go"
	docker compose run -it --rm -v ./harness/app/go.mod:/opt/prism-harness/go.mod -v ./harness/app/go.sum:/opt/prism-harness/go.sum harness go get github.com/prism-pdp/prism-go
	$(MAKE) build-img

simcheck:
	rm -rf $(HARNESS_HOST_PATH)/*
	$(MAKE) harness@run CMD="harness --sim setup 0010 PRIVKEY_0 0011 PRIVKEY_1"
	$(MAKE) harness@run CMD="harness --sim enroll auditor tpa1 0012"
	$(MAKE) harness@run CMD="harness --sim enroll auditor tpa2 0013"
	$(MAKE) harness@run CMD="harness --sim enroll user    su1  0014 PRIVKEY_4"
	$(MAKE) harness@run CMD="harness --sim enroll user    su2  0015 PRIVKEY_5"
	$(MAKE) harness@run CMD="aide testdata $(HARNESS_CONTAINER_PATH)/cache/dummy.data 100K 1"
	$(MAKE) harness@run CMD="harness --sim upload su1 $(HARNESS_CONTAINER_PATH)/cache/dummy.data 100"
	$(MAKE) harness@run CMD="aide testdata $(HARNESS_CONTAINER_PATH)/cache/dummy.data 100K 2"
	$(MAKE) harness@run CMD="harness --sim upload su1 $(HARNESS_CONTAINER_PATH)/cache/dummy.data 100"
	$(MAKE) harness@run CMD="harness --sim upload su2 $(HARNESS_CONTAINER_PATH)/cache/dummy.data 50"
	$(MAKE) harness@run CMD="harness --sim challenge su1 0.55 1.0"
	$(MAKE) harness@run CMD="harness --sim proof su1"
	$(MAKE) harness@run CMD="harness --sim --detected-list $(HARNESS_CONTAINER_PATH)/cache/detected.list audit tpa1 su1"

ethcheck-main:
	rm -rf $(HARNESS_HOST_PATH)/*
	mkdir $(HARNESS_HOST_PATH)/cache
	fallocate -l 100M $(HARNESS_HOST_PATH)/cache/dummy.data
	$(MAKE) harness@run CMD="harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) setup $(ADDRESS_0) $(PRIVKEY_0) $(ADDRESS_1) $(PRIVKEY_1)"
	$(MAKE) harness@run CMD="harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) enroll auditor tpa1 $(ADDRESS_2)"
	$(MAKE) harness@run CMD="harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) enroll auditor tpa2 $(ADDRESS_3)"
	$(MAKE) harness@run CMD="harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) enroll user    su1  $(ADDRESS_4) $(PRIVKEY_4)"
	$(MAKE) harness@run CMD="harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_0) enroll user    su2  $(ADDRESS_5) $(PRIVKEY_5)"
	$(MAKE) harness@run CMD="harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_1) upload su1 $(HARNESS_CONTAINER_PATH)/cache/dummy.data 100"
	$(MAKE) harness@run CMD="harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_1) upload su2 $(HARNESS_CONTAINER_PATH)/cache/dummy.data 50"
	$(MAKE) harness@run CMD="harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_4) challenge su1 0.55 1.0"
	$(MAKE) harness@run CMD="harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_1) proof su1"
	$(MAKE) harness@run CMD="harness $(ETHERNET_OPTS) $(ETHERNET_SENDER_OPTS_2) audit tpa1 su1"

show-accounts:
	@$(MAKE) testnet@run CMD="show-accounts"

rpc:
	@echo METHOD:$(METHOD), PARAMS:[$(PARAMS)]
	@curl -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0", "method":"$(METHOD)", "params":[$(PARAMS)], "id":10}' 127.0.0.1:8545

rpc-test:
	@$(MAKE) --no-print-directory rpc METHOD="eth_accounts"

build-img:
	$(MAKE) harness@build-img

testnet@run:
	@docker compose run -it --rm testnet $(CMD)

testnet@exec:
	@docker compose exec testnet /entrypoint.sh $(CMD)

harness@run:
	@docker compose run -it --rm harness $(CMD)

logs:
	@docker compose logs -f

docker-run-eval:
	@docker run -it --rm -v ./eval:/var/lib/prism-harness/eval prism-test/harness aide $(MODE) /var/lib/prism-harness/eval/$(TYPE)/logs /var/lib/prism-harness/eval/$(TYPE)/results
