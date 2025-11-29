include ./server/.env

CURRENT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
SHELL = /bin/sh

.PHONY: help
help:        ## Print available targets.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

.PHONY: run
run:         ## Start production web server.
	@make start-containers

.PHONY: dev
dev:		## Start development database.
	@echo "Starting app containers"
	@docker compose --env-file ./server/.env up tibia_char_database

.PHONY: stop
stop:        ## Stop development web server.
	@docker compose --env-file ./server/.env down

.PHONY: clean
clean:       ## Clearing existing data.
	@echo "Clearing existing data"
	@docker compose down --volumes --env-file ./server/.env up

.PHONY: start-containers
start-containers:
	@echo "Starting app containers"
	@docker compose --env-file ./server/.env up

.PHONY: create-migration
create-migration:
	@echo "Creating migrations"
	@cd ./server; ./migrate create -ext sql -dir database/migrations -seq $(name)

.PHONY: migrate
migrate:
	@echo "Executing migrations"
	@cd ./server; ./migrate -database ${DATABASE_URL} -path database/migrations up

.PHONY: rollback
rollback:
	@echo "Executing migrations"
	@cd ./server; ./migrate -database ${DATABASE_URL} -path database/migrations down