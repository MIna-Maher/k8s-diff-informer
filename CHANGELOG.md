# Changelog

## [Unreleased]

## [1.0.0] - 2026-10-07

### Added
- Stable GoReleaser workflow for GitHub releases and Linux multi-architecture container images.
- Version, commit, build date, and builder metadata in the executable, available through `--version`.
- Packaged Helm chart attached to the stable GitHub Release.
- Documented maintainer release checklist and local release preflight.

### Changed
- Helm chart version now matches the planned stable application version (`1.0.0`).
- Beta publishing remains tied to `main`; stable publishing only runs for plain semantic-version tags.

### Fixed
- GoReleaser now uses the checked-in Dockerfile and a config supported by current GoReleaser v2.
- The normal Docker build uses the Go toolchain required by `go.mod`.

## [v1.0.0-beta.1] - 2025-07-14 [BETA]

### Notes
- Initial beta release for testing; not recommended for production.
- Feedback and issues: https://github.com/MIna-Maher/k8s-diff-informer/issues
