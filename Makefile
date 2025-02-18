##
# Makefile configuration
# See: https://www.gnu.org/software/make/manual/make.html
.ONESHELL:
.DEFAULT_GOAL = help

.PHONY: build
build: ## Build the source files into a single binary.
	./scripts/build.sh

.PHONY: checks
checks: ## Run a series of source code checks.
	./scripts/checks.sh

.PHONY: docker
docker: ## Run the Docker image build.
	./scripts/docker.sh

.PHONY: start
start: ## Launch the service
	./scripts/start.sh

.PHONY: test
test: ## Run the unit test suite.
	./scripts/test.sh

.PHONY: help
help:  ## Print this help.
	grep -E '^[a-z.A-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Silence output by default, use `VERBOSE=1 make <command>` to enable.
ifndef VERBOSE
.SILENT:
endif