#!/bin/bash

# test/run_tests.sh
set -e

echo "🧪 Running k8s-diff-informer tests..."

## Set up environment variables
#export SLACK_WEBHOOK_URL="https://example.invalid/slack-webhook"
#export CLUSTER_NAME="test-cluster"
#export WATCHED_RESOURCE_NAMES="pods,deployments,services"
#export WATCHED_NAMESPACES="default,test-namespace"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[TEST] $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Clean test cache
print_status "Cleaning test cache..."
go clean -testcache

# Test 1: Unit tests for diff package
print_status "Running diff package unit tests..."
if go test -v ./pkg/diff/...; then
    print_success "Diff package tests passed"
else
    print_error "Diff package tests failed"
    exit 1
fi

# Test 2: Unit tests for slack package
print_status "Running slack package tests..."
if go test -v ./internal/slack/...; then
    print_success "Slack package tests passed"
else
    print_error "Slack package tests failed"
    exit 1
fi

# Test 4: Simple integration tests
print_status "Running simple integration tests..."
if go test -v ./test/integration/ -run "^TestBasic|^TestSlack|^TestPod|^TestDeployment|^TestConcurrent|^TestField|^TestEnvironment|^TestEndToEnd"; then
    print_success "Simple integration tests passed"
else
    print_error "Simple integration tests failed"
    exit 1
fi

# Test 5: Build test
print_status "Testing build..."
if go build -o /tmp/k8s-diff-informer ./cmd/k8s-diff-informer/; then
    print_success "Build test passed"
    rm -f /tmp/k8s-diff-informer
else
    print_error "Build test failed"
    exit 1
fi

# Test 6: Race condition test
print_status "Running race condition tests..."
if go test -race -short ./pkg/... ./internal/slack/...; then
    print_success "Race condition tests passed"
else
    print_error "Race condition tests failed"
    exit 1
fi

print_success "All tests completed successfully! 🎉"

echo ""
echo "Test Summary:"
echo "✅ Diff package unit tests"
echo "✅ Slack package tests"
echo "✅ Simple integration tests"
echo "✅ Build test"
echo "✅ Race condition tests"
echo ""
echo "To run individual test files:"
echo "  go test -v ./pkg/diff/"
echo "  go test -v ./internal/slack/"
echo "  go test -v ./test/integration/ -run TestBasicDiffComputation"
echo "  go test -v ./test/integration/ -run TestSlackIntegration"