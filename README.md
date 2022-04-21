# go-junit-report

go-junit-report is a tool that converts [`go test`] output to an XML report,
suitable for applications that expect JUnit-style XML reports (e.g. [Jenkins]).

The test output [parser] and JUnit report [formatter] are also available as Go
packages.

[![Build status][github-actions-badge]][github-actions-link]

## Install from package (recommended)

Pre-built packages for Windows, macOS and Linux are found on the [Releases]
page.

## Install from source

Download and install the latest stable version from source by running:

```bash
go install github.com/jstemmer/go-junit-report@latest
```

## Usage

go-junit-report reads the full `go test` output from stdin and writes JUnit
compatible XML to stdout. In order to capture build errors as well as test
output, redirect both stdout and stderr to go-junit-report.

```bash
go test -v 2>&1 | go-junit-report > report.xml
```

Parsing benchmark output is also supported, for example:

```bash
go test -v -bench . -count 5 2>&1 | go-junit-report > report.xml
```

If you want go-junit-report to exit with a non-zero exit code when it encounters
build errors or test failures, set the `-set-exit-code` flag.

Run `go-junit-report -help` for a list of all supported flags.

If you want to know the current version, run:

```bash
go-junit-report -version
> go-junit-report v1.0.0-dev  (HEAD)
```

## Contributing

See [CONTRIBUTING.md].

[`go test`]: https://pkg.go.dev/cmd/go#hdr-Test_packages
[Jenkins]: https://www.jenkins.io/
[parser]: https://pkg.go.dev/github.com/jstemmer/go-junit-report/parser
[formatter]: https://pkg.go.dev/github.com/jstemmer/go-junit-report/formatter
[github-actions-badge]: https://github.com/jstemmer/go-junit-report/actions/workflows/main.yml/badge.svg
[github-actions-link]: https://github.com/jstemmer/go-junit-report/actions
[Releases]: https://github.com/jstemmer/go-junit-report/releases
[CONTRIBUTING.md]: https://github.com/jstemmer/go-junit-report/blob/master/CONTRIBUTING.md
