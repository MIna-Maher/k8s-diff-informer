package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/MIna-Maher/k8s-diff-informer/internal/slack"
	"k8s.io/klog/v2"
)

// Handler manages task processing
type Handler struct {
	slackClient *slack.Client
	metrics     MetricsRecorder
}

// NewHandler creates a new task handler
func NewHandler(slackClient *slack.Client, metrics MetricsRecorder) *Handler {
	return &Handler{
		slackClient: slackClient,
		metrics:     metrics,
	}
}

// Handle routes tasks to appropriate handlers based on task type
func (h *Handler) Handle(ctx context.Context, task *Task) error {
	klog.V(3).Infof("Handling task %s of type %s", task.ID, task.Type)

	switch task.Type {
	case TypeSlackNotification:
		return h.handleSlackNotification(ctx, task)
	case TypeBatchedNotifications:
		return h.handleBatchedNotifications(ctx, task)
	case TypeResourceDiffCompute:
		return h.handleResourceDiff(ctx, task)
	default:
		return fmt.Errorf("unknown task type: %s", task.Type)
	}
}

// handleSlackNotification processes a Slack notification task
func (h *Handler) handleSlackNotification(ctx context.Context, task *Task) error {
	payload, ok := task.Payload.(*SlackNotificationPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for SlackNotification, expected *SlackNotificationPayload")
	}

	klog.V(2).Infof("Sending Slack notification: %s (resource: %s/%s)",
		payload.Title, payload.Namespace, payload.ResourceName)

	msg := &slack.Message{
		Title:  payload.Title,
		Text:   payload.Text,
		Footer: payload.Footer,
	}

	// Check for context cancellation before sending
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled before sending notification")
	default:
	}

	start := time.Now()
	err := h.slackClient.SendMessage(msg)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Failed to send Slack notification for %s/%s: %v",
			payload.Namespace, payload.ResourceName, err)
		return fmt.Errorf("slack notification failed: %w", err)
	}

	klog.V(2).Infof("Slack notification sent successfully in %v", duration)
	return nil
}

// handleBatchedNotifications processes multiple notifications as a batch
func (h *Handler) handleBatchedNotifications(ctx context.Context, task *Task) error {
	payload, ok := task.Payload.(*BatchedNotificationsPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for BatchedNotifications")
	}

	klog.V(2).Infof("Processing batch of %d notifications for cluster %s",
		len(payload.Notifications), payload.ClusterName)

	// Combine notifications into a single message
	var combinedText string
	resourceCount := len(payload.Notifications)

	for i, notif := range payload.Notifications {
		combinedText += fmt.Sprintf("%d. %s\n%s\n", i+1, notif.Title, notif.Text)
		if i < resourceCount-1 {
			combinedText += "\n---\n\n"
		}
	}

	msg := &slack.Message{
		Title: fmt.Sprintf("*Batched Notifications: %d events in %v*",
			resourceCount, payload.TimeWindow),
		Text:   combinedText,
		Footer: fmt.Sprintf("Cluster: %s", payload.ClusterName),
	}

	err := h.slackClient.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send batched notifications: %w", err)
	}

	klog.V(2).Infof("Batched notification sent successfully (%d events)", resourceCount)
	return nil
}

// handleResourceDiff processes resource diff computation
func (h *Handler) handleResourceDiff(ctx context.Context, task *Task) error {
	payload, ok := task.Payload.(*ResourceDiffPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for ResourceDiff")
	}

	klog.V(3).Infof("Computing diff for resource %s/%s",
		payload.Namespace, payload.ResourceName)

	// This is a placeholder - in practice, you might want to:
	// 1. Compute the diff
	// 2. Store it somewhere
	// 3. Or trigger another action based on the diff

	// For now, just log that we would compute it
	klog.V(3).Infof("Diff computation completed for %s/%s",
		payload.Namespace, payload.ResourceName)

	return nil
}
