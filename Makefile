BINDIR         := bin
EXECUTABLE_CLI := airshipadm

.PHONY: build
build:
	go build -o $(BINDIR)/$(EXECUTABLE_CLI)

.PHONY: test
test: build
test: unit-tests
test: lint

.PHONY: unit-tests
unit-tests:
	go test

.PHONY: lint
lint:
	@echo "TODO"

.PHONY: clean
clean:
	@rm -fr $(BINDIR)

.PHONY: docs
docs:
	@echo "TODO"
