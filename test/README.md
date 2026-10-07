# Tests

This directory contains integration tests, fixtures, helpers, and mocks for k8s-diff-informer.

## Run the Go suite

From the repository root:

~~~sh
go test -race ./...
~~~

Run one package or test by name:

~~~sh
go test ./pkg/diff/...
go test ./test/integration/...
go test ./test/integration -run '^TestBasicDiffComputation$'
~~~

Run the race detector on the concurrent queue and notification code:

~~~sh
go test -race ./internal/queue ./internal/slack
~~~

## Validate Helm changes

Helm chart checks run locally without a Kubernetes cluster. Install Helm 3 first.

~~~sh
python3 -B -m unittest discover -s test/helm
helm lint --strict ./deployment/helm --set slack.existingSecret=example
~~~

The Helm tests render the chart and check its default resources, Secret handling, optional Prometheus Operator objects, custom HTTP ports, and rejection of invalid values.

## Test layout

- <code>integration/</code>: simulated resource events and notification workflows.
- <code>fixtures/</code>: sample Kubernetes API objects used by integration tests.
- <code>helpers/</code>: integration-test helpers.
- <code>mocks/</code>: mock notification clients.
- <code>helm/</code>: Helm render and validation checks.

Most tests use local fixtures, mocks, or HTTP test servers. They do not send notifications to a real Slack workspace.
