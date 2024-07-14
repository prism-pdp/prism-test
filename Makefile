MAKEFLAGS += --no-print-directory
.SILENT: show-contract-addr show-private-key

CONTRACT = BaseCounter

build-img:
	@docker compose build testnet
	@docker compose build harness

docker-run:
	docker compose run -it --rm $(SERVICE) $(CMD)

docker-exec:
	@docker compose exec $(SERVICE) /entrypoint.sh $(CMD)

shell:
	docker compose run $(SERVICE) bash

test:
	$(MAKE) testnet-shutdown
	$(MAKE) testnet-startup
	$(MAKE) harness-mkconf SERVER="http://testnet:8545" PRIV_KEY=$(file < cache/private.key) CONTRACT_ADDR=$(file < cache/contract.addr)
	$(MAKE) harness-run
	$(MAKE) testnet-shutdown

testnet-build:
	$(MAKE) docker-run SERVICE="testnet" CMD='forge build'

testnet-clean:
	$(MAKE) docker-run SERVICE="testnet" CMD='forge clean'

testnet-startup:
	$(MAKE) testnet-build
	docker compose up -d testnet
	$(MAKE) docker-exec SERVICE="testnet" CMD="deploy $(CONTRACT)"
	@$(MAKE) show-contract-addr | tee cache/contract.addr
	@$(MAKE) show-private-key   | tee cache/private.key

testnet-shutdown:
	docker compose down testnet

testnet-test:
	$(MAKE) docker-run SERVICE="testnet" CMD='forge test'

harness-run:
	$(MAKE) docker-run SERVICE="harness" CMD='go run main.go ./cache/config.json'

harness-mkconf:
	$(MAKE) docker-run SERVICE="harness" CMD="make-conf $(SERVER) $(PRIV_KEY) $(CONTRACT_ADDR)"

show-contract-addr:
	$(MAKE) docker-exec SERVICE="testnet" CMD="show-contract-addr"

show-private-key:
	$(MAKE) docker-exec SERVICE="testnet" CMD="show-private-key"

deploy:
	$(MAKE) docker-exec CMD="/entrypoint.sh deploy $(CONTRACT)"

init:
	$(MAKE) docker-run CMD='forge init --no-git .'

clean:
	$(MAKE) docker-run CMD='rm -rf ./volumes/testnet/*'

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

rpc:
	@echo METHOD:$(METHOD), PARAMS:[$(PARAMS)]
	@curl -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0", "method":"$(METHOD)", "params":[$(PARAMS)], "id":10}' 127.0.0.1:8545

rpc-test:
	@$(MAKE) --no-print-directory rpc METHOD="eth_accounts"

abigen:
	$(MAKE) docker-run CMD="abigen $(CONTRACT)"

startup:
	$(MAKE) up
	$(MAKE) deploy CONTRACT=$(CONTRACT) | tee cache/deploy.log
	$(MAKE) docker-run CMD="derive-private-key" | tee cache/derive-private-key.log
	$(MAKE) abigen CONTRACT=$(CONTRACT)
