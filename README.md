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

## Usage

`gdt-kube` is a Go library and is intended to be included in your own Go
application's test code as a Go package dependency.

Import the `gdt` and `gdt-kube` libraries in a Go test file:

```go
import (
    "github.com/jaypipes/gdt"
    gdtkube "github.com/jaypipes/gdt-kube"
)
```

In a standard Go test function, use the `gdt.From()` function to instantiate a
test object (either a `Scenario` or a `Suite`) that can be `Run()` with a
standard Go `context.Context` and a standard Go `*testing.T` type:

```go
func TestExample(t *testing.T) {
    s, err := gdt.From("path/to/test.yaml")
    if err != nil {
        t.Fatalf("failed to load tests: %s", err)
    }

    ctx := context.Background()
    err = s.Run(ctx, t)
    if err != nil {
        t.Fatalf("failed to run tests: %s", err)
    }
}
```

To execute the tests, just run `go test` per the standard Go testing practice.

`gdt` is a *declarative testing framework* and the meat of your tests is going
to be in the YAML files that describe the actions and assertions for one or
more tests. Read on for an explanation of how to write tests in this
declarative YAML format.

## `gdt-kube` test file structure

A `gdt` test scenario (or just "scenario") is simply a YAML file.

All `gdt` scenarios have the following fields:

* `name`: (optional) string describing the contents of the test file. If
  missing or empty, the filename is used as the name
* `description`: (optional) string with longer description of the test file
  contents
* `defaults`: (optional) is a map, keyed by a plugin name, of default options
  and configuration values for that plugin.
* `require`: (optional) list of strings indicating named fixtures that will be
  started before any of the tests in the file are run
* `tests`: list of [`Spec`][basespec] specializations that represent the
  runnable test units in the test scenario.

[basespec]: https://github.com/jaypipes/gdt-core/blob/e1d23e0974447de0bcd273f151edebeebc2b96c6/spec/spec.go#L27-L39

### `gdt-kube` test configuration defaults

To set `gdt-kube`-specific default configuration values for the test scenario,
set the `defaults.kube` field to an object containing any of these fields:

* `defaults.kube.config`: (optional) file path to a `kubeconfig` to use for the
  test scenario.
* `defaults.kube.context`: (optional) string containing the name of the kube
  context to use for the test scenario.
* `defaults.kube.namespace`: (optional) string containing the Kubernetes
  namespace to use when performing some action for the test scenario.

As an example, let's say that I wanted to override the Kubernetes namespace and
the kube context used for a particular test scenario. I would do the following:

```yaml
name: example-test-with-defaults
defaults:
  kube:
    context: my-kube-context
    namespace: my-namespace
```

### `gdt-kube` test spec structure

All `gdt` test specs have the same [base fields][base-spec-fields]:

* `name`: (optional) string describing the test unit.
* `description`: (optional) string with longer description of the test unit.
* `timeout`: (optional) an object containing [timeout information][timeout] for the test
  unit.
* `timeout.after`: a string duration of time the test unit is expected to
  complete within.
* `timeout.expected`: a bool indicating that the test unit is expected to not
  complete before `timeout.after`. This is really only useful in unit testing.

[base-spec-fields]: https://github.com/jaypipes/gdt#gdt-test-spec-structure
[timeout]: https://github.com/jaypipes/gdt-core/blob/e1d23e0974447de0bcd273f151edebeebc2b96c6/types/timeout.go#L11-L22

`gdt-kube` test specs have some additional fields that allow you to take some
action against a Kubernetes API and assert that the response from the API
matches some expectation:

* `config`: (optional) file path to the `kubeconfig` to use for this specific
  test. This allows you to override the `defaults.config` value from the test
  scenario.
* `context`: (optional) string containing the name of the kube context to use
  for this specific test. This allows you to override the `defaults.context`
  value from the test scenario.
* `namespace`: (optional) string containing the name of the Kubernetes
  namespace to use when performing some action for this specific test. This
  allows you to override the `defaults.namespace` value from the test scenario.
* `kube`: (optional) an object containing actions and assertions the test takes
  against the Kubernetes API server.
* `kube.get`: (optional) string containing either a resource specifier (e.g.
  `pods`, `po/nginx` or a file path to a YAML manifest containing resources
  that will be read from the Kubernetes API server.
* `kube.create`: (optional) string containing either a file path to a YAML
  manifest or a string of raw YAML containing the resource(s) to create.
* `kube.delete`: (optional) string containing either a resource specifier (e.g.
  `pods`, `po/nginx` or a file path to a YAML manifest containing resources
  that will be deleted.
* `kube.assert`: (optional) object containing assertions to make about the
  action performed by the test.
* `kube.assert.error`: (optional) string to match a returned error from the
  Kubernetes API server.
* `kube.assert.len`: (optional) int with the expected number of items returned.
* `kube.assert.notfound`: (optional) bool indicating the test author expects
  the Kubernetes API to return a 404/Not Found for a resource.
* `kube.assert.unknown`: (optional) bool indicating the test author expects the
  Kubernetes API server to respond that it does not know the type of resource
  attempting to be fetched or created.

Here are some examples of `gdt-kube` tests.

Testing that a Pod with the name `nginx` exists:

```yaml
name: test-nginx-pod-exists
tests:
 - kube:
     get: pods/nginx
 # These are equivalent. "kube.get" is a shortcut for the longer object.field
 # form above.
 - kube.get: pods/nginx
```

Testing that a Pod with the name `nginx` *does not* exist:

```yaml
name: test-nginx-pod-not-exist
tests:
 - kube:
     get: pods/nginx
     assert:
       notfound: true
```

Testing that a Pod with the name `nginx` exists by the specified timeout
(essentially, `gdt-kube` will retry the get call and assertion until the end of
the timeout):

```yaml
name: test-nginx-pod-exists-within-1-minute
tests:
 - kube:
     get: pods/nginx
     timeout: 1m
```

Testing creation and subsequent fetch then delete of a Pod, specifying the Pod
definition contained in a YAML file:

```yaml
name: create-get-delete-pod
description: create, get and delete a Pod
require:
  - kind
tests:
  - name: create-pod
    kube:
      create: manifests/nginx-pod.yaml
  - name: pod-exists
    kube:
      get: pods/nginx
  - name: delete-pod
    kube:
      delete: pods/nginx
```

Testing creation and subsequent fetch then delete of a Pod, specifying the Pod
definition using an inline YAML blob:

```yaml
name: create-get-delete-pod
description: create, get and delete a Pod
require:
  - kind
tests:
  # "kube.create" is a shortcut for the longer object->field format
  - kube.create: |
        apiVersion: v1
        kind: Pod
        metadata:
          name: nginx
        spec:
          containers:
          - name: nginx
            image: nginx
            imagePullPolicy: IfNotPresent
  # "kube.get" is a shortcut for the longer object->field format
  - kube.get: pods/nginx
  # "kube.delete" is a shortcut for the longer object->field format
  - kube.delete: pods/nginx
```

You can mix other `gdt` test types in a single `gdt` test scenario. For
example, here we are testing the creation of a Pod, waiting a little while,
then using the `gdt` `exec` test type to test SSH connectivity to the Pod.

```yaml
name: create-check-ssh
description: create a Deployment then check SSH connectivity
require:
  - kind
tests:
  - kube.create: manifests/deployment.yaml
  - exec: sleep 30
  - exec: ssh -T someuser@ip
```

## Determining Kubernetes config, context and namespace values

When evaluating how to construct a Kubernetes client `gdt-kube` uses the following
precedence to determine the `kubeconfig` and kube context:

1) The individual test spec's `config` or `context` value
2) Any `gdt` Fixture that exposes a `gdt.kube.config` or `gdt.kube.context`
   state key (e.g. [`KindFixture`][kind-fixture]).
3) The test file's `defaults.kube` `config` or `context` value.

For the `kubeconfig` file path, if none of the above yielded a value, the
following precedence is used to determine the `kubeconfig`:

4) A non-empty `KUBECONFIG` environment variable pointing at a file.
5) In-cluster config if running in cluster.
6) `$HOME/.kube/config` if it exists.

[kube-fixture]: https://github.com/jaypipes/gdt-kube/blob/main/fixtures/kind/kind.go

## `gdt-kube` Fixtures

`gdt` Fixtures are objects that help set up and tear down a testing
environment. The `gdt-kube` library has some utility fixtures to make testing
with Kubernetes easier.

### `KindFixture`

The `KindFixture` eases integration of `gdt-kube` tests with the KinD local
Kubernetes development system.

To use it, import the `gdt-kube/fixtures/kind` package:

```go
import (
    "github.com/jaypipes/gdt"
    gdtkube "github.com/jaypipes/gdt-kube"
    gdtkind "github.com/jaypipes/gdt-kube/fixtures/kind"
)
```

and then register the fixture with your `gdt` `Context`, like so:

```go
func TestExample(t *testing.T) {
    s, err := gdt.From("path/to/test.yaml")
    if err != nil {
        t.Fatalf("failed to load tests: %s", err)
    }

    ctx := context.Background()
    ctx = gdt.RegisterFixture(ctx, "kind", gdtkind.New())
    err = s.Run(ctx, t)
    if err != nil {
        t.Fatalf("failed to run tests: %s", err)
    }
}
```

In your test file, you would list the "kind" fixture in the `requires` list:

```yaml
name: example-using-kind
require:
 - kind
tests:
 - kube.get: pods/nginx
```

## Contributing and acknowledgements

`gdt` was inspired by [Gabbi](https://github.com/cdent/gabbi), the excellent
Python declarative testing framework. `gdt` tries to bring the same clear,
concise test definitions to the world of Go functional testing.

Contributions to `gdt-kube` are welcomed! Feel free to open a Github issue or
submit a pull request.
