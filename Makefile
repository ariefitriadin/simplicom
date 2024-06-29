# This version-strategy uses git tags to set the version string
VERSION := $(shell git describe --tags --always --dirty)
GIT_COMMIT := $(shell git rev-list -1 HEAD)

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

version: ## Show version
	@echo $(VERSION) \(git commit: $(GIT_COMMIT)\)

sqlc-auth-gen:
	@echo "Running migrations Auth Service"
	sqlc -f cmd/auth-service/internal/app/config/sqlc.yaml generate

migration-auth: ## Run migrations Auth Service
	@echo "Running migrations Auth Service, example: make migration-auth OP=up, or OP=down"
	dbmate -d cmd/auth-service/migrations ${OP}

migration-user: ## Run migrations User Service
	@echo "Running migrations Order Service"
	dbmate -d cmd/order-service/migrations ${OP}
