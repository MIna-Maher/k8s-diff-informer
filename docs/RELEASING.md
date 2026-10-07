# Stable release process

The repository is prepared for its first stable release, planned as `v1.0.0`.
Stable releases are created from plain semantic-version tags. Pushing `v1.0.0`
starts the GitHub Actions workflow that builds the release, uploads its artifacts,
and publishes Linux `amd64` and `arm64` images to GHCR. The `latest` image tag is
updated only by this stable-release workflow. The existing `beta` image remains
published from `main`.

The workflow does not run for prerelease tags such as `v1.0.0-beta.1`.

## One-time GitHub setup

The stable workflow uses the repository's `GITHUB_TOKEN`; no personal token is
needed. It requires `contents: write` to create the GitHub Release and
`packages: write` to push images.

The GHCR package is public, so images can be pulled anonymously. Confirm the
package remains public after the first stable image push. Publishing an image
does not change an existing package's visibility.

## Prepare a release

1. Merge the release changes to `main` and wait for the CI checks to pass.
2. Update `CHANGELOG.md`: move the release notes out of `Unreleased`, use a
   dated heading such as `## [1.0.0] - YYYY-MM-DD`, and review the user-facing
   changes. GoReleaser also generates release notes from commits since the
   previous tag; review those in the GitHub Release after publishing.
3. Set `version` and `appVersion` in `deployment/helm/Chart.yaml` to the exact
   version without the `v` prefix. The chart is packaged as a `.tgz` and attached
   to the GitHub Release.
4. Confirm the README and installation guide accurately state whether the
   image is available publicly. For `v1.0.0`, the matching chart values already
   use image tag `1.0.0`.
5. On a clean, up-to-date `main` checkout with Go, Helm 3, Docker Buildx, and
   GoReleaser 2.12 or newer installed, run:

   ```sh
   ./scripts/release.sh v1.0.0
   ```

   The preflight checks the branch and worktree, confirms the tag is unused,
   verifies the chart versions, runs Go build/race tests/vet and Helm checks,
   validates `.goreleaser.yml`, and builds a local snapshot. It does not create
   a tag, GitHub Release, or published image.
6. Commit and push any final changelog, chart, or documentation edits. Rerun the
   preflight after those edits so it checks the exact commit that will be tagged.
7. Review the preflight output, then create and push the annotated stable tag:

   ```sh
   git tag -a v1.0.0 -m "k8s-diff-informer v1.0.0"
   git push origin v1.0.0
   ```

Pushing the tag is the publication step. The release workflow validates that
the tag, chart version, and chart `appVersion` match, reruns Go and Helm checks,
then uses GoReleaser to publish. Do not reuse or move a published tag; make a
new patch version to correct a release.

## Published artifacts

For `v1.0.0`, the workflow publishes:

- A non-prerelease GitHub Release with Linux `amd64` and `arm64` archives,
  `checksums.txt`, the packaged Helm chart, and the changelog.
- `ghcr.io/mina-maher/k8s-diff-informer:1.0.0` as a multi-platform image.
- `ghcr.io/mina-maher/k8s-diff-informer:latest`, updated only by stable tags.
- Executable version metadata shown by `k8s-diff-informer --version`.

The chart can be installed from the downloaded release archive with:

```sh
helm upgrade --install diff-monitor ./k8s-diff-informer-1.0.0.tgz \
  --namespace monitoring --create-namespace \
  --set slack.existingSecret=informer-slack --wait --timeout 5m
```

Create the Kubernetes Secret first, following the [Helm installation guide](../deployment/helm/README.md).

## After publishing

- [x] Confirm the GitHub Release is not marked as a prerelease and includes all
  expected assets.
- [x] Inspect the `1.0.0` and `latest` GHCR manifests and verify both architectures.
- [ ] From a clean client, pull the image anonymously and install the published
  chart archive.
- [ ] Confirm `/ready` succeeds after informer synchronization and verify a test
  resource change produces the expected Slack notification.
- [x] Update the README and GitHub Pages to document the published stable image.
- Keep `CHANGELOG.md` entries for published versions; start the next changes in
  the `Unreleased` section.
