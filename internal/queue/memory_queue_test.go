package queue

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockMetricsRecorder for testing
type MockMetricsRecorder struct {
	mu              sync.Mutex
	enqueuedCount   int
	processedCount  int
	failedCount     int
	retriedCount    int
	droppedCount    int
	queueSize       int
	workers         int
	processingTimes []time.Duration
}

func (m *MockMetricsRecorder) RecordTaskEnqueued(taskType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enqueuedCount++
}

func (m *MockMetricsRecorder) RecordTaskProcessed(taskType string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processedCount++
	m.processingTimes = append(m.processingTimes, duration)
}

func (m *MockMetricsRecorder) RecordTaskFailed(taskType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failedCount++
}

func (m *MockMetricsRecorder) RecordTaskRetried(taskType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.retriedCount++
}

func (m *MockMetricsRecorder) RecordTaskDropped(taskType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.droppedCount++
}

func (m *MockMetricsRecorder) UpdateQueueSize(size int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queueSize = size
}

func (m *MockMetricsRecorder) UpdateQueueWorkers(workers int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workers = workers
}

func (m *MockMetricsRecorder) GetCounts() (enqueued, processed, failed, retried, dropped int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.enqueuedCount, m.processedCount, m.failedCount, m.retriedCount, m.droppedCount
}

func TestNewMemoryQueue(t *testing.T) {
	metrics := &MockMetricsRecorder{}
	queue := NewMemoryQueue(5, 100, metrics)

	assert.NotNil(t, queue)
	assert.Equal(t, 5, queue.workers)
	assert.Equal(t, 100, queue.queueSize)
	assert.False(t, queue.isRunning)
}

func TestQueueStartAndStop(t *testing.T) {
	metrics := &MockMetricsRecorder{}
	queue := NewMemoryQueue(2, 10, metrics)

	// Simple handler that does nothing
	handler := func(ctx context.Context, task *Task) error {
		return nil
	}

	// Start the queue
	queue.Start(handler)
	assert.True(t, queue.isRunning)

	// Wait a bit for workers to start
	time.Sleep(100 * time.Millisecond)

	// Stop the queue
	queue.Stop()
	assert.False(t, queue.isRunning)
}

func TestEnqueueTask(t *testing.T) {
	metrics := &MockMetricsRecorder{}
	queue := NewMemoryQueue(2, 10, metrics)

	processedTasks := make(chan string, 10)
	handler := func(ctx context.Context, task *Task) error {
		processedTasks <- task.ID
		return nil
	}

	queue.Start(handler)
	defer queue.Stop()

	// Enqueue a task
	task := &Task{
		ID:         "test-task-1",
		Type:       TypeSlackNotification,
		Payload:    &SlackNotificationPayload{Title: "Test"},
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	err := queue.Enqueue(task)
	assert.NoError(t, err)

	// Wait for task to be processed
	select {
	case processedID := <-processedTasks:
		assert.Equal(t, "test-task-1", processedID)
	case <-time.After(2 * time.Second):
		t.Fatal("Task was not processed in time")
	}

	// Check metrics
	enqueued, processed, _, _, _ := metrics.GetCounts()
	assert.Equal(t, 1, enqueued)
	assert.Equal(t, 1, processed)
}

func TestQueueFull(t *testing.T) {
	metrics := &MockMetricsRecorder{}
	queue := NewMemoryQueue(1, 2, metrics)

	started := make(chan struct{})
	release := make(chan struct{})
	var startedOnce sync.Once
	var releaseOnce sync.Once
	handler := func(ctx context.Context, task *Task) error {
		startedOnce.Do(func() { close(started) })
		<-release
		return nil
	}

	queue.Start(handler)
	defer queue.Stop()
	defer func() { releaseOnce.Do(func() { close(release) }) }()

	assert.NoError(t, queue.Enqueue(&Task{ID: "processing-task", Type: TypeSlackNotification}))
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not start processing the first task")
	}

	// Keep the worker occupied while filling both buffered slots.
	assert.NoError(t, queue.Enqueue(&Task{ID: "queued-task-1", Type: TypeSlackNotification}))
	assert.NoError(t, queue.Enqueue(&Task{ID: "queued-task-2", Type: TypeSlackNotification}))

	err := queue.Enqueue(&Task{ID: "dropped-task", Type: TypeSlackNotification})
	assert.EqualError(t, err, "queue is full")

	_, _, _, _, dropped := metrics.GetCounts()
	assert.Equal(t, 1, dropped)
	releaseOnce.Do(func() { close(release) })
}

func TestRetryLogic(t *testing.T) {
	metrics := &MockMetricsRecorder{}
	queue := NewMemoryQueue(1, 10, metrics)

	attemptCount := 0
	var mu sync.Mutex

	handler := func(ctx context.Context, task *Task) error {
		mu.Lock()
		attemptCount++
		currentAttempt := attemptCount
		mu.Unlock()

		// Fail first 2 attempts, succeed on 3rd
		if currentAttempt < 3 {
			return errors.New("simulated failure")
		}
		return nil
	}

	queue.Start(handler)
	defer queue.Stop()

	task := &Task{
		ID:         "retry-task",
		Type:       TypeSlackNotification,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	err := queue.Enqueue(task)
	assert.NoError(t, err)

	// Wait for retries to complete
	time.Sleep(5 * time.Second)

	mu.Lock()
	finalAttempts := attemptCount
	mu.Unlock()

	// Should have been attempted 3 times (1 initial + 2 retries)
	assert.Equal(t, 3, finalAttempts)

	// Check metrics
	_, processed, _, retried, _ := metrics.GetCounts()
	assert.Equal(t, 1, processed)
	assert.Equal(t, 2, retried) // 2 retries after initial failure
}

func TestConcurrentEnqueue(t *testing.T) {
	metrics := &MockMetricsRecorder{}
	queue := NewMemoryQueue(5, 100, metrics)

	processedCount := 0
	var mu sync.Mutex

	handler := func(ctx context.Context, task *Task) error {
		mu.Lock()
		processedCount++
		mu.Unlock()
		return nil
	}

	queue.Start(handler)
	defer queue.Stop()

	// Enqueue tasks concurrently
	const numTasks = 50
	var wg sync.WaitGroup

	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			task := &Task{
				ID:         "concurrent-task-" + string(rune(id)),
				Type:       TypeSlackNotification,
				MaxRetries: 1,
			}
			_ = queue.Enqueue(task)
		}(i)
	}

	wg.Wait()

	// Wait for all tasks to be processed
	time.Sleep(2 * time.Second)

	mu.Lock()
	final := processedCount
	mu.Unlock()

	// All tasks should be processed
	assert.Equal(t, numTasks, final)
}

func TestQueueSize(t *testing.T) {
	metrics := &MockMetricsRecorder{}
	queue := NewMemoryQueue(1, 10, metrics)

	handler := func(ctx context.Context, task *Task) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	queue.Start(handler)
	defer queue.Stop()

	// Enqueue several tasks
	for i := 0; i < 5; i++ {
		task := &Task{
			ID:   "task-" + string(rune(i)),
			Type: TypeSlackNotification,
		}
		_ = queue.Enqueue(task)
	}

	// Check queue size
	size := queue.Size()
	assert.GreaterOrEqual(t, size, 0)
	assert.LessOrEqual(t, size, 5)
}

func TestQueueStats(t *testing.T) {
	metrics := &MockMetricsRecorder{}
	queue := NewMemoryQueue(3, 50, metrics)

	handler := func(ctx context.Context, task *Task) error {
		return nil
	}

	queue.Start(handler)

	stats := queue.Stats()
	assert.Equal(t, 3, stats["workers"])
	assert.Equal(t, 50, stats["queue_size"])
	assert.True(t, stats["is_running"].(bool))

	queue.Stop()

	stats = queue.Stats()
	assert.False(t, stats["is_running"].(bool))
}

func TestGracefulShutdown(t *testing.T) {
	metrics := &MockMetricsRecorder{}
	queue := NewMemoryQueue(2, 10, metrics)

	processedTasks := make(chan string, 10)
	handler := func(ctx context.Context, task *Task) error {
		time.Sleep(200 * time.Millisecond) // Simulate work
		processedTasks <- task.ID
		return nil
	}

	queue.Start(handler)

	// Enqueue tasks
	for i := 0; i < 5; i++ {
		task := &Task{
			ID:   "shutdown-task-" + string(rune(i)),
			Type: TypeSlackNotification,
		}
		_ = queue.Enqueue(task)
	}

	// Give workers time to pick up tasks
	time.Sleep(100 * time.Millisecond)

	// Stop should wait for in-flight tasks
	queue.Stop()

	// Count processed tasks
	processedCount := len(processedTasks)
	close(processedTasks)

	// At least some tasks should have been processed
	assert.Greater(t, processedCount, 0)
}

func TestTaskAutoRetry(t *testing.T) {
	metrics := &MockMetricsRecorder{}
	queue := NewMemoryQueue(1, 10, metrics)

	attempts := make(map[string]int)
	var mu sync.Mutex

	handler := func(ctx context.Context, task *Task) error {
		mu.Lock()
		attempts[task.ID]++
		mu.Unlock()

		// Always fail
		return errors.New("permanent failure")
	}

	queue.Start(handler)
	defer queue.Stop()

	task := &Task{
		ID:         "always-fail-task",
		Type:       TypeSlackNotification,
		MaxRetries: 2,
	}

	err := queue.Enqueue(task)
	assert.NoError(t, err)

	// Wait for all retry attempts
	time.Sleep(10 * time.Second)

	mu.Lock()
	totalAttempts := attempts["always-fail-task"]
	mu.Unlock()

	// Should be attempted 3 times (1 initial + 2 retries)
	assert.Equal(t, 3, totalAttempts)
}
func BenchmarkEnqueue(b *testing.B) {
	metrics := &MockMetricsRecorder{}
	queue := NewMemoryQueue(10, 10000, metrics)

	handler := func(ctx context.Context, task *Task) error {
		return nil
	}

	queue.Start(handler)
	defer queue.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := &Task{
			ID:   "bench-task",
			Type: TypeSlackNotification,
		}
		_ = queue.Enqueue(task)
	}
}

func BenchmarkProcessing(b *testing.B) {
	metrics := &MockMetricsRecorder{}
	queue := NewMemoryQueue(10, 10000, metrics)

	handler := func(ctx context.Context, task *Task) error {
		// Simulate minimal work
		time.Sleep(1 * time.Microsecond)
		return nil
	}

	queue.Start(handler)
	defer queue.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := &Task{
			ID:   "bench-task",
			Type: TypeSlackNotification,
		}
		_ = queue.Enqueue(task)
	}

	// Wait for all tasks to complete
	for queue.Size() > 0 {
		time.Sleep(10 * time.Millisecond)
	}
}
