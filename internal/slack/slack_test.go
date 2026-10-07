package slack

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlackClientCreation(t *testing.T) {
	testCases := []struct {
		name        string
		webhookURL  string
		clusterName string
	}{
		{
			name:        "valid webhook and cluster",
			webhookURL:  "https://example.invalid/slack-webhook",
			clusterName: "test-cluster",
		},
		{
			name:        "empty webhook URL",
			webhookURL:  "",
			clusterName: "test-cluster",
		},
		{
			name:        "empty cluster name",
			webhookURL:  "https://example.invalid/slack-webhook",
			clusterName: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewSlackClient(tc.webhookURL, tc.clusterName)
			assert.NotNil(t, client)
			assert.Equal(t, tc.webhookURL, client.WebhookURL)
			assert.Equal(t, tc.clusterName, client.ClusterName)
		})
	}
}

func TestMessageSending(t *testing.T) {
	testCases := []struct {
		name           string
		webhookURL     string
		message        *Message
		serverResponse int
		expectError    bool
	}{
		{
			name:       "successful message sent",
			webhookURL: "placeholder", // will be replaced with test server URL
			message: &Message{
				Title:  "Test Title",
				Text:   "Test Text",
				Footer: "Test Footer",
			},
			serverResponse: http.StatusOK,
			expectError:    false,
		},
		{
			name:       "server error response",
			webhookURL: "placeholder",
			message: &Message{
				Title:  "Test Title",
				Text:   "Test Text",
				Footer: "Test Footer",
			},
			serverResponse: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:       "bad request response",
			webhookURL: "placeholder",
			message: &Message{
				Title:  "Test Title",
				Text:   "Test Text",
				Footer: "Test Footer",
			},
			serverResponse: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Validate request method
				assert.Equal(t, http.MethodPost, r.Method)

				// Validate content type
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(tc.serverResponse)
			}))
			defer server.Close()

			// Create client with test server URL
			client := NewSlackClient(server.URL, "test-cluster")
			err := client.SendMessage(tc.message)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmptyWebhookURL(t *testing.T) {
	client := NewSlackClient("", "test-cluster")
	message := &Message{
		Title:  "Test Title",
		Text:   "Test Text",
		Footer: "Test Footer",
	}

	// Should not return error but should not attempt to send
	err := client.SendMessage(message)
	assert.NoError(t, err)
}

func TestMessageFields(t *testing.T) {
	message := &Message{
		Title:  "Resource Updated",
		Text:   "```diff text here```",
		Footer: "Cluster: test-cluster",
	}

	assert.Equal(t, "Resource Updated", message.Title)
	assert.Equal(t, "```diff text here```", message.Text)
	assert.Equal(t, "Cluster: test-cluster", message.Footer)
}

func TestConcurrentMessageSending(t *testing.T) {
	// Create a test server that simulates some latency
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Simulate network latency
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSlackClient(server.URL, "test-cluster-from-unit-test")

	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	// Send messages concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			message := &Message{
				Title:  "Concurrent Test" + strconv.Itoa(id),
				Text:   "Test message",
				Footer: "test-cluster",
			}
			results <- client.SendMessage(message)
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-results:
			assert.NoError(t, err)
			//Timeout mechanism
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent requests")
		}
	}
}

//func TestMessageSendingSlackPostWebhook(t *testing.T) {
//	// You'll need a real Slack webhook URL for testing
//	// Consider using environment variable or test configuration
//	webhookURL := "https://example.invalid/slack-webhook"
//	if webhookURL == "" {
//		t.Skip("SLACK_WEBHOOK_URL_TEST environment variable not set, skipping unit test")
//	}
//
//	testCases := []struct {
//		name        string
//		message     *Message
//		expectError bool
//	}{
//		{
//			name: "successful message sent",
//			message: &Message{
//				Title:  "Test Title",
//				Text:   "Test Text",
//				Footer: "Test Footer",
//			},
//			expectError: false,
//		},
//		{
//			name: "empty message",
//			message: &Message{
//				Title:  "",
//				Text:   "",
//				Footer: "",
//			},
//			expectError: false, // Slack might accept empty messages
//		},
//		{
//			name: "message with special characters",
//			message: &Message{
//				Title:  "Test with 特殊字符 and émojis 🚀",
//				Text:   "Testing special chars: @channel #general",
//				Footer: "Footer with <https://example.com|link>",
//			},
//			expectError: false,
//		},
//	}
//
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			// Convert your Message to slack.WebhookMessage
//			webhookMsg := &slack.WebhookMessage{
//				Text: tc.message.Text,
//				Attachments: []slack.Attachment{
//					{
//						Title:  tc.message.Title,
//						Text:   tc.message.Text,
//						Footer: tc.message.Footer,
//						Color:  "good", // or whatever color logic you use
//					},
//				},
//			}
//
//			// Use Slack.PostWebhook directly
//			err := slack.PostWebhook(webhookURL, webhookMsg)
//
//			if tc.expectError {
//				assert.Error(t, err)
//			} else {
//				assert.NoError(t, err)
//			}
//		})
//	}
//}
