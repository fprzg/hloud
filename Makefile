include config/.env-api

## help: print this message
.PHONY: help
help:
	@echo 'Usage'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]


## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	cd ./api && go run ./cmd --port=4000

## run/frontned: run the cmd/front
.PHONY: run/frontend
run/frontend:
	cd ./frontend && go run ./cmd/main.go --port=4001
