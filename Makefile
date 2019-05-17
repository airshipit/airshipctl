SHELL := /bin/bash

GO_FLAGS       := -ldflags '-extldflags "-static"'

BINDIR         := bin
EXECUTABLE_CLI := airshipctl

SCRIPTS_DIR    := scripts

# linting
LINTER_CMD     := "github.com/golangci/golangci-lint/cmd/golangci-lint" run
ADDTL_LINTERS  := goconst,gofmt,unparam

# go options
PKG          := ./...
TESTS        := .

# coverage
COVER        := $(SCRIPTS_DIR)/coverage_test.sh
COVER_FILE   := cover.out
MIN_COVERAGE := 70


.PHONY: build
build:
	@CGO_ENABLED=0 go build -o $(BINDIR)/$(EXECUTABLE_CLI) $(GO_FLAGS)

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

.PHONY: docs
docs:
	@echo "TODO"

.PHONY: update-golden
update-golden: TESTFLAGS += -update -v
update-golden: PKG = github.com/ian-howell/airshipctl/cmd
update-golden:
	@go test $(PKG) $(TESTFLAGS)
