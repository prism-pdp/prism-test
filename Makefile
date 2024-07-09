.SILENT: test

build-img:
	@docker compose build sol

docker-run:
	docker compose run -it --rm $(SERVICE) $(CMD)

docker-exec:
	@docker compose exec -it SERVICE=$(SERVICE) $(CMD)

sol-build:
	$(MAKE) docker-run SERVICE="sol" CMD='forge build'

sol-clean:
	$(MAKE) docker-run SERVICE="sol" CMD='forge clean'

sol-test:
	$(MAKE) docker-run SERVICE="sol" CMD='forge test'

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
