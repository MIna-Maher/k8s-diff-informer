#!/usr/bin/env bash

set -euo pipefail

rm -rf .release-artifacts
mkdir -p .release-artifacts
helm package ./deployment/helm --destination .release-artifacts
