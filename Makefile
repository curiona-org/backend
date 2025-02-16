.DEFAULT_GOAL := help

help: ## Displays all the available commands
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
.PHONY: help

run: ## Runs the application
	@go run cmd/application/main.go
.PHONY: run

docker-compose: ## Starts the docker-compose services
	@docker-compose up -d

docker-compose-stop: ## Stops the docker-compose services
	@docker-compose stop

docker-compose-down: ## Stops and removes the docker-compose services
	@docker-compose down

migrate: ## Migrates the database to the most recent version
	@go run cmd/migrate/main.go -command up
.PHONY: migrate

migrate-down: ## Roll back the migration version by 1
	@go run cmd/migrate/main.go -command down
.PHONY: migrate-down

migrate-reset: ## Roll back all the migrations
	@go run cmd/migrate/main.go -command reset
.PHONY: migrate-reset

format: ## Runs go fmt and go vet
	@go fmt ./... && go vet -v ./...
.PHONY: format

test: ## Runs tests
	@go test \
		-shuffle=on \
		-count=1 \
		-short \
		-timeout=5m \
		./... \
		-coverprofile=./coverage/coverage.out
.PHONY: test

bench: ## Runs benchmarks
	@go test \
		-tags=benchmark \
		-bench=. \
		-benchmem \
		-run=^$$ \
		./...
.PHONY: bench

coverage: ## Shows the coverage of the tests
	@go tool cover -html=./coverage/coverage.out -o=./coverage/coverage.html
.PHONY: coverage

coverage-serve: coverage ## Serves the coverage report
	@serve ./coverage
.PHONY: coverage-serve

lint: ## Runs golangci-lint
	@golangci-lint run
.PHONY: lint