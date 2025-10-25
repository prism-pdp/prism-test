SHELL := /bin/bash

MAKEFLAGS += --no-print-directory

# chose one from prism-sol/src
CONTRACT = XZ21

TRIAL_COUNT = 100

HARNESS_HOST_PATH = ./harness/volume
HARNESS_CONTAINER_PATH = /var/lib/prism-harness

CACHE_ACCOUNTS = ./cache/accounts.env

-include ./cache/accounts.env
-include ./cache/contract-addr.env

$(CACHE_ACCOUNTS):
	$(MAKE) show-accounts > ./cache/accounts.env

simtest:
	cd test && ./run_simtest.sh

ethtest:
	$(MAKE) testnet@down
	$(MAKE) testnet@up
	cd test && ./run_ethtest.sh

experiment:
	$(MAKE) experiment-gentags
	$(MAKE) experiment-auditing
	$(MAKE) experiment-contract
	$(MAKE) experiment-frequency

experiment-gentags:
	$(MAKE) test-gentags
	$(MAKE) eval-gentags
	$(MAKE) test-gentags-down

experiment-auditing:
	$(MAKE) test-auditing
	$(MAKE) eval-auditing
	$(MAKE) test-auditing-down

experiment-contract:
	$(MAKE) test-contract
	$(MAKE) eval-contract
	$(MAKE) test-contract-down

experiment-frequency:
	$(MAKE) test-frequency
	$(MAKE) eval-frequency
	$(MAKE) test-frequency-down

test-gentags: $(CACHE_ACCOUNTS)
	$(MAKE) test-gentags-down
	rm -rf ./eval/gentags/logs/*
	docker compose -f docker-compose-eval-gentags.yaml up

test-auditing: $(CACHE_ACCOUNTS)
	$(MAKE) test-auditing-down
	rm -rf ./eval/auditing/logs/*
	docker compose -f docker-compose-eval-auditing.yaml up

test-contract: $(CACHE_ACCOUNTS)
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

test-frequency: $(CACHE_ACCOUNTS)
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
	$(MAKE) aide@eval MODE="eval-gentags" TYPE="gentags"

eval-auditing:
	rm -f ./eval/auditing/results/*
	$(MAKE) aide@eval MODE="eval-auditing" TYPE="auditing"

eval-contract:
	rm -f ./harness/app/eval/contract/results/*
	$(MAKE) aide@eval MODE="eval-contract" TYPE="contract"

eval-frequency:
	rm -f ./eval/frequency/results/*
	$(MAKE) aide@eval MODE="eval-frequency" TYPE="frequency"

graph-gentags:
	docker run --rm -v ./eval/gentags:/share -w /share \
		prism/graph \
		./run_make_graph.sh

graph-auditing:
	docker run --rm -v ./eval/auditing:/share -w /share \
		prism/graph \
		./run_make_graph.sh

build-graph:
	docker build -t prism/graph -f docker/Dockerfile.graph docker

upgrade:
	$(MAKE) harness@upgrade

show-accounts:
	@$(MAKE) testnet@run CMD="show-accounts"

build-img:
	$(MAKE) harness@build-img

rpc:
	@echo METHOD:$(METHOD), PARAMS:[$(PARAMS)]
	@curl -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0", "method":"$(METHOD)", "params":[$(PARAMS)], "id":10}' 127.0.0.1:8545

rpc-test:
	@$(MAKE) --no-print-directory rpc METHOD="eth_accounts"

testnet@up: $(CACHE_ACCOUNTS)
	docker compose up -d testnet
	$(MAKE) testnet@exec CMD="deploy $(CONTRACT) $(PRIVKEY_0) $(ADDRESS_1)" | tee ./cache/contract.addr
	echo CONTRACT_ADDR=`cat ./cache/contract.addr` > ./cache/contract-addr.env

testnet@down:
	docker compose down testnet

testnet@shell:
	$(MAKE) testnet@run CMD="sh"

testnet@exec:
	@docker compose exec testnet /entrypoint.sh $(CMD)

testnet@run:
	@docker compose run --rm testnet $(CMD)

harness@shell:
	$(MAKE) harness@run CMD="bash"

harness@build-img:
	docker build -t prism-test/harness ./harness

harness@run:
	@docker compose run -it --rm harness $(CMD)

harness@upgrade:
	docker run --rm -v ./harness/app:/opt/prism-harness prism-test/harness go get github.com/prism-pdp/prism-go
	$(MAKE) build-img

logs:
	@docker compose logs -f

aide@eval:
	@docker run -it --rm -v ./eval:/var/lib/prism-harness/eval prism-test/harness aide $(MODE) /var/lib/prism-harness/eval/$(TYPE)/logs /var/lib/prism-harness/eval/$(TYPE)/results
