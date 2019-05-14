SHELL := /bin/bash

BINDIR         := bin
EXECUTABLE_CLI := airshipctl

SCRIPTS_DIR    := scripts

PLUGIN_DIR     := _plugins
PLUGIN_BIN := $(PLUGIN_DIR)/bin
PLUGIN_INT := $(patsubst $(PLUGIN_DIR)/internal/%,$(PLUGIN_BIN)/%.so,$(wildcard $(PLUGIN_DIR)/internal/*))
PLUGIN_EXT := $(wildcard $(PLUGIN_DIR)/external/*)

# linting
LINTER_CMD     := "github.com/golangci/golangci-lint/cmd/golangci-lint" run
ADDTL_LINTERS  := goconst,gofmt,lll,unparam

# go options
PKG          := ./...
TESTS        := .

# coverage
COVER        := $(SCRIPTS_DIR)/coverage_test.sh
COVER_FILE   := cover.out
MIN_COVERAGE := 70


.PHONY: build
build:
	@go build -o $(BINDIR)/$(EXECUTABLE_CLI)

.PHONY: test
test: build
test: lint
test: TESTFLAGS += -race -v
test: unit-tests
test: cover

.PHONY: unit-tests
unit-tests: build
	@echo "Performing unit test step..."
	@go test -run $(TESTS) $(PKG) $(TESTFLAGS) -coverprofile=$(COVER_FILE)
	@echo "All unit tests passed"

.PHONY: cover
cover: unit-tests
	@./$(COVER) $(COVER_FILE) $(MIN_COVERAGE)


.PHONY: lint
lint:
	@echo "Performing linting step..."
	@go run ${LINTER_CMD} --enable ${ADDTL_LINTERS}
	@echo "Linting completed successfully"

.PHONY: clean
clean:
	@rm -fr $(BINDIR)
	@rm -fr $(COVER_FILE)

.PHONY: plugin-clean
plugin-clean:
	@rm -fr $(PLUGIN_BIN)

.PHONY: docs
docs:
	@echo "TODO"

.PHONY: update-golden
update-golden: TESTFLAGS += -update -v
update-golden: PKG = github.com/ian-howell/airshipctl/cmd
update-golden:
	@go test $(PKG) $(TESTFLAGS)

.PHONY: plugin
plugin: $(PLUGIN_INT)
	@for plugin in $(PLUGIN_EXT); do $(MAKE) -C $${plugin}; done

$(PLUGIN_BIN)/%.so: $(PLUGIN_DIR)/*/%/*.go
	@go build -buildmode=plugin -o $@ $^
