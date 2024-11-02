MAKEFLAGS += --no-print-directory

include accounts.env

# dpduado-sol/srcの中から選ぶ
CONTRACT = XZ21

shell:
	docker compose run $(SERVICE) bash

aide@build:
	$(MAKE) docker-run SERVICE="harness" CMD="go build -o bin/aide ./cmd/aide"

aide@testdata:
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide testdata ./cache/1m-0100.dat 1M 100"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide testdata ./cache/1m-0200.dat 1M 200"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide testdata ./cache/1m-0300.dat 1M 300"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide testdata ./cache/1m-0400.dat 1M 400"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide testdata ./cache/1m-0500.dat 1M 500"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide testdata ./cache/1m-0600.dat 1M 600"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide testdata ./cache/1m-0700.dat 1M 700"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide testdata ./cache/1m-0800.dat 1M 800"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide testdata ./cache/1m-0900.dat 1M 900"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/aide testdata ./cache/1m-1000.dat 1M 1000"

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
	$(MAKE) docker-run SERVICE="harness" CMD="go build -o bin/harness ./cmd/dpduado"

harness@upgrade:
	sed -i '/github.com\/dpduado\/dpduado-go/d' harness/app/go.mod
	sed -i '/github.com\/dpduado\/dpduado-go/d' harness/app/go.sum
	$(MAKE) docker-run SERVICE="harness" CMD="go get github.com/dpduado/dpduado-go"
	$(MAKE) build-img

harness@test-sim:
	rm -rf harness/app/cache/*
	fallocate -l 100M harness/app/cache/dummy.data
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim setup $(ADDRESS_0) $(PRIVKEY_0) $(ADDRESS_1) $(PRIVKEY_1)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim enroll auditor tpa1 $(ADDRESS_2)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim enroll auditor tpa2 $(ADDRESS_3)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim enroll user    su1  $(ADDRESS_4) $(PRIVKEY_4)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim enroll user    su2  $(ADDRESS_5) $(PRIVKEY_5)"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim upload su1 cache/dummy.data 100"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim upload su2 cache/dummy.data 50"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim challenge su1"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim proof"
	$(MAKE) docker-run SERVICE="harness" CMD="./bin/harness --sim audit tpa1"

harness@eval-tag:

#harness@eval-proof:

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
