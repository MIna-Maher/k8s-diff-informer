# Testing Guide for k8s-diff-informer

This directory contains comprehensive tests for the k8s-diff-informer project.

## Test Structure

```
test/
├── integration/           # Integration tests
│   ├── simple_integration_test.go
│   ├── real_world_scenarios_test.go
│   └── performance_test.go
├── helpers/              # Test helper functions
│   └── test_helpers.go
├── mocks/               # Mock implementations
│   └── mock_clients.go
├── fixtures/            # Test data fixtures
│   └── kubernetes_objects.go
├── reports/             # Generated test reports
├── run_tests.sh         # Simple test runner
└── README.md           # This file
```

## Running Tests

### Quick Start
```bash
# Run all tests with simple runner
./test/run_tests.sh

# Run comprehensive test suite
./scripts/run-tests.sh

# Run only unit tests
go test -v ./pkg/... ./internal/config/... ./internal/slack/...

# Run only integration tests  
go test -v ./test/integration/...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o test/reports/coverage.html
```

### Detailed Testing
```bash
# Run end-to-end tests (requires Docker and kind)
./scripts/test-e2e.sh

# Run specific test suites
go test -v ./pkg/diff/...
go test -v ./internal/config/...
go test -v ./test/integration/...

# Run specific test functions
go test -v ./test/integration/ -run TestBasicDiffComputation
go test -v ./test/integration/ -run TestSlackIntegration
```

## Test Categories

### Unit Tests
- **Location**: `pkg/diff/`, `internal/config/`, `internal/slack/`
- **Purpose**: Test individual components in isolation
- **Coverage**: Diff computation, configuration loading, Slack client

### Integration Tests
- **Location**: `test/integration/`
- **Purpose**: Test component interactions and workflows
- **Coverage**: End-to-end scenarios, error handling, performance

### Performance Tests
- **Purpose**: Performance testing and benchmarking
- **Usage**: `go test -bench=. -benchmem ./...`

## Test Infrastructure

### Fixtures (`test/fixtures/`)
Pre-built Kubernetes objects for testing:
- **Pods**: Basic, multi-container, with volumes
- **Deployments**: Basic, scaled, updated
- **Services**: ClusterIP, NodePort, LoadBalancer
- **ConfigMaps**: Basic configuration and environment variables

### Mocks (`test/mocks/`)
Mock implementations for external dependencies:
- **MockSlackClient**: Thread-safe Slack webhook simulation
- **Configurable failure modes**: Network errors, timeouts, etc.

### Helpers (`test/helpers/`)
Utility functions for test setup:
- Object creation and modification functions
- Type conversion utilities
- Assertion helpers

## Example Test Usage

```go
func TestPodUpdate(t *testing.T) {
    // Use fixtures
    originalPod := fixtures.PodFixtures.BasicPod
    updatedPod := helpers.UpdatePodImage(originalPod, "nginx:1.15")
    
    // Convert to unstructured
    oldUnstructured := helpers.ConvertToUnstructured(t, originalPod)
    newUnstructured := helpers.ConvertToUnstructured(t, updatedPod)
    
    // Test diff computation
    diffResult := diff.ComputeDiff(oldUnstructured.Object, newUnstructured.Object)
    
    // Use mock for Slack testing
    mockSlack := mocks.NewMockSlackClient("test-url", "test-cluster")
    err := mockSlack.SendMessage(&slack.Message{
        Title: "Pod Updated",
        Text:  diffResult,
    })
    
    assert.NoError(t, err)
    assert.Len(t, mockSlack.GetSentMessages(), 1)
}
```

## Configuration

### Environment Variables for Testing
```bash
export SLACK_WEBHOOK_URL=https://hooks.slack.com/test
export CLUSTER_NAME=test-cluster  
export WATCHED_RESOURCE_NAMES=pods,deployments,services
export WATCHED_NAMESPACES=default,test-namespace
```

### Test Runners
- **`test/run_tests.sh`**: Simple, fast test runner (recommended for development)
- **`scripts/run-tests.sh`**: Comprehensive test suite with reporting
- **`scripts/test-e2e.sh`**: End-to-end tests with real Kubernetes cluster

## Specific Test Files

### Integration Tests
- **`simple_integration_test.go`**: Core integration tests without flag conflicts
- **`real_world_scenarios_test.go`**: Complex scenarios like rolling updates
- **`performance_test.go`**: Performance and load testing

### Unit Tests
- **`pkg/diff/diff_test.go`**: Diff computation and field removal
- **`internal/config/config_test.go`**: Configuration loading and parsing
- **`internal/slack/slack_enhanced_test.go`**: Enhanced Slack client tests

## Performance Testing

```bash
# Run benchmarks
go test -bench=. -benchmem ./test/integration/

# Run performance tests
go test -v ./test/integration/ -run TestPerformance

# Run memory tests
go test -v ./test/integration/ -run TestMemoryUsage
```

## Coverage Requirements

- **Target**: 80% minimum coverage
- **Report**: Generated in `test/reports/coverage.html`
- **Command**:
  ```bash
  go test -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out -o test/reports/coverage.html
  ```

## Troubleshooting

### Common Issues

1. **Flag redefinition errors**
    - Use `test/run_tests.sh` which avoids flag conflicts
    - Run tests individually: `go test -v ./test/integration/ -run TestBasicDiffComputation`

2. **Race condition errors**
    - Run with `-race` flag to detect issues: `go test -race ./...`
    - Check mock thread safety in concurrent tests

3. **Missing environment variables**
    - Set required env vars: `SLACK_WEBHOOK_URL`, `CLUSTER_NAME`
    - Use the test runner scripts which set these automatically

### Debug Mode
```bash
# Run tests with verbose output
go test -v -race ./...

# Run specific test with debug info
go test -v -run TestSpecificFunction ./path/to/package

# Run with coverage and race detection
go test -v -race -coverprofile=coverage.out ./...
```

## Test Coverage Areas

✅ **Diff Computation**: Value changes, additions, removals, nested structures  
✅ **Field Filtering**: Metadata cleanup, nested field removal  
✅ **Slack Integration**: Message sending, error handling, concurrent operations  
✅ **Kubernetes Objects**: Pod/Deployment lifecycle simulation  
✅ **Configuration**: Environment variable handling and validation  
✅ **Error Scenarios**: Network failures, service unavailability  
✅ **Performance**: High-volume events, memory usage, benchmarks  
✅ **Race Conditions**: Thread safety validation

## Contributing

When adding new tests:

1. **Follow existing patterns**: Use fixtures, helpers, and mocks consistently
2. **Include edge cases**: Test both positive and negative scenarios
3. **Add benchmarks**: For performance-critical code paths
4. **Update documentation**: Add new test categories to this README
5. **Test isolation**: Ensure tests don't interfere with each other

For more details, see the main project README and contribution guidelines.