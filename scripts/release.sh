#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
Usage: ./scripts/release.sh vMAJOR.MINOR.PATCH

Run the stable-release preflight on a clean, up-to-date main branch. This
script validates and builds a local snapshot; it does not create or push tags
and it does not publish a GitHub release or container image.
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -ne 1 || ! "$1" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  usage >&2
  exit 2
fi
tag="$1"
version="${tag#v}"

if [[ "$(git branch --show-current)" != "main" ]]; then
  echo "Run the release preflight from the main branch." >&2
  exit 1
fi
if [[ -n "$(git status --porcelain)" ]]; then
  echo "Commit or remove working-tree changes before running the release preflight." >&2
  exit 1
fi

git fetch origin main --quiet
if [[ "$(git rev-parse HEAD)" != "$(git rev-parse origin/main)" ]]; then
  echo "Local main must match origin/main before release preparation." >&2
  exit 1
fi
if git show-ref --verify --quiet "refs/tags/$tag" || git ls-remote --exit-code --tags origin "refs/tags/$tag" >/dev/null 2>&1; then
  echo "Tag $tag already exists locally or on origin." >&2
  exit 1
fi

chart_version="$(sed -n 's/^version: //p' deployment/helm/Chart.yaml)"
app_version="$(sed -n 's/^appVersion: "\(.*\)"/\1/p' deployment/helm/Chart.yaml)"
if [[ "$chart_version" != "$version" || "$app_version" != "$version" ]]; then
  echo "Update Chart.yaml version and appVersion to $version before release." >&2
  exit 1
fi

if ! command -v helm >/dev/null 2>&1 || ! command -v docker >/dev/null 2>&1; then
  echo "Helm 3 and Docker with Buildx are required for the release preflight." >&2
  exit 1
fi
goreleaser_version="$(goreleaser --version | awk '/^GitVersion:/ {print $2}' | sed 's/^v//')"
if [[ -z "$goreleaser_version" ]]; then
  echo "GoReleaser is required (version 2.12 or newer)." >&2
  exit 1
fi
IFS=. read -r goreleaser_major goreleaser_minor _ <<<"$goreleaser_version"
if (( goreleaser_major < 2 || (goreleaser_major == 2 && goreleaser_minor < 12) )); then
  echo "GoReleaser 2.12 or newer is required; found $goreleaser_version." >&2
  exit 1
fi

echo "Running Go release checks..."
go build ./...
go test -race ./...
go vet ./...

echo "Running Helm release checks..."
python3 -B -m unittest discover -s test/helm -v
helm lint --strict ./deployment/helm --set slack.existingSecret=preflight-placeholder
helm template release ./deployment/helm --set slack.existingSecret=preflight-placeholder >/dev/null

echo "Validating GoReleaser configuration and building a local snapshot..."
goreleaser check
goreleaser release --snapshot --clean

echo
echo "Preflight passed for $tag. No tag, GitHub release, or image was published."
echo "After reviewing CHANGELOG.md and pushing this commit to main, create and push the stable tag:"
echo "  git tag -a $tag -m 'k8s-diff-informer $tag'"
echo "  git push origin $tag"
