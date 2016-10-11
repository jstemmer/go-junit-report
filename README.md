# go-junit-report

Converts `go test` output to an xml report, suitable for applications that
expect junit xml reports (e.g. [Jenkins](http://jenkins-ci.org)).

[![Build Status](https://travis-ci.org/improbable-io/go-junit-report.svg)](https://travis-ci.org/improbable-io/go-junit-report)


## Installation

Go version 1.1 or higher is required. Install or update using the `go get`
command:

	go get -u github.com/improbable-io/go-junit-report

## Usage

go-junit-report reads the `go test` verbose output from standard in and writes
junit compatible XML to standard out.

	go test -v | go-junit-report > report.xml

