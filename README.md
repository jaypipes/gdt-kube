# Go Declarative Testing - Kubernetes ![go test workflow](https://github.com/jaypipes/gdt-kube/actions/workflows/gate-tests.yml/badge.svg)

[`gdt`][gdt] is a testing library that allows test authors to cleanly describe tests
in a YAML file. `gdt` reads YAML files that describe a test's assertions and
then builds a set of Go structures that the standard Go
[`testing`](https://golang.org/pkg/testing/) package can execute.

[gdt]: https://github.com/jaypipes/gdt

This `gdt-kube` repository is a companion Go library for `gdt` that allows test
authors to cleanly describe functional tests of Kubernetes resources and
actions using a simple, clear YAML format. `gdt-kube` parses YAML files that
describe Kubernetes client/API requests and assertions about those client
calls.

## Installation

`gdt-kube` is a Golang library and is intended to be included in your own Golang
application's test code as a Golang package dependency.

Install `gdt-kube` into your `$GOPATH` by executing:

```
go get -u github.com/jaypipes/gdt-kube
```

## `gdt-kube` test file structure

TODO

## Contributing and acknowledgements

`gdt` was inspired by [Gabbi](https://github.com/cdent/gabbi), the excellent
Python declarative testing framework. `gdt` tries to bring the same clear,
concise test definitions to the world of Go functional testing.

Contributions to `gdt-kube` are welcomed! Feel free to open a Github issue or
submit a pull request.
