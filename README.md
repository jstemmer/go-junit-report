go-junit-report
===============

Converts `go test` output to an xml report, suitable for applications that
expect junit xml reports (e.g. [Jenkins](http://jenkins-ci.org)).

Installation
------------

	go get github.com/jstemmer/go-junit-report

Usage
-----

	go test -v | go-junit-report > report.xml

