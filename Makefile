BINDIR         := bin
EXECUTABLE_CLI := airshipadm

# linting
LINTER_CMD     := "github.com/golangci/golangci-lint/cmd/golangci-lint" run
ADDTL_LINTERS  := goconst,gofmt,lll,unparam

# go options
PKG        := ./...
TESTS      := .


.PHONY: build
build:
	@go build -o $(BINDIR)/$(EXECUTABLE_CLI)

.PHONY: test
test: build
test: lint
test: TESTFLAGS += -race -v
test: unit-tests

.PHONY: unit-tests
unit-tests:
	@echo "Performing unit test step..."
	@go test -run $(TESTS) $(PKG) $(TESTFLAGS)
	@echo "All unit tests passed"

.PHONY: lint
lint:
	@echo "Performing linting step..."
	@go run ${LINTER_CMD} --enable ${ADDTL_LINTERS}
	@echo "Linting completed successfully"

.PHONY: clean
clean:
	@rm -fr $(BINDIR)

.PHONY: docs
docs:
	@echo "TODO"
