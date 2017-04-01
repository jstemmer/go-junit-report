# go-junit-report

Converts `go test` output to an xml report, suitable for applications that
expect junit xml reports (e.g. [Jenkins](http://jenkins-ci.org)).

[![Build Status][travis-badge]][travis-link]
[![Report Card][report-badge]][report-link]

## Installation

Go version 1.1 or higher is required. Install or update using the `go get`
command:

```bash
go get -u github.com/jstemmer/go-junit-report
```

## Usage

go-junit-report reads the `go test` verbose output from standard in and writes
junit compatible XML to standard out.

```bash
go test -v 2>&1 | go-junit-report > report.xml
```

[travis-badge]: https://travis-ci.org/jstemmer/go-junit-report.svg
[travis-link]: https://travis-ci.org/jstemmer/go-junit-report
[report-badge]: https://goreportcard.com/badge/github.com/jstemmer/go-junit-report
[report-link]: https://goreportcard.com/report/github.com/jstemmer/go-junit-report
