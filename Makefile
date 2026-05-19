.PHONY: test lint build bench docker clean release help

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD)
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS  = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: ## Run all tests with race detection and coverage
	go test -v -race -coverprofile=coverage.out $$(go list ./... | grep -v '/cmd/')

coverage: test ## Display coverage report
	go tool cover -func=coverage.out | tee coverage.txt
	@echo ""
	@COVERAGE=$$(grep total coverage.txt | awk '{print $$3}' | sed 's/%//'); \
	if (( $$(echo "$$COVERAGE < 60.0" | bc -l) )); then \
		echo "Coverage $$COVERAGE% is below minimum 60%"; \
		exit 1; \
	fi; \
	echo "Coverage: $$COVERAGE%"

lint: ## Run golangci-lint
	golangci-lint run --timeout=5m ./...

build: ## Build the grimorio binary
	go build -ldflags "$(LDFLAGS)" -v -o grimorio ./cmd/grimorio

build-migrate: ## Build the migrate-v1-to-v2 binary
	go build -ldflags "$(LDFLAGS)" -v -o migrate-v1-to-v2 ./cmd/migrate-v1-to-v2

bench: ## Run benchmarks
	bash scripts/bench.sh run

bench-compare: ## Compare benchmarks against baseline
	bash scripts/bench.sh compare

bench-save: ## Save current benchmarks as baseline
	bash scripts/bench.sh save-baseline

docker: ## Build Docker image
	docker build -t grimorio:mcp-$(VERSION) .

clean: ## Clean build artifacts
	rm -f grimorio migrate-v1-to-v2 coverage.out coverage.txt
	rm -rf dist/

release: ## Create release binaries (requires scripts/release.sh)
	bash scripts/release.sh

changelog: ## Preview changelog for unreleased commits
	git-cliff --config cliff.toml --unreleased --tag "next"

changelog-update: ## Generate and prepend changelog for a specific tag
	@if [ -z "$(TAG)" ]; then \
		echo "Usage: make changelog-update TAG=v3.4.0"; \
		exit 1; \
	fi
	git-cliff --config cliff.toml --unreleased --tag "$(TAG)" --prepend CHANGELOG.md

changelog-all: ## Regenerate full changelog from all tags
	git-cliff --config cliff.toml --tag "next" > CHANGELOG.md
