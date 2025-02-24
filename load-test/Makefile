VUS ?= 100
DURATION ?= "30s"
K6 := k6

.PHONY: help
help: ## Displays all the available commands
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: health
health: ## Checks the health of the application
	${K6} run --summary-trend-stats "min,avg,max,p(50),p(95),p(99)" src/health.ts

.PHONY: register
register: ## Registers a new user
	${K6} run --summary-trend-stats "min,avg,max,p(50),p(95),p(99)" src/register.ts