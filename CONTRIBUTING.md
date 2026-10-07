# Contributing

Thanks for helping improve k8s-diff-informer. Bug reports, focused fixes, documentation improvements, and feature proposals are welcome.

## Before you start

- Search existing issues and pull requests to avoid duplicating work.
- For a significant behavior change, open an issue first to discuss the approach.
- Never include real Slack webhook URLs, kubeconfig files, tokens, private cluster data, or unredacted Kubernetes objects in an issue or pull request.
- Keep changes focused and update documentation when behavior or configuration changes.

## Development setup

The project requires Go 1.23 or newer. Helm 3 is required for chart checks. A Kubernetes cluster and Slack webhook are not required to run the local checks.

```sh
git clone https://github.com/MIna-Maher/k8s-diff-informer.git
cd k8s-diff-informer
gofmt -w $(find cmd internal pkg test -name '*.go')
go build ./...
go test ./...
python3 -B -m unittest discover -s test/helm
helm lint --strict ./deployment/helm --set slack.existingSecret=example
```

`go test ./...` currently includes a known failing assertion in `test/integration/real_world_scenarios_test.go` (`TestConfigMapUpdates`), documented in the README. Run focused packages while working on unrelated changes, and include the relevant checks in your pull request. Do not use a real Slack webhook or cluster credentials for tests.

## Repository layout

- `cmd/k8s-diff-informer/`: application entry point.
- `internal/`: configuration, Kubernetes informers, queue, Slack client, HTTP endpoints, and metrics.
- `pkg/diff/`: resource normalization and diff logic.
- `deployment/helm/`: chart, templates, schema, and installation guide.
- `test/`: Go integration tests, fixtures, mocks, and Helm checks.
- `docs/` and `site/`: project and published documentation.

## Pull request workflow

1. Create a branch from `main` with a concise name, such as `fix/readiness-probe`.
2. Make the smallest change that addresses the issue. Add or update tests for changed behavior.
3. Run `gofmt` on changed Go files, the relevant Go tests, and Helm checks when chart files change.
4. Update user-facing documentation when commands, defaults, configuration, or behavior change.
5. Open a pull request against `main`. Describe the problem, the change, and how you validated it. Link related issues and include screenshots only when they help explain a user-facing change.
6. Respond to review feedback and keep the pull request focused.

Maintainers may request changes or close contributions that are out of scope, unsafe, or inconsistent with the project's [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to follow it.
