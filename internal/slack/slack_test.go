package slack

import (
	"testing"
)

func TestNewSlackClient(t *testing.T) {
	testCases := []struct {
		name        string
		webhookURL  string
		clusterName string
		expectError bool
	}{
		{
			name:        "Empty webhook URL",
			webhookURL:  "",
			expectError: true,
			clusterName: "test-cluster",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Slackclient := NewSlackClient(tc.webhookURL, tc.clusterName)

			if Slackclient.WebhookURL != tc.webhookURL {
				t.Errorf("Expected WebhookURL %s, got %s", tc.webhookURL, Slackclient.WebhookURL)
			}
		})
	}
}

//func TestSendMessage(t *testing.T) {
//	testCases := []struct {
//		name           string
//		webhookURL     string
//		message        *Message
//		serverResponse int
//		expectError    bool
//	}{
//		{
//			name:       "Successful message send",
//			webhookURL: "http://example.com/webhook",
//			message: &Message{
//				Title:  "Test Title",
//				Text:   "Test Text",
//				Footer: "Test Footer",
//			},
//			serverResponse: http.StatusOK,
//			expectError:    false,
//		},
//		{
//			name:       "Server error",
//			webhookURL: "http://example.com/webhook",
//			message: &Message{
//				Title:  "Test Title",
//				Text:   "Test Text",
//				Footer: "Test Footer",
//			},
//			serverResponse: http.StatusInternalServerError,
//			expectError:    true,
//		},
//		{
//			name:       "Empty webhook URL",
//			webhookURL: "",
//			message: &Message{
//				Title:  "Test Title",
//				Text:   "Test Text",
//				Footer: "Test Footer",
//			},
//			serverResponse: http.StatusOK,
//			expectError:    false, // should not return error, just skip sending
//		},
//	}
//
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			// Create a test server that returns the desired response
//			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//				// Check request method and content type
//				if r.Method != http.MethodPost {
//					t.Errorf("Expected POST request, got %s", r.Method)
//				}
//
//				if r.Header.Get("Content-Type") != "application/json" {
//					t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
//				}
//
//				// Return the configured status code
//				w.WriteHeader(tc.serverResponse)
//			}))
//			defer server.Close()
//
//			// Override the webhook URL if it's not empty
//			actualWebhookURL := tc.webhookURL
//			if tc.webhookURL != "" {
//				actualWebhookURL = server.URL
//			}
//
//			// Create client and send message
//			client := NewClient(actualWebhookURL, "test-cluster")
//			err := client.SendMessage(tc.message)
//
//			// Check if error status matches expectation
//			if (err != nil) != tc.expectError {
//				t.Errorf("Expected error: %v, got: %v", tc.expectError, err != nil)
//			}
//		})
//	}
//}

// Mock implementation for testing - this would normally be in a separate mock file or test helper
type MockSlackClient struct {
	WebhookURL     string
	ClusterName    string
	SentMessages   []*Message
	ShouldFailSend bool
}

func NewMockSlackClient(webhookURL, clusterName string) *MockSlackClient {
	return &MockSlackClient{
		WebhookURL:     webhookURL,
		ClusterName:    clusterName,
		SentMessages:   make([]*Message, 0),
		ShouldFailSend: false,
	}
}

func (m *MockSlackClient) SendMessage(msg *Message) error {
	if m.ShouldFailSend {
		return &mockSendError{}
	}

	m.SentMessages = append(m.SentMessages, msg)
	return nil
}

type mockSendError struct{}

func (e *mockSendError) Error() string {
	return "mock send error"
}
