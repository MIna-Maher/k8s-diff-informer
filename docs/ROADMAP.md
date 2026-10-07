# Open source release roadmap

This roadmap prepares k8s-diff-informer for its first stable public release, planned as <code>v1.0.0</code>. It follows useful project practices from [ai9s](https://github.com/AymanZahran/ai9s): usable installation docs, clear contribution paths, automated checks, and consistent releases.

Keep the project name and existing MIT license. Prioritize Helm installation and container distribution. The planned release artifacts are a GitHub Release tagged <code>v1.0.0</code>, container images tagged <code>1.0.0</code>, and a Helm chart with <code>version: 1.0.0</code> and <code>appVersion: "1.0.0"</code>.

## 1. Clean up the repository and exposed credentials

- [x] Add <code>.gitignore</code> and <code>.dockerignore</code> for IDE files, coverage reports, binaries, and local configuration.
- [x] Remove embedded Slack webhook values from current source and test fixtures; use clearly fake placeholders.
- [ ] Revoke any genuine Slack webhook found in the repository. Historical copies remain in Git.
- [x] Inspect reachable local Git history for exposed credentials and document the result.
- [x] Preserve unrelated local changes during cleanup.

See the [credential cleanup findings](CREDENTIAL-CLEANUP.md). Webhook revocation is an owner action; replacing current values does not invalidate historical copies.

## 2. Make installation reliable

- [x] Make default Helm values independent of private image pull secrets and Prometheus Operator CRDs.
- [x] Validate Slack settings, queue values, ports, watched resource lists, and replica count.
- [x] Default to one replica and avoid overlapping pods during upgrades.
- [x] Replace wildcard resource permissions with explicit read-only default rules and document extensions.
- [x] Parse <code>--kubeconfig</code> and reject unsupported synchronous queue mode.
- [x] Tie readiness to informer cache synchronization and return HTTP errors when the metrics server cannot bind.

See the [Helm installation guide](../deployment/helm/README.md). A live-cluster installation still needs an image built from this source or a published image.

## 3. Write the public documentation

- [x] Replace the planning README with a user guide and preserve this roadmap.
- [x] Add a sanitized notification example and an architecture diagram.
- [x] Provide Helm quick-start instructions that explain image availability.
- [x] Document application environment variables and Helm configuration.
- [x] Explain limitations, monitoring, troubleshooting, and sensitive-resource handling.
- [x] Correct the Helm and testing guide filenames, links, and outdated instructions.
- [x] Add a GitHub Pages site source and deployment workflow.

The Pages site source is under <code>site/</code>. Publishing it requires enabling GitHub Actions as the Pages source in repository settings. See the [GitHub Pages documentation](https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages) and the [site home page](../site/index.md).

## 4. Add contributor and community files

- [ ] Add <code>CONTRIBUTING.md</code> with development commands, repository structure, contribution workflow, and review expectations.
- [ ] Add <code>CODE_OF_CONDUCT.md</code> and <code>SECURITY.md</code> with a private vulnerability reporting channel.
- [ ] Add bug report and feature request issue forms.
- [ ] Add a pull request template with validation and documentation prompts.
- [ ] Keep the MIT license and attribution.

## 5. Add checks for contributions

- [ ] Add formatting, build, Go tests with race detection, and lint checks.
- [ ] Add dependency vulnerability scanning.
- [ ] Add Helm linting and rendering checks, including optional monitoring disabled.
- [ ] Ensure pull request checks need no production credentials, cluster, or real Slack messages.
- [ ] Configure dependency updates for Go modules and GitHub Actions.
- [x] Add focused Go and Helm regression coverage for phase 2.

## 6. Prepare and document the first stable release

- [ ] Reconcile the Docker publishing workflow, release script, and GoReleaser.
- [ ] Resolve GoReleaser's reference to the missing <code>Dockerfile.goreleaser</code>.
- [ ] Publish stable releases from semantic version tags such as <code>v1.0.0</code>. The existing main-branch workflow publishes the <code>beta</code> tag.
- [ ] Configure GoReleaser for a stable GitHub Release with <code>prerelease: false</code>.
- [ ] Publish Linux <code>amd64</code> and <code>arm64</code> images under <code>1.0.0</code> and update <code>latest</code> only on stable releases.
- [ ] Package a Helm chart with matching <code>1.0.0</code> chart and application versions.
- [ ] Add build version information, release notes, and a maintained changelog.
- [ ] Document and validate the maintainer release process.

## Completion criteria

- A new user can follow the README, install on a clean cluster, and receive an expected Slack notification.
- A contributor can run the documented checks without private credentials.
- Release artifacts can be built and validated through the documented process.
- The <code>v1.0.0</code> GitHub Release, container images, Helm chart, and installation instructions use consistent versions.

Complete these phases in order as reviewable changes. A custom domain, additional package managers, and custom branding can follow the first usable release.
