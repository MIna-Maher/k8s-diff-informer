package slack

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/slack-go/slack"
	"k8s.io/klog/v2"
)

// MetricsRecorder interface for recording metrics
type MetricsRecorder interface {
	RecordSlackNotification(status, cluster string, duration time.Duration)
	RecordSlackError(errorType, cluster string)
}

// Client represents a Slack client for sending notifications
type Client struct {
	WebhookURL      string
	ClusterName     string
	metricsRecorder MetricsRecorder
}

// Message represents a Slack message
type Message struct {
	Title  string
	Text   string
	Footer string
}

// NewSlackClient creates a new Slack client without metrics
func NewSlackClient(webhookURL, clusterName string) *Client {
	return &Client{
		WebhookURL:      webhookURL,
		ClusterName:     clusterName,
		metricsRecorder: nil,
	}
}

// NewSlackClientWithMetrics creates a new Slack client with metrics recording
func NewSlackClientWithMetrics(webhookURL, clusterName string, metricsRecorder MetricsRecorder) *Client {
	return &Client{
		WebhookURL:      webhookURL,
		ClusterName:     clusterName,
		metricsRecorder: metricsRecorder,
	}
}

// SendMessage sends a message to the Slack webhook
func (c *Client) SendMessage(msg *Message) error {
	start := time.Now()

	if c.WebhookURL == "" {
		klog.Warning("Slack webhook URL is not set, skipping notification")

		// Record metrics if available
		if c.metricsRecorder != nil {
			c.metricsRecorder.RecordSlackError("webhook_url_missing", c.ClusterName)
		}
		return nil
	}

	attachment := slack.Attachment{
		Text:        msg.Text,
		Pretext:     msg.Title,
		Footer:      msg.Footer,
		ServiceName: "K8s-Diff-Informer",
		MarkdownIn:  []string{"text", "pretext"},
		Color:       "#4599DF",
		Ts:          json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
		FooterIcon:  "https://raw.githubusercontent.com/kubernetes/kubernetes/master/logo/logo.png",
	}

	err := slack.PostWebhook(c.WebhookURL, &slack.WebhookMessage{
		IconEmoji:   ":kubernetes:",
		Attachments: []slack.Attachment{attachment},
	})

	duration := time.Since(start)

	// Record metrics if available
	if c.metricsRecorder != nil {
		if err != nil {
			c.metricsRecorder.RecordSlackNotification("error", c.ClusterName, duration)
			c.metricsRecorder.RecordSlackError("webhook_request_failed", c.ClusterName)
		} else {
			c.metricsRecorder.RecordSlackNotification("success", c.ClusterName, duration)
		}
	}

	if err != nil {
		klog.Errorf("Failed to send message to Slack: %v", err)
		return fmt.Errorf("failed to send message to Slack: %v", err)
	}

	klog.V(2).Infof("Sent message to Slack: %s (duration: %v)", msg.Title, duration)
	return nil
}
