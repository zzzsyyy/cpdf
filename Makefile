.PHONY: help deps dev build release clean

CGO_ENABLED=0
VERSION=$(shell git describe --abbrev=0 --tags 2>/dev/null || echo "$version")
COMMIT=$(shell git rev-parse --short HEAD || echo "$commit")
GOCMD=go

all: help

deps: ## Install any required dependencies

ifeq ($(filter dev,$(MAKECMDGOALS)),dev)
    DEBUG = 1
endif

dev: build ## Build debug version

help: ## Show help message
	@echo "cpdf - Tools of processing PDF"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

build: ## build cli
ifeq ($(DEBUG), 1)
	@echo "Building in debug mode..."
	@$(GOCMD) build \
		-ldflags "\
		-X main.Version=$(VERSION) \
		-X main.Commit=$(COMMIT)" \
		.
else
	@$(GOCMD) build \
		-ldflags "-w -s \
		-X main.Version=$(VERSION) \
		-X main.Commit=$(COMMIT)" \
		.
endif
	@./cpdf -v

fmt: ## Format sources files
	@$(GOCMD) fmt ./...

clean: ## Remove untracked files
	@git clean -f -d -x
