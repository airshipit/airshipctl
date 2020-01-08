# Developers Guide

This guide explains how to set up your environment for developing on
airshipctl.

## Environment Expectations

- Go 1.13
- Docker
- Git

## Building airshipctl

We use Make to build our programs. The simplest way to get started is:

```console
$ make build
```

NOTE: The airshipctl application is a go module.  This means you do not need to
clone the repository into `$GOPATH` in order to build it.  You should be able to
build it from any directory.

This will build the airshipctl binary.

To run all the tests including linting and coverage reports, run `make test`. To
run all tests in a containerized environment, run `make docker-image-unit-tests`
or `make docker-image-lint`

To run airshipctl locally, you can run `bin/airshipctl`.

## Docker Images

If you want to build an `airshipctl` Docker image, run `make docker-image`.

Pre-built images are already available at [quay.io](
  https://quay.io/airshipit/airshipctl).

## Contribution Guidelines

We welcome contributions. This project has set up some guidelines in order to
ensure that (a) code quality remains high, (b) the project remains consistent,
and \(c\) contributions follow the open source legal requirements. Our intent
is not to burden contributors, but to build elegant and high-quality open source
code so that our users will benefit.

Make sure you have read and understood the main airshipctl [Contributing
Guide](../../CONTRIBUTING.md)

### Structure of the Code

The code for the airshipctl project is organized as follows:

- The individual programs are located in `cmd/`. Code inside of `cmd/` is not
  designed for library re-use.
- Shared libraries are stored in `pkg/`.
- Both commands and shared libraries may require test data fixtures. These
  should be placed in a `testdata/` subdirectory within the command or library.
- The `testutil/` directory contains functions that are helpful for unit tests.
- The `zuul.d/` directory contains Zuul YAML definitions for CI/CD jobs to run.
- The `playbooks/` directory contains playbooks that the Zuul CI/CD jobs will
  run.
- The `tools/` directory contains scripts used by the Makefile and CI/CD
  pipeline.
- The `docs/` folder is used for documentation and examples.

Go dependencies are managed by `go mod` and stored in `go.mod` and `go.sum`

### Git Conventions

We use Git for our version control system. The `master` branch is the home of
the current development candidate. Releases are tagged.

We accept changes to the code via Gerrit pull requests. One workflow for doing
this is as follows:

1. `git clone` the `opendev.org/airship/airshipctl` repository.
2. Create a new working branch (`git checkout -b feat/my-feature`) and do your
   work on that branch.
3. When you are ready for us to review, push your branch to gerrit using
   `git-review`.  For more information on the gerrit workflow, see the [OpenDev
   documentation](
     https://docs.openstack.org/contributors/common/setup-gerrit.html).

### Go Conventions

We follow the Go coding style standards very closely. Typically, running `go
fmt` will make your code beautiful for you.

We also typically follow the conventions of `golangci-lint`.

Read more:

- Effective Go [introduces
  formatting](https://golang.org/doc/effective_go.html#formatting).
- The Go Wiki has a great article on
  [formatting](https://github.com/golang/go/wiki/CodeReviewComments).

### Testing

In order to ensure that all package unit tests follow the same standards and use
the same frameworks, airshipctl has a document outlining [specific test
guidelines](testing-guidelines.md) maintained separately.
