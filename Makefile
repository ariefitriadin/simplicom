# This version-strategy uses git tags to set the version string
VERSION := $(shell git describe --tags --always --dirty)
GIT_COMMIT := $(shell git rev-list -1 HEAD)
DBPATHAUTH := cmd/auth-service/migrations
DBPRDPATH := cmd/product-service/migrations
DATABASE_AUTH=postgres://user:secret@localhost:5432/authdb?sslmode=disable
DATABASE_PRODUCT=postgres://user:secret@localhost:5432/productdb?sslmode=disable

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

sqlc-prd-gen:
	@echo "Running migrations Product Service"
	sqlc -f cmd/product-service/internal/app/config/sqlc.yaml generate

migration-auth: ## Run migrations Auth Service
	@echo "Running migrations Auth Service, example: make migration-auth OP=up, or OP=down"
	dbmate -d cmd/auth-service/migrations -u ${DATABASE_AUTH} ${OP}

migration-prd: ## Run migrations Product Service
	@echo "Running migrations Product Service, example: make migration-prd OP=up, or OP=down"
	dbmate -d cmd/product-service/migrations -u ${DATABASE_PRODUCT} ${OP}

dbnewtable: ## generate new table  TABLE=create_products_table
	@echo "generate new table"
	dbmate -d $(DBPRDPATH) new ${TABLE}
