package mocks

import (
	"fmt"
	"sync"

	"github.com/MIna-Maher/k8s-diff-informer/internal/slack"
)

// MockSlackClient implements a mock Slack client for testing
type MockSlackClient struct {
	mu             sync.RWMutex
	SentMessages   []*slack.Message
	ShouldFailSend bool
	FailureMessage string
	WebhookURL     string
	ClusterName    string
	CallCount      int
}

// NewMockSlackClient creates a new mock Slack client
func NewMockSlackClient(webhookURL, clusterName string) *MockSlackClient {
	return &MockSlackClient{
		WebhookURL:     webhookURL,
		ClusterName:    clusterName,
		SentMessages:   make([]*slack.Message, 0),
		FailureMessage: "mock send error",
	}
}

// SendMessage simulates sending a message to Slack
func (m *MockSlackClient) SendMessage(msg *slack.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CallCount++

	if m.ShouldFailSend {
		return fmt.Errorf(m.FailureMessage)
	}

	// Deep copy the message to avoid race conditions
	msgCopy := &slack.Message{
		Title:  msg.Title,
		Text:   msg.Text,
		Footer: msg.Footer,
	}

	m.SentMessages = append(m.SentMessages, msgCopy)
	return nil
}

// GetSentMessages returns all sent messages
func (m *MockSlackClient) GetSentMessages() []*slack.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	messages := make([]*slack.Message, len(m.SentMessages))
	copy(messages, m.SentMessages)
	return messages
}

// ClearMessages clears all sent messages
func (m *MockSlackClient) ClearMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SentMessages = make([]*slack.Message, 0)
	m.CallCount = 0
}

// GetCallCount returns the number of times SendMessage was called
func (m *MockSlackClient) GetCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.CallCount
}

// SetShouldFail configures whether SendMessage should fail
func (m *MockSlackClient) SetShouldFail(shouldFail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ShouldFailSend = shouldFail
}

// SetFailureMessage sets the error message returned when failing
func (m *MockSlackClient) SetFailureMessage(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FailureMessage = message
}

// GetLastMessage returns the last sent message, or nil if none
func (m *MockSlackClient) GetLastMessage() *slack.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.SentMessages) == 0 {
		return nil
	}

	return m.SentMessages[len(m.SentMessages)-1]
}

// HasMessageWithTitle checks if any sent message has the given title
func (m *MockSlackClient) HasMessageWithTitle(title string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, msg := range m.SentMessages {
		if msg.Title == title {
			return true
		}
	}
	return false
}

// HasMessageWithText checks if any sent message contains the given text
func (m *MockSlackClient) HasMessageWithText(text string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, msg := range m.SentMessages {
		if msg.Text == text {
			return true
		}
	}
	return false
}

// CountMessagesWithTitle counts messages with a specific title
func (m *MockSlackClient) CountMessagesWithTitle(title string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, msg := range m.SentMessages {
		if msg.Title == title {
			count++
		}
	}
	return count
}
