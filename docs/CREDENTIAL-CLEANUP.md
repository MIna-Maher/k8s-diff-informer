# Credential cleanup before publication

## Findings from 2026-10-06

A pattern scan of the working tree and 219 Git objects reachable from local
refs found genuine-looking Slack webhook URLs in current files and historical
blobs. Values are intentionally omitted from this report.

Affected paths in reachable history:

- `cmd/k8s-diff-informer/main.go`
- `internal/slack/slack.go` (historical content)
- `internal/slack/slack_test.go`
- `test/integration/integration_test.go`
- `test/run_tests.sh`

Current source comments and fixture values now use
`https://example.invalid/slack-webhook`. The two Slack tests that previously
sent to embedded webhooks now use their existing local HTTP test servers.

The scan also checked patterns for Slack tokens, GitHub tokens, AWS access key
IDs, and private key headers. No matches for those categories were found.
This was a bounded pattern scan, not proof that every possible credential is
absent. Remote-only refs, forks, external caches, and unreachable Git objects
were not inspected. The webhook URLs were not contacted or tested for validity.

## Required owner action

- [ ] Revoke every genuine exposed webhook in the owning Slack app/workspace,
      including older webhook values retained only in history.
- [ ] Create replacement webhooks if needed and update deployed Kubernetes
      Secrets or local environment configuration outside version control.
- [ ] Confirm revocation before publishing the release.

Adding ignore rules or replacing current values does not invalidate credentials
or remove old commits. Revocation is required even if history is later rewritten.

## History cleanup decision

The reachable history contains webhook values, so deleting them from current
files is insufficient to remove them from the repository. History cleanup is
recommended before publication if the intention is to remove those values from
reachable commits. Coordinate that operation with the owner and collaborators:
it changes commit IDs and may require force-pushing affected branches and tags.

No history rewrite, force-push, or credential revocation was performed during
phase 1. Once revocation is confirmed, agree on affected refs and collaborator
recovery steps before rewriting history. A rewrite cannot remove copies held in
other clones, forks, or caches.

## Validation

- The working tree rescan found no remaining genuine-looking Slack webhook URLs.
- Ignore checks confirmed local configuration, IDE files, coverage output, and
  build artifacts are excluded from Git, while source files and `.env.example`
  remain trackable.
- Existing unrelated modified and untracked files were verified unchanged using
  SHA-256 comparisons.
- `git diff --check` passed.
- The `internal/slack` tests passed, including the tests now using local servers.
- The five integration tests changed by this cleanup passed when run separately.
- The full `test/integration` package failed in the unchanged
  `TestRealWorldScenariosTestSuite/TestConfigMapUpdates` test: the computed diff
  did not contain the expected `port` text. That failure remains unresolved.
