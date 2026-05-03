include ./app/.env

CURRENT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
SHELL = /bin/sh

.PHONY: help
help:        ## Print available targets.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

.PHONY: create-migration
create-migration:
	@echo "Creating migrations"
	@cd ./app; ./migrate create -ext sql -dir database/migrations -seq $(name)

.PHONY: migrate
migrate:
	@echo "Executing migrations"
	@cd ./app; ./migrate -database ${DATABASE_URL} -path database/migrations up

.PHONY: rollback
rollback:
	@echo "Executing migrations"
	@cd ./app; ./migrate -database ${DATABASE_URL} -path database/migrations down