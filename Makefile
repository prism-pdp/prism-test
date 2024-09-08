MAKEFLAGS += --no-print-directory

# dpduado-sol/srcの中から選ぶ
CONTRACT = XZ21

shell:
	docker compose run $(SERVICE) bash

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
	@$(MAKE) docker-exec SERVICE="testnet" CMD="deploy $(CONTRACT)" | tee ./cache/contract.addr

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
	$(MAKE) docker-run SERVICE="harness" CMD="go build -o bin/harness ."

harness@run@sim:
	$(MAKE) harness@run-setup@sim
	$(MAKE) harness@run-upload@sim
	$(MAKE) harness@run-audit@sim
harness@run-setup@sim:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness sim setup"
harness@run-upload@sim:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness sim upload"
harness@run-audit@sim:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness sim audit"

harness@run:
	$(MAKE) harness@run-setup
	$(MAKE) harness@run-upload
	$(MAKE) harness@run-audit
harness@run-setup:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness eth setup ws://testnet:8545 $(file < cache/contract.addr)"
harness@run-upload:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness eth upload ws://testnet:8545 $(file < cache/contract.addr)"
harness@run-audit:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness eth audit ws://testnet:8545 $(file < cache/contract.addr)"

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
