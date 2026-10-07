package integration

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MIna-Maher/k8s-diff-informer/internal/slack"
	"github.com/MIna-Maher/k8s-diff-informer/pkg/diff"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// MockSlackClient for testing
type MockSlackClient struct {
	mu             sync.RWMutex
	SentMessages   []*slack.Message
	ShouldFailSend bool
	WebhookURL     string
	ClusterName    string
}

// NewMockSlackClient creates a new instance of MockSlackClient
func NewMockSlackClient(webhookURL, clusterName string) *MockSlackClient {
	return &MockSlackClient{
		WebhookURL:   webhookURL,
		ClusterName:  clusterName,
		SentMessages: make([]*slack.Message, 0),
	}
}

// SendMessage simulates sending a message to Slack
func (m *MockSlackClient) SendMessage(msg *slack.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ShouldFailSend {
		return fmt.Errorf("mock send error..")
	}

	// Deep copy to avoid race conditions
	msgCopy := &slack.Message{
		Title:  msg.Title,
		Text:   msg.Text,
		Footer: msg.Footer,
	}
	m.SentMessages = append(m.SentMessages, msgCopy)
	return nil
}

// GetSentMessages returns a copy of the sent messages
func (m *MockSlackClient) GetSentMessages() []*slack.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	messages := make([]*slack.Message, len(m.SentMessages))
	copy(messages, m.SentMessages)
	return messages
}

// ClearMessages clears the sent messages
func (m *MockSlackClient) ClearMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SentMessages = make([]*slack.Message, 0)
}

// Helper functions
func createTestPod(name, namespace, image string) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			ResourceVersion: "1000",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: image,
				},
			},
		},
	}
}

func createTestDeployment(name, namespace, image string, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			ResourceVersion: "2000",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: image,
						},
					},
				},
			},
		},
	}
}

// toUnstructured converts a runtime.Object to an unstructured.Unstructured
// transforms any Kubernetes object into the generic unstructured format while ensuring test failures are caught early and clearly reported.
func toUnstructured(t *testing.T, obj runtime.Object) *unstructured.Unstructured {
	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj) //Converts typed object to map[string]interface{}
	require.NoError(t, err)                                                          //stop execution on failure
	// Output: map[string]interface{}
	return &unstructured.Unstructured{Object: unstructuredMap}
}

// Tests

func TestBasicDiffComputation(t *testing.T) {
	testCases := []struct {
		name     string
		oldMap   map[string]interface{}
		newMap   map[string]interface{}
		expected string
	}{
		{
			name: "simple value change",
			oldMap: map[string]interface{}{
				"replicas": 2,
				"image":    "nginx:1.14",
			},
			newMap: map[string]interface{}{
				"replicas": 3,
				"image":    "nginx:1.15",
			},
			expected: "- image: nginx:1.14\n+ image: nginx:1.15\n- replicas: 2\n+ replicas: 3\n",
		},
		{
			name: "no changes",
			oldMap: map[string]interface{}{
				"replicas": 2,
				"image":    "nginx:1.14",
			},
			newMap: map[string]interface{}{
				"replicas": 2,
				"image":    "nginx:1.14",
			},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := diff.ComputeDiff(tc.oldMap, tc.newMap)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSlackIntegration(t *testing.T) {
	mockSlack := NewMockSlackClient("https://example.invalid/slack-webhook", "test-cluster")

	message := &slack.Message{
		Title:  "Test Resource Added",
		Text:   "```Resource Added: test-pod```",
		Footer: "Cluster: test-cluster",
	}

	err := mockSlack.SendMessage(message)
	require.NoError(t, err)

	messages := mockSlack.GetSentMessages()
	require.Len(t, messages, 1)
	assert.Equal(t, "Test Resource Added", messages[0].Title)
	assert.Equal(t, "```Resource Added: test-pod```", messages[0].Text)
}

func TestPodLifecycleSimulation(t *testing.T) {
	// Create a test pod
	pod := createTestPod("test-pod", "default", "nginx:1.14")
	unstructuredPod := toUnstructured(t, pod)

	// Simulate update - change image
	updatedPod := pod.DeepCopy()
	updatedPod.Spec.Containers[0].Image = "nginx:1.15"
	updatedPod.ResourceVersion = "1001"
	updatedUnstructured := toUnstructured(t, updatedPod)

	// Extract specs and compute diff
	originalSpec, _, _ := unstructured.NestedMap(unstructuredPod.Object, "spec")
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
	assert.Contains(t, diffResult, "nginx:1.14")
	assert.Contains(t, diffResult, "nginx:1.15")
}

func TestDeploymentScaling(t *testing.T) {
	// Create original deployment
	deployment := createTestDeployment("test-deployment", "default", "myapp:v1", 2)
	originalUnstructured := toUnstructured(t, deployment)

	// Scale up deployment
	scaledDeployment := deployment.DeepCopy()
	scaledDeployment.Spec.Replicas = int32Ptr(5)
	scaledDeployment.ResourceVersion = "2001"
	scaledUnstructured := toUnstructured(t, scaledDeployment)

	// Extract specs and compute diff
	originalSpec, _, _ := unstructured.NestedMap(originalUnstructured.Object, "spec")
	scaledSpec, _, _ := unstructured.NestedMap(scaledUnstructured.Object, "spec")

	diffResult := diff.ComputeDiff(originalSpec, scaledSpec)
	assert.Contains(t, diffResult, "replicas: 2")
	assert.Contains(t, diffResult, "replicas: 5")
}

func TestConcurrentSlackOperations(t *testing.T) {
	mockSlack := NewMockSlackClient("https://example.invalid/slack-webhook", "test-cluster")

	const numGoroutines = 5
	const messagesPerGoroutine = 3

	var wg sync.WaitGroup
	results := make(chan error, numGoroutines*messagesPerGoroutine)

	// Start multiple goroutines sending messages concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				message := &slack.Message{
					Title:  fmt.Sprintf("Message from goroutine %d, iteration %d", id, j),
					Text:   fmt.Sprintf("Test content %d-%d", id, j),
					Footer: "test-cluster",
				}

				err := mockSlack.SendMessage(message)
				results <- err
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Check all results
	errorCount := 0
	for err := range results {
		if err != nil {
			errorCount++
		}
	}

	assert.Equal(t, 0, errorCount, "Expected no errors in concurrent operations")

	// Verify total number of messages
	messages := mockSlack.GetSentMessages()
	expectedCount := numGoroutines * messagesPerGoroutine
	actualCount := len(messages)

	assert.Equal(t, expectedCount, actualCount,
		"Expected %d messages, got %d", expectedCount, actualCount)
}

func TestSlackFailureHandling(t *testing.T) {
	mockSlack := NewMockSlackClient("https://example.invalid/slack-webhook", "test-cluster")
	mockSlack.ShouldFailSend = true

	message := &slack.Message{
		Title:  "Test Message",
		Text:   "Test content",
		Footer: "Test cluster",
	}

	err := mockSlack.SendMessage(message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock send error")

	// Verify no messages were sent
	messages := mockSlack.GetSentMessages()
	assert.Len(t, messages, 0)
}

func TestFieldRemoval(t *testing.T) {
	testMap := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":              "test-pod",
			"namespace":         "default",
			"creationTimestamp": "2023-01-01T00:00:00Z",
			"uid":               "12345",
			"resourceVersion":   "1000",
		},
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{
					"name":  "nginx",
					"image": "nginx:1.14",
				},
			},
		},
		"status": map[string]interface{}{
			"phase": "Running",
		},
	}

	fieldsToRemove := []string{
		"metadata.creationTimestamp",
		"metadata.uid",
		"metadata.resourceVersion",
		"status",
	}

	result := diff.RemoveFields(testMap, fieldsToRemove)

	// Check that removed fields are not present
	metadata := result["metadata"].(map[string]interface{})
	assert.NotContains(t, metadata, "creationTimestamp")
	assert.NotContains(t, metadata, "uid")
	assert.NotContains(t, metadata, "resourceVersion")
	assert.NotContains(t, result, "status")

	// Check that other fields are still present
	assert.Contains(t, metadata, "name")
	assert.Contains(t, metadata, "namespace")
	assert.Contains(t, result, "spec")
}

func TestEnvironmentVariableConfiguration(t *testing.T) {
	// Store original values
	originalSlackURL := os.Getenv("SLACK_WEBHOOK_URL")
	originalClusterName := os.Getenv("CLUSTER_NAME")

	// Test valid configuration
	t.Run("valid configuration", func(t *testing.T) {
		os.Setenv("SLACK_WEBHOOK_URL", "https://example.invalid/slack-webhook")
		os.Setenv("CLUSTER_NAME", "test-cluster")

		slackURL := os.Getenv("SLACK_WEBHOOK_URL")
		clusterName := os.Getenv("CLUSTER_NAME")

		assert.Equal(t, "https://example.invalid/slack-webhook", slackURL)
		assert.Equal(t, "test-cluster", clusterName)
	})

	// Restore original values
	if originalSlackURL == "" {
		os.Unsetenv("SLACK_WEBHOOK_URL")
	} else {
		os.Setenv("SLACK_WEBHOOK_URL", originalSlackURL)
	}

	if originalClusterName == "" {
		os.Unsetenv("CLUSTER_NAME")
	} else {
		os.Setenv("CLUSTER_NAME", originalClusterName)
	}
}

func TestEndToEndWorkflow(t *testing.T) {
	// Setup
	mockSlack := NewMockSlackClient("https://example.invalid/slack-webhook", "integration-test-cluster")

	// Create original pod
	originalPod := createTestPod("test-pod", "default", "nginx:1.14")
	originalUnstructured := toUnstructured(t, originalPod)

	// Create updated pod
	updatedPod := originalPod.DeepCopy()
	updatedPod.Spec.Containers[0].Image = "nginx:1.15"
	updatedPod.ResourceVersion = "1001"
	updatedUnstructured := toUnstructured(t, updatedPod)

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

	// Verify diff was computed
	assert.NotEmpty(t, diffResult)
	assert.Contains(t, diffResult, "nginx:1.14")
	assert.Contains(t, diffResult, "nginx:1.15")

	// Send notification
	message := &slack.Message{
		Title:  "Pod Updated",
		Text:   fmt.Sprintf("```%s```", diffResult),
		Footer: "Cluster: integration-test-cluster",
	}

	err := mockSlack.SendMessage(message)
	require.NoError(t, err)

	// Verify notification was sent
	messages := mockSlack.GetSentMessages()
	require.Len(t, messages, 1)
	assert.Equal(t, "Pod Updated", messages[0].Title)
	assert.Contains(t, messages[0].Text, "nginx:1.15")
	assert.Contains(t, messages[0].Footer, "integration-test-cluster")
}

// Helper function
func int32Ptr(i int32) *int32 {
	return &i
}
