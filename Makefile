BINDIR         := bin
EXECUTABLE_CLI := airshipadm

# go options
PKG        := ./...
TESTS      := .

.PHONY: build
build:
	go build -o $(BINDIR)/$(EXECUTABLE_CLI)

.PHONY: test
test: build
test: TESTFLAGS += -race -v
test: unit-tests
test: lint

.PHONY: unit-tests
unit-tests:
	go test -run $(TESTS) $(PKG) $(TESTFLAGS)

.PHONY: lint
lint:
	@echo "TODO"

.PHONY: clean
clean:
	@rm -fr $(BINDIR)

.PHONY: docs
docs:
	@echo "TODO"
