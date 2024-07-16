MAKEFLAGS += --no-print-directory
.SILENT: show-contract-addr show-private-key

CONTRACT = BaseCounter

shell:
	docker compose run $(SERVICE) bash

test:
	$(MAKE) testnet/shutdown
	$(MAKE) testnet/startup
	$(MAKE) harness/mkconf SERVER="http://testnet:8545"
	$(MAKE) harness/run
	$(MAKE) testnet/shutdown

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
	$(MAKE) docker-exec SERVICE="testnet" CMD="deploy $(CONTRACT)"
	@$(MAKE) show-contract-addr | tee cache/contract.addr
	@$(MAKE) show-private-key   | tee cache/private.key

testnet/shutdown:
	docker compose down testnet

testnet/test:
	$(MAKE) docker-run SERVICE="testnet" CMD='forge test'

harness/shell:
	$(MAKE) docker-run SERVICE="harness" CMD="bash"

harness/run:
	$(MAKE) docker-run SERVICE="harness" CMD='go run main.go ./cache/config.json'

harness/mkconf:
	$(MAKE) docker-run SERVICE="harness" CMD="make-conf $(SERVER) $(file < cache/private.key) $(file < cache/contract.addr)"

harness/clean:
	rm -rf ./harness/app/cache/*

show-contract-addr:
	$(MAKE) docker-exec SERVICE="testnet" CMD="show-contract-addr"

show-private-key:
	$(MAKE) docker-exec SERVICE="testnet" CMD="show-private-key"

rpc:
	@echo METHOD:$(METHOD), PARAMS:[$(PARAMS)]
	@curl -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0", "method":"$(METHOD)", "params":[$(PARAMS)], "id":10}' 127.0.0.1:8545

rpc-test:
	@$(MAKE) --no-print-directory rpc METHOD="eth_accounts"

build-img:
	@docker compose build

docker-run:
	docker compose run -it --rm $(SERVICE) $(CMD)

docker-exec:
	@docker compose exec $(SERVICE) /entrypoint.sh $(CMD)
