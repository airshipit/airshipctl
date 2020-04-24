# Developer's Guide

This guide explains how to set up your environment for developing on
airshipctl.

## Environment expectations

- Git
- Go 1.13
- Docker

### Installing Git

Instructions to install Git are [here][12].

### Installing Go 1.13

Instructions to install Golang are [here][13].

### Installing Docker

Instructions to install Docker are [here][14].

Additionally, there is a script in the airshipctl directory named as
`00_setup.sh`. This script will download all the required binaries and
packages. This script can be checked [here][1].

### Proxy

If your organization requires development behind a proxy server, you will need
to define the following environment variables with your organization's
information:

```
HTTP_PROXY=http://username:password@host:port
HTTPS_PROXY=http://username:password@host:port
NO_PROXY="localhost,127.0.0.1"
USE_PROXY=true
```

When running the gate scripts in `tools/gate` locally, you will need to add your
proxy information to
`roles/airshipctl-systemwide-executable/defaults/main.yaml`.

## Building airshipctl

The simplest way to get started is:

```sh
git clone https://opendev.org/airship/airshipctl.git
```

NOTE: The airshipctl application is a Go module. This means that there is no
need to clone the repository into the $GOPATH directory in order to build it.
You should be able to build it from any directory as long as $GOPATH is
defined correctly.

The following command will build the airshipctl binary.

```sh
make build
```

This will compile airshipctl and place the resulting binary into the bin
directory. To run all the tests including linting and coverage reports, run
`make test`. To run all tests in a containerized environment, run
`make docker-image-test-suite`.

## Docker Images

If you want to build an `airshipctl` Docker image, run `make docker-image`.
Pre-built images are already available at [quay.io][2]. Moreover, in the
directory `airshipctl/tools/gate/`, different scripts are present which will
run and download all the required images. The script [10_build_gate.sh][3]
will download all the required images.

## Contribution Guidelines

We welcome contributions. This project has set up some guidelines in order to
ensure that

- code quality remains high
- the project remains consistent, and
- contributions follow the open source legal requirements.

Our intent is not to burden contributors, but to build elegant and
high-quality open source code so that our users will benefit.
Make sure you have read and understood the main airshipctl
[Contributing Guide][4].

## Structure of the Code

The code for the airshipctl project is organized as follows:

- The client-facing code is located in `cmd/`. Code inside of `cmd/` is not
designed for library reuse.
- Shared libraries are stored in `pkg/`.
- Both commands and shared libraries may require test data fixtures. These
should be placed in a `testdata/` subdirectory within the command or library.
- The `testutil/` directory contains functions that are helpful for unit
tests.
- The `zuul.d/` directory contains Zuul YAML definitions for CI/CD jobs to
run.
- The `playbooks/` directory contains playbooks that the Zuul CI/CD jobs will
run.
- The `tools/` directory contains scripts used by the Makefile and CI/CD
pipeline.
- The `tools/gate` directory consists of different scripts. These scripts
will setup the environment as per requirements and install all the required
packages and binaries. This will also download all the required docker images.
- The `docs/` folder is used for documentation and examples.
- Go dependencies are managed by `go mod` and stored in `go.mod` and `go.sum`

## Git Conventions

We use Git for our version control system. The `master` branch is the home of
the current development candidate. Releases are tagged.
We accept changes to the code via Gerrit pull requests. One workflow for doing
this is as follows:

1. `git clone` the `airshipctl` repository. For this run the command:

    ```sh
    git clone https://opendev.org/airship/airshipctl.git
    ```

2. Use [OpenDev documentation][5] to setup Gerrit with the repo.

3. When set, use [this guide][6] to learn the OpenDev development workflow,
in a sandbox environment. You can then apply the learnings to start developing
airshipctl.

## Go Conventions

We follow the Go coding style standards very closely. Typically, running
`goimports -w -local opendev.org/airship/airshipctl ./` will make your code
beautiful for you.

We also typically follow the conventions of `golangci-lint`.
Read more:

- Effective Go [introduces formatting][7].
- The Go Wiki has a great article on [formatting][8].

## Testing

In order to ensure that all package unit tests follow the same standard and
use the same frameworks, airshipctl has a document outlining
[specific test guidelines][9] maintained separately.
Moreover, there are few scripts in directory `tools/gate` which run different
tests. The script [20_run_test_runner.sh][10] will generate airshipctl config
file and verify cluster and kubectl version. The script
[22_test_config.sh][11] will generate kube config, set airshipctl config and
verify the cluster.

[1]: https://github.com/airshipit/airshipctl/blob/master/tools/gate/00_setup.sh
[2]: https://quay.io/airshipit/airshipctl
[3]: https://github.com/airshipit/airshipctl/blob/master/tools/gate/10_build_gate.sh
[4]: https://github.com/airshipit/airshipctl/blob/master/CONTRIBUTING.md
[5]: https://docs.openstack.org/contributors/common/setup-gerrit.html
[6]: https://docs.opendev.org/opendev/infra-manual/latest/sandbox.html
[7]: https://golang.org/doc/effective_go.html#formatting
[8]: https://github.com/golang/go/wiki/CodeReviewComments
[9]: https://github.com/airshipit/airshipctl/blob/master/docs/source/testing-guidelines.md
[10]: https://github.com/airshipit/airshipctl/blob/master/tools/gate/20_run_test_runner.sh
[11]: https://github.com/airshipit/airshipctl/blob/master/tools/gate/22_test_configs.sh
[12]: https://git-scm.com/book/en/v2/Getting-Started-Installing-Git
[13]: https://golang.org/doc/install
[14]: https://docs.docker.com/get-docker/
