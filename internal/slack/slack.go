package slack

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/slack-go/slack"
	"k8s.io/klog/v2"
)

// Client represents a Slack client for sending notifications
type Client struct {
	WebhookURL  string
	ClusterName string
}

// Message represents a Slack message
type Message struct {
	Title  string
	Text   string
	Footer string
}

// NewSlackClient creates a new Slack client
func NewSlackClient(webhookURL, clusterName string) *Client {
	return &Client{
		WebhookURL:  webhookURL,
		ClusterName: clusterName,
	}
}

// SendMessage sends a message to the Slack webhook
func (c *Client) SendMessage(msg *Message) error {
	if c.WebhookURL == "" {
		klog.Warning("Slack webhook URL is not set, skipping notification")
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
	if err != nil {
		klog.Errorf("Failed to send message to Slack: %v", err)
		return fmt.Errorf("failed to send message to Slack: %v", err)
	}

	klog.V(2).Infof("Sent message to Slack: %s", msg.Title)
	return nil
}
