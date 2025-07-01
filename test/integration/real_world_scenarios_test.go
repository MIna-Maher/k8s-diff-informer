package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/MIna-Maher/k8s-diff-informer/internal/slack"
	"github.com/MIna-Maher/k8s-diff-informer/pkg/diff"
	"github.com/MIna-Maher/k8s-diff-informer/test/fixtures"
	"github.com/MIna-Maher/k8s-diff-informer/test/helpers"
	"github.com/MIna-Maher/k8s-diff-informer/test/mocks"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// RealWorldScenariosTestSuite tests real-world scenarios
type RealWorldScenariosTestSuite struct {
	suite.Suite
	mockSlack *mocks.MockSlackClient
}

func (suite *RealWorldScenariosTestSuite) SetupTest() {
	suite.mockSlack = mocks.NewMockSlackClient("https://hooks.slack.com/test", "test-cluster")
}

// TestDeploymentRollingUpdate simulates a rolling update scenario
func (suite *RealWorldScenariosTestSuite) TestDeploymentRollingUpdate() {
	// Original deployment
	originalDeployment := fixtures.DeploymentFixtures.BasicDeployment
	originalUnstructured, err := fixtures.ToUnstructured(originalDeployment)
	suite.Require().NoError(err)

	// Updated deployment with new image
	updatedDeployment := fixtures.DeploymentFixtures.UpdatedDeployment
	updatedUnstructured, err := fixtures.ToUnstructured(updatedDeployment)
	suite.Require().NoError(err)

	// Extract specs and compute diff
	originalSpec, _, _ := unstructured.NestedMap(originalUnstructured.Object, "spec")
	updatedSpec, _, _ := unstructured.NestedMap(updatedUnstructured.Object, "spec")

	fieldsToRemove := []string{
		"metadata.creationTimestamp",
		"metadata.uid",
		"metadata.resourceVersion",
		"status",
	}

	diff.RemoveFields(originalSpec, fieldsToRemove)
	diff.RemoveFields(updatedSpec, fieldsToRemove)

	diffResult := diff.ComputeDiff(originalSpec, updatedSpec)

	// Verify the diff contains the image change
	assert.Contains(suite.T(), diffResult, "nginx:1.14")
	assert.Contains(suite.T(), diffResult, "nginx:1.15")

	// Simulate sending notification
	message := &slack.Message{
		Title:  "Deployment Updated",
		Text:   diffResult,
		Footer: "Cluster: test-cluster",
	}

	err = suite.mockSlack.SendMessage(message)
	assert.NoError(suite.T(), err)

	// Verify notification was sent
	messages := suite.mockSlack.GetSentMessages()
	assert.Len(suite.T(), messages, 1)
	assert.Contains(suite.T(), messages[0].Text, "nginx:1.15")
}

// TestMultipleResourceUpdates simulates multiple resources being updated simultaneously
func (suite *RealWorldScenariosTestSuite) TestMultipleResourceUpdates() {
	testCases := []struct {
		name         string
		originalObj  runtime.Object
		updatedObj   runtime.Object
		expectedDiff bool
	}{
		{
			name:         "Pod image update",
			originalObj:  fixtures.PodFixtures.BasicPod,
			updatedObj:   helpers.UpdatePodImage(fixtures.PodFixtures.BasicPod, "nginx:1.15"),
			expectedDiff: true,
		},
		{
			name:         "Deployment scaling",
			originalObj:  fixtures.DeploymentFixtures.BasicDeployment,
			updatedObj:   fixtures.DeploymentFixtures.ScaledDeployment,
			expectedDiff: true,
		},
		{
			name:         "Service port change",
			originalObj:  fixtures.ServiceFixtures.ClusterIPService,
			updatedObj:   fixtures.ServiceFixtures.NodePortService,
			expectedDiff: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			originalUnstructured, err := fixtures.ToUnstructured(tc.originalObj)
			suite.Require().NoError(err)

			updatedUnstructured, err := fixtures.ToUnstructured(tc.updatedObj)
			suite.Require().NoError(err)

			// Extract specs
			originalSpec, _, _ := unstructured.NestedMap(originalUnstructured.Object, "spec")
			updatedSpec, _, _ := unstructured.NestedMap(updatedUnstructured.Object, "spec")

			// Compute diff
			diffResult := diff.ComputeDiff(originalSpec, updatedSpec)

			if tc.expectedDiff {
				assert.NotEmpty(suite.T(), diffResult, "Expected diff for %s", tc.name)
			} else {
				assert.Empty(suite.T(), diffResult, "Expected no diff for %s", tc.name)
			}

			// Send notification if there's a diff
			if diffResult != "" {
				message := &slack.Message{
					Title:  tc.name,
					Text:   diffResult,
					Footer: "Cluster: test-cluster",
				}

				err = suite.mockSlack.SendMessage(message)
				assert.NoError(suite.T(), err)
			}
		})
	}

	// Verify all expected notifications were sent
	messages := suite.mockSlack.GetSentMessages()
	assert.Len(suite.T(), messages, 3) // All test cases should generate diffs
}

// TestNamespaceFiltering tests namespace filtering logic
func (suite *RealWorldScenariosTestSuite) TestNamespaceFiltering() {
	watchedNamespaces := []string{"default", "production", "staging"}

	testCases := []struct {
		namespace   string
		shouldWatch bool
	}{
		{"default", true},
		{"production", true},
		{"staging", true},
		{"development", false},
		{"kube-system", false},
		{"", false},
	}

	for _, tc := range testCases {
		suite.Run(tc.namespace, func() {
			// Simulate the namespace filtering logic
			shouldWatch := false
			for _, ns := range watchedNamespaces {
				if ns == tc.namespace {
					shouldWatch = true
					break
				}
			}

			assert.Equal(suite.T(), tc.shouldWatch, shouldWatch,
				"Namespace %s watch expectation failed", tc.namespace)
		})
	}
}

// TestHighVolumeEvents simulates high volume of events
func (suite *RealWorldScenariosTestSuite) TestHighVolumeEvents() {
	const numEvents = 100
	const concurrentWorkers = 10

	// Channel to collect results
	results := make(chan error, numEvents)

	// Generate high volume of events
	for i := 0; i < concurrentWorkers; i++ {
		go func(workerID int) {
			for j := 0; j < numEvents/concurrentWorkers; j++ {
				message := &slack.Message{
					Title:  "High Volume Event",
					Text:   "Event from worker",
					Footer: "test-cluster",
				}

				err := suite.mockSlack.SendMessage(message)
				results <- err
			}
		}(i)
	}

	// Collect all results
	for i := 0; i < numEvents; i++ {
		select {
		case err := <-results:
			assert.NoError(suite.T(), err)
		case <-time.After(5 * time.Second):
			suite.T().Fatal("Timeout waiting for high volume events")
		}
	}

	// Verify all messages were sent
	messages := suite.mockSlack.GetSentMessages()
	assert.Len(suite.T(), messages, numEvents)
}

// TestErrorHandling tests various error scenarios
func (suite *RealWorldScenariosTestSuite) TestErrorHandling() {
	testCases := []struct {
		name          string
		setupError    func()
		expectedError bool
	}{
		{
			name: "Slack service unavailable",
			setupError: func() {
				suite.mockSlack.SetShouldFail(true)
				suite.mockSlack.SetFailureMessage("service unavailable")
			},
			expectedError: true,
		},
		{
			name: "Network timeout",
			setupError: func() {
				suite.mockSlack.SetShouldFail(true)
				suite.mockSlack.SetFailureMessage("network timeout")
			},
			expectedError: true,
		},
		{
			name: "Invalid webhook URL",
			setupError: func() {
				suite.mockSlack.SetShouldFail(true)
				suite.mockSlack.SetFailureMessage("invalid webhook URL")
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup error condition
			tc.setupError()

			message := &slack.Message{
				Title:  "Test Error Handling",
				Text:   "This should fail",
				Footer: "test-cluster",
			}

			err := suite.mockSlack.SendMessage(message)

			if tc.expectedError {
				assert.Error(suite.T(), err)
			} else {
				assert.NoError(suite.T(), err)
			}

			// Reset for next test
			suite.mockSlack.SetShouldFail(false)
		})
	}
}

// TestConfigMapUpdates tests ConfigMap update scenarios
func (suite *RealWorldScenariosTestSuite) TestConfigMapUpdates() {
	originalConfigMap := fixtures.ConfigMapFixtures.BasicConfigMap

	// Create updated ConfigMap
	updatedConfigMap := originalConfigMap.DeepCopy()
	updatedConfigMap.Data["config.yaml"] = `
server:
  port: 9090
  host: 0.0.0.0
database:
  url: postgres://localhost:5432/newdb
`
	updatedConfigMap.ResourceVersion = "4001"

	// Convert to unstructured
	originalUnstructured, err := fixtures.ToUnstructured(originalConfigMap)
	suite.Require().NoError(err)

	updatedUnstructured, err := fixtures.ToUnstructured(updatedConfigMap)
	suite.Require().NoError(err)

	// Extract data and compute diff
	originalData, _, _ := unstructured.NestedMap(originalUnstructured.Object, "data")
	updatedData, _, _ := unstructured.NestedMap(updatedUnstructured.Object, "data")

	diffResult := diff.ComputeDiff(originalData, updatedData)

	// Verify the diff contains the configuration changes
	assert.NotEmpty(suite.T(), diffResult)
	assert.Contains(suite.T(), diffResult, "port")

	// Send notification
	message := &slack.Message{
		Title:  "ConfigMap Updated",
		Text:   diffResult,
		Footer: "Cluster: test-cluster",
	}

	err = suite.mockSlack.SendMessage(message)
	assert.NoError(suite.T(), err)

	// Verify notification
	messages := suite.mockSlack.GetSentMessages()
	assert.Len(suite.T(), messages, 1)
	assert.Contains(suite.T(), messages[0].Title, "ConfigMap")
}

func TestRealWorldScenariosTestSuite(t *testing.T) {
	suite.Run(t, new(RealWorldScenariosTestSuite))
}
