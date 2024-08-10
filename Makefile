MAKEFLAGS += --no-print-directory

# dpduado-sol/srcの中から選ぶ
CONTRACT = XZ21

shell:
	docker compose run $(SERVICE) bash

test/sim:
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

testnet/startup:
	docker compose up -d testnet
	@$(MAKE) docker-exec SERVICE="testnet" CMD="deploy $(CONTRACT)" | tee ./cache/contract.addr

testnet/shutdown:
	docker compose down testnet

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
	$(MAKE) docker-run SERVICE="harness" CMD="go build -o bin/harness ."

harness@run:
	$(MAKE) harness@run-setup
	$(MAKE) harness@run-upload

harness@run-setup:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness http://testnet:8545 $(file < cache/contract.addr) setup"

harness@run-upload:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness http://testnet:8545 $(file < cache/contract.addr) upload"

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
