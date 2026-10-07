.PHONY: run test generate mocks openapi fmt vet check

ifneq (,$(wildcard .env))
include .env
export
endif

# === App ===

run: ## Run application locally
	go run ./cmd/server

test: ## Format, vet and run tests with the race detector
	$(MAKE) fmt
	$(MAKE) vet
	go test -race ./...

generate: ## Generate mocks and OpenAPI server code
	$(MAKE) mocks
	$(MAKE) openapi

mocks:
	go tool mockery

openapi:
	go generate ./internal/http/openapi

fmt:
	gofmt -w cmd internal

vet:
	go vet ./...

.PHONY: help # http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ##
	@awk '/^# === .* ===$$/ { section=$$0; sub(/^# === /, "", section); sub(/ ===$$/, "", section); printf "\n\033[1;33m== %s ==\033[0m\n", section; next } /^[a-zA-Z0-9_-]+:.*## .*$$/ { target=$$0; sub(/:.*/, "", target); description=$$0; sub(/^.*## /, "", description); printf "\033[36m%-30s\033[0m %s\n", target, description }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
