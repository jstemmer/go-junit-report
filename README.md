# doctest-junit-report

Converts [doctest]\(for c++) output to an xml report, suitable for applications that
expect junit xml reports (e.g. [Jenkins](http://jenkins-ci.org)).

[![Build Status][travis-badge]][travis-link]
[![Go Report Card][report-badge]][reoprt-link]
[![codecov][codecov-badge]][codecov-link]

## Installation

You can get binary from github release page.

-> [Release Page][release-link]

or, use `go get`:

```bash
go get -u github.com/ujiro99/doctest-junit-report
```

## Usage

doctest-junit-report reads the output of doctest binary from standard in and writes
junit compatible XML to standard out.

```bash
${test_binary} -s -d | doctest-junit-report > report.xml
```

[doctest]: https://github.com/onqtam/doctest
[travis-badge]: https://travis-ci.org/ujiro99/doctest-junit-report.svg?branch=master
[travis-link]: https://travis-ci.org/ujiro99/doctest-junit-report
[report-badge]: https://goreportcard.com/badge/github.com/ujiro99/doctest-junit-report
[reoprt-link]: https://goreportcard.com/report/github.com/ujiro99/doctest-junit-report
[codecov-badge]: https://codecov.io/gh/ujiro99/doctest-junit-report/branch/master/graph/badge.svg
[codecov-link]: https://codecov.io/gh/ujiro99/doctest-junit-report
[release-link]: https://github.com/ujiro99/doctest-junit-report/releases/latest
