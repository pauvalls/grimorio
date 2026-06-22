.PHONY: test lint build bench docker clean release help dev fmt tidy

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD)
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS  = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

dev: ## Start air for hot-reload development
	@command -v air >/dev/null 2>&1 || { \
		echo "air is not installed. Install with: go install github.com/air-verse/air@latest"; \
		exit 1; \
	}
	air

fmt: ## Run gofmt and goimports on all .go files
	gofmt -w .
	@command -v goimports >/dev/null 2>&1 || { \
		echo "goimports not installed. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
		exit 1; \
	}
	goimports -w .

tidy: ## Run go mod tidy and verify
	go mod tidy
	go mod verify

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

install: ## Install grimorio via install.sh (binary) or source build fallback
ifeq ($(wildcard install.sh),install.sh)
	bash install.sh
else
	@echo "install.sh not found, falling back to source build..."
	go build -ldflags "$(LDFLAGS)" -v -o grimorio ./cmd/grimorio
	mkdir -p $(HOME)/.local/bin
	cp grimorio $(HOME)/.local/bin/grimorio
	mkdir -p $(HOME)/.config/opencode/plugins/grimorio
	cp grimorio $(HOME)/.config/opencode/plugins/grimorio/grimorio
	@echo "Installed grimorio from source to $(HOME)/.local/bin/grimorio"
endif

update: ## Update grimorio via grimorio update (fallback to install.sh --update)
	@which grimorio >/dev/null 2>&1 && grimorio update || bash install.sh --update

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

release-tag: ## Auto-detect next version and create/push tag
	@echo "Determining next version from conventional commits..."
	@LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	MAJOR=$$(echo "$$LATEST_TAG" | sed 's/v//' | cut -d. -f1); \
	MINOR=$$(echo "$$LATEST_TAG" | sed 's/v//' | cut -d. -f2); \
	PATCH=$$(echo "$$LATEST_TAG" | sed 's/v//' | cut -d. -f3); \
	HAS_BREAKING=$$(git log "$$LATEST_TAG..HEAD" --oneline | grep -c "!:" 2>/dev/null || echo 0); \
	HAS_FEAT=$$(git log "$$LATEST_TAG..HEAD" --oneline | grep -c "feat" 2>/dev/null || echo 0); \
	echo "Latest: $$LATEST_TAG  Breaking: $$HAS_BREAKING  Feat: $$HAS_FEAT"; \
	if [ "$$HAS_BREAKING" -gt 0 ]; then \
		MAJOR=$$((MAJOR + 1)); \
		NEW_TAG="v$$MAJOR.0.0"; \
	elif [ "$$HAS_FEAT" -gt 0 ]; then \
		MINOR=$$((MINOR + 1)); \
		NEW_TAG="v$$MAJOR.$$MINOR.0"; \
	else \
		PATCH=$$((PATCH + 1)); \
		NEW_TAG="v$$MAJOR.$$MINOR.$$PATCH"; \
	fi; \
	echo "Creating tag: $$NEW_TAG"; \
	git tag -a "$$NEW_TAG" -m "Release $$NEW_TAG"; \
	git push origin "$$NEW_TAG"
