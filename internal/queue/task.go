package queue

import (
	"time"
)

// Task types
const (
	TypeSlackNotification    = "slack:notification"
	TypeBatchedNotifications = "slack:batch"
	TypeResourceDiffCompute  = "resource:diff"
)

// SlackNotificationPayload represents data for a Slack notification task
type SlackNotificationPayload struct {
	Title        string
	Text         string
	Footer       string
	ClusterName  string
	ResourceType string
	ResourceName string
	Namespace    string
	EventType    string // add, update, delete
	Timestamp    time.Time
}

// BatchedNotificationsPayload represents multiple notifications to be batched
type BatchedNotificationsPayload struct {
	Notifications []*SlackNotificationPayload
	ClusterName   string
	TimeWindow    time.Duration
}

// ResourceDiffPayload represents data for computing resource diffs
type ResourceDiffPayload struct {
	ResourceType   string
	ResourceName   string
	Namespace      string
	OldSpec        map[string]interface{}
	NewSpec        map[string]interface{}
	FieldsToRemove []string
}

// NewSlackNotificationTask creates a new Slack notification task
func NewSlackNotificationTask(payload *SlackNotificationPayload) *Task {
	return &Task{
		Type:       TypeSlackNotification,
		Payload:    payload,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}
}

// NewBatchedNotificationsTask creates a new batched notifications task
func NewBatchedNotificationsTask(payload *BatchedNotificationsPayload) *Task {
	return &Task{
		Type:       TypeBatchedNotifications,
		Payload:    payload,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}
}

// NewResourceDiffTask creates a new resource diff computation task
func NewResourceDiffTask(payload *ResourceDiffPayload) *Task {
	return &Task{
		Type:       TypeResourceDiffCompute,
		Payload:    payload,
		MaxRetries: 2, // Fewer retries for compute tasks
		CreatedAt:  time.Now(),
	}
}
