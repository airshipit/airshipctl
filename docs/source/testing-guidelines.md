# Testing Guidelines

This document lays out several guidelines to ensure quality and
consistency throughout `airshipctl`'s test bed.

## Testing packages

The `airshipctl` project uses the [testify] library, a thin wrapper around Go's
builtin `testing` package. The `testify` package provides the following
packages:

* `assert`: Functions from this package can be used to replace most calls to
  `t.Error`
* `require`: Contains the same functions as above, but these functions should
  replace calls to `t.Fatal`
* `mock`: Contains the `Mock` mechanism, granting the ability to mock out
  structs

## Test coverage

Tests should cover at least __80%__ of the codebase. Anything less will cause
the CI gates to fail. A developer should assert that their code meets this
criteria before submitting a change. This check can be performed with one of
the following `make` targets:

```
# Runs all unit tests, then computes and reports the coverage
make cover
```

Good practice is to assert that the changed packages have not decreased in
coverage. The coverage check can be run for a specific package with a command
such as the following.
```
make cover PKG=./pkg/foo
```

Additional testing should be done to ensure that the proposed change meets an
expected level of quality. These tests include:

```
# Tidy, to ensure go.mod is up to date
make tidy

# Lint, to ensure code meets linting requirements
make lint

# Update-golden, to ensure the golden test data reflects the current test cases
make update-golden
```

When the above are done, if you would like to perform the same dockerized container
testing as the CI gates you can do so via:

```
make docker-image-test-suite
```

**NOTE**: If test cases are deleted you must first run make update-golden, and
commit your changes prior to running the ``docker-image-test-suite`` make target.
Otherwise the ``docker-image-test-suite`` make target (and the CI job) will fail.

## Test directory structure

Test files end in `_test.go`, and sit next to the tested file. For example,
`airshipctl/pkg/foo/foo.go` should be tested by
`airshipctl/pkg/foo/foo_test.go`. A test's package name should also end in
`_test`, unless that file intends to test unexported fields and method, at
which point it should be in the package under test.

Go will ignore any files stored in a directory called `testdata`, therefore all
non-Go test files (such as expected output or example input) should be stored
there.

Any mocks for a package should be stored in a sub-package ending in `mocks`.
Each mocked struct should have its own file, where the filename describes the
struct, i.e. a file containing a mocked `Fooer` should be stored at
`mocks/fooer.go`. Mocked files can be either handwritten or generated via
[mockery]. The `mockery` tool can generate files in this fashion with the
following command.
```
mockery -all -case snake
```

An example file structure might look something like the following.
```
airshipctl/pkg/foo
├── foo.go
├── foo_test.go
├── mocks
│   └── fooer.go
└── testdata
    └── example-input.yaml
```

## Testing guidelines

This section annotates various standards for unit tests in `airshipctl`. These
should be thought of as "guidelines" rather than "rules".

* Using [table-tests] prevents a lot of duplicated code.
* Using [subtests] allows tests to provide much more fine-grained results.
* Calls to methods from `testify/require` be reserved for situations in which
  the test should fail immediately (e.g. during test setup). Generally,
  `testify/assert` should be preferred.

## How to write unit tests for files listed under the `cmd` package

Go files listed under the `cmd` package should be relatively slim. Their
purpose is to be a client of the `pkg` package. Most of these files will
contain no more than a single function which creates and returns a
`cobra.Command`. Nonetheless, these functions need to be tested. To help
alleviate some of the difficulties that come with testing a CLI, `airshipctl`
provides several helper structs and functions under the `testutil` package.

As an example, suppose you have the following function:

```
func NewVersionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version number of airshipctl",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			clientV := version.clientVersion()
			w := util.NewTabWriter(out)
			defer w.Flush()
			fmt.Fprintf(w, "%s:\t%s\n", "airshipctl", clientV)
		},
	}
	return versionCmd
}
```

Testing this functionality is easy with the use of the pre-built
`testutil.CmdTest`:

```
func TestVersion(t *testing.T) {
	versionCmd := cmd.NewVersionCommand()
	cmdTests := []*testutil.CmdTest{
		{
			Name:    "version",
			CmdLine: "",
			Cmd:     versionCmd,
			Error:   nil,
		},
		{
			Name:    "version-help",
			CmdLine: "--help",
			Cmd:     versionCmd,
			Error:   nil,
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}
```

The above test uses `CmdTest` structs, which are then fed to the `RunTest`
function.  This function provides abstraction around running a command on the
command line and comparing its output to a "golden file" (the pre-determined
expected output). The following describes the fields of the `CmdTest` struct.

* `Name` - The name for this test. This field *must* be unique, as it will be
  used while naming the golden file
* `CmdLine` - The arguments and flags to pass to the command
* `Cmd` - The actual instance of a `cobra.Command` to run. The above example
  reuses the command, but more complex tests may require different instances
  (e.g. to pass in a different `Settings` object)
* `Error` - The expected error for the command to return. This can be omitted
  if this test doesn't expect an error

Once you've written your test, you can generate the associated golden files by
running `make update-golden`, which invokes the "update" mode for
`testutil.RunTest`. When the command has completed, you can view the output in
the associated files in the `testdata` directory next to your command. Note
that these files are easily discoverable from the output of `git status`. When
you're certain that the golden files are correct, you can add them to the repo.

[mockery]: https://github.com/vektra/mockery
[subtests]: https://blog.golang.org/subtests
[table-tests]: https://github.com/golang/go/wiki/TableDrivenTests
[testify]: https://github.com/stretchr/testify
