package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

// Task represents a unit of work to be processed
type Task struct {
	ID         string
	Type       string
	Payload    interface{}
	Retries    int
	MaxRetries int
	CreatedAt  time.Time
	EnqueuedAt time.Time
}

// TaskHandler is a function type that processes a task
type TaskHandler func(context.Context, *Task) error

// MemoryQueue implements an in-memory task queue with worker pool
type MemoryQueue struct {
	tasks        chan *Task
	workers      int
	queueSize    int
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	handler      TaskHandler
	metrics      MetricsRecorder
	mu           sync.RWMutex
	isRunning    bool
	taskCount    int64
	droppedCount int64
}

// MetricsRecorder interface for recording queue metrics
type MetricsRecorder interface {
	RecordTaskEnqueued(taskType string)
	RecordTaskProcessed(taskType string, duration time.Duration)
	RecordTaskFailed(taskType string)
	RecordTaskRetried(taskType string)
	RecordTaskDropped(taskType string)
	UpdateQueueSize(size int)
	UpdateQueueWorkers(workers int)
}

// NewMemoryQueue creates a new in-memory queue
func NewMemoryQueue(workers, queueSize int, metrics MetricsRecorder) *MemoryQueue {
	ctx, cancel := context.WithCancel(context.Background())

	return &MemoryQueue{
		tasks:     make(chan *Task, queueSize),
		workers:   workers,
		queueSize: queueSize,
		ctx:       ctx,
		cancel:    cancel,
		metrics:   metrics,
		isRunning: false,
	}
}

// Start initializes the worker pool and begins processing tasks
func (mq *MemoryQueue) Start(handler TaskHandler) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if mq.isRunning {
		klog.Warning("Queue is already running")
		return
	}

	mq.handler = handler
	mq.isRunning = true

	klog.Infof("Starting memory queue with %d workers and buffer size %d", mq.workers, mq.queueSize)

	// Start worker goroutines
	for i := 0; i < mq.workers; i++ {
		mq.wg.Add(1)
		go mq.worker(i)
	}

	// Start metrics updater
	go mq.metricsUpdater()

	if mq.metrics != nil {
		mq.metrics.UpdateQueueWorkers(mq.workers)
	}
}

// worker processes tasks from the queue
func (mq *MemoryQueue) worker(id int) {
	defer mq.wg.Done()

	klog.V(2).Infof("Queue worker %d started", id)

	for {
		select {
		case <-mq.ctx.Done():
			klog.V(2).Infof("Queue worker %d shutting down", id)
			return

		case task, ok := <-mq.tasks:
			if !ok {
				klog.V(2).Infof("Queue worker %d: channel closed", id)
				return
			}

			mq.processTask(id, task)
		}
	}
}

// processTask handles a single task with retry logic
func (mq *MemoryQueue) processTask(workerID int, task *Task) {
	start := time.Now()
	waitTime := start.Sub(task.EnqueuedAt)

	klog.V(3).Infof("Worker %d processing task %s (type: %s, attempt: %d/%d, wait time: %v)",
		workerID, task.ID, task.Type, task.Retries+1, task.MaxRetries+1, waitTime)

	err := mq.handler(mq.ctx, task)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Worker %d failed to process task %s: %v (attempt %d/%d)",
			workerID, task.ID, err, task.Retries+1, task.MaxRetries+1)

		if mq.metrics != nil {
			mq.metrics.RecordTaskFailed(task.Type)
		}

		// Retry logic with exponential backoff
		if task.Retries < task.MaxRetries {
			task.Retries++

			// Calculate backoff duration (exponential: 1s, 2s, 4s, 8s, ...)
			backoffDuration := time.Duration(1<<uint(task.Retries-1)) * time.Second
			if backoffDuration > 30*time.Second {
				backoffDuration = 30 * time.Second // Cap at 30 seconds
			}

			klog.Infof("Retrying task %s after %v (attempt %d/%d)",
				task.ID, backoffDuration, task.Retries+1, task.MaxRetries+1)

			if mq.metrics != nil {
				mq.metrics.RecordTaskRetried(task.Type)
			}

			// Wait before retry
			time.Sleep(backoffDuration)

			// Re-enqueue the task
			select {
			case mq.tasks <- task:
				klog.V(3).Infof("Task %s re-enqueued for retry", task.ID)
			case <-mq.ctx.Done():
				klog.Warningf("Cannot retry task %s: queue shutting down", task.ID)
			default:
				klog.Errorf("Cannot retry task %s: queue is full", task.ID)
				if mq.metrics != nil {
					mq.metrics.RecordTaskDropped(task.Type)
				}
			}
		} else {
			klog.Errorf("Task %s failed permanently after %d attempts", task.ID, task.MaxRetries+1)
		}
	} else {
		klog.V(2).Infof("Worker %d successfully processed task %s in %v", workerID, task.ID, duration)

		if mq.metrics != nil {
			mq.metrics.RecordTaskProcessed(task.Type, duration)
		}
	}
}

// Enqueue adds a task to the queue
func (mq *MemoryQueue) Enqueue(task *Task) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if !mq.isRunning {
		return fmt.Errorf("queue is not running")
	}

	// Set enqueue timestamp
	task.EnqueuedAt = time.Now()

	// Generate ID if not set
	if task.ID == "" {
		task.ID = fmt.Sprintf("%s-%d", task.Type, time.Now().UnixNano())
	}

	// Set default max retries if not specified
	if task.MaxRetries == 0 {
		task.MaxRetries = 3
	}

	select {
	case mq.tasks <- task:
		mq.taskCount++
		klog.V(3).Infof("Task %s enqueued (queue size: %d/%d)", task.ID, len(mq.tasks), mq.queueSize)

		if mq.metrics != nil {
			mq.metrics.RecordTaskEnqueued(task.Type)
		}
		return nil

	default:
		mq.droppedCount++
		klog.Errorf("Queue is full, dropping task %s (type: %s)", task.ID, task.Type)

		if mq.metrics != nil {
			mq.metrics.RecordTaskDropped(task.Type)
		}
		return fmt.Errorf("queue is full")
	}
}

// Stop gracefully shuts down the queue
func (mq *MemoryQueue) Stop() {
	mq.mu.Lock()
	if !mq.isRunning {
		mq.mu.Unlock()
		return
	}
	mq.isRunning = false
	mq.mu.Unlock()

	klog.Info("Stopping memory queue...")
	pendingTasks := len(mq.tasks)

	if pendingTasks > 0 {
		klog.Infof("Draining %d pending tasks...", pendingTasks)
	}

	// Signal shutdown
	mq.cancel()

	// Cancellation stops workers. Keep the channel open while in-flight retry
	// handlers may still send to it; closing it here can panic during shutdown.

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		mq.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		klog.Info("All queue workers stopped gracefully")
	case <-time.After(30 * time.Second):
		klog.Warning("Queue shutdown timeout - some tasks may not have completed")
	}

	klog.Infof("Queue stopped. Total processed: %d, Dropped: %d", mq.taskCount, mq.droppedCount)
}

// Size returns the current number of tasks in the queue
func (mq *MemoryQueue) Size() int {
	return len(mq.tasks)
}

// Stats returns queue statistics
func (mq *MemoryQueue) Stats() map[string]interface{} {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	return map[string]interface{}{
		"workers":         mq.workers,
		"queue_size":      mq.queueSize,
		"current_size":    len(mq.tasks),
		"is_running":      mq.isRunning,
		"tasks_processed": mq.taskCount,
		"tasks_dropped":   mq.droppedCount,
	}
}

// metricsUpdater periodically updates queue size metrics
func (mq *MemoryQueue) metricsUpdater() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mq.ctx.Done():
			return
		case <-ticker.C:
			if mq.metrics != nil {
				mq.metrics.UpdateQueueSize(len(mq.tasks))
			}
		}
	}
}
