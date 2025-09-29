package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// Resource event counters
	ResourceEventsTotal *prometheus.CounterVec

	// Slack notification metrics
	SlackNotificationsTotal   *prometheus.CounterVec
	SlackNotificationDuration prometheus.Histogram
	SlackNotificationErrors   *prometheus.CounterVec

	// Informer metrics
	InformerSyncStatus         *prometheus.GaugeVec
	InformerLastSyncTime       *prometheus.GaugeVec
	InformerEventsProcessed    *prometheus.CounterVec
	InformerProcessingDuration prometheus.Histogram

	// Application health metrics
	ApplicationUptime      prometheus.Gauge
	WatchedResourcesCount  prometheus.Gauge
	WatchedNamespacesCount prometheus.Gauge

	// Resource diff metrics
	ResourceDiffComputations *prometheus.CounterVec
	ResourceDiffDuration     prometheus.Histogram

	// Cache metrics
	ResourceCacheSize   *prometheus.GaugeVec
	ResourceCacheHits   *prometheus.CounterVec
	ResourceCacheMisses *prometheus.CounterVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		// Resource event counters
		ResourceEventsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "k8s_diff_informer_resource_events_total",
				Help: "Total number of Kubernetes resource events processed",
			},
			[]string{"resource_type", "resource_name", "event_type", "namespace", "cluster"},
		),

		// Slack notification metrics
		SlackNotificationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "k8s_diff_informer_slack_notifications_total",
				Help: "Total number of Slack notifications sent",
			},
			[]string{"status", "cluster"},
		),

		SlackNotificationDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "k8s_diff_informer_slack_notification_duration_seconds",
				Help:    "Duration of Slack notification requests",
				Buckets: prometheus.DefBuckets,
			},
		),

		SlackNotificationErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "k8s_diff_informer_slack_notification_errors_total",
				Help: "Total number of Slack notification errors",
			},
			[]string{"error_type", "cluster"},
		),

		// Informer metrics
		InformerSyncStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "k8s_diff_informer_sync_status",
				Help: "Status of informer sync (1 = synced, 0 = not synced)",
			},
			[]string{"resource_type", "cluster"},
		),

		InformerLastSyncTime: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "k8s_diff_informer_last_sync_timestamp",
				Help: "Timestamp of last successful informer sync",
			},
			[]string{"resource_type", "cluster"},
		),

		InformerEventsProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "k8s_diff_informer_events_processed_total",
				Help: "Total number of events processed by informers",
			},
			[]string{"resource_type", "event_type", "cluster"},
		),

		InformerProcessingDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "k8s_diff_informer_processing_duration_seconds",
				Help:    "Duration of event processing",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
			},
		),

		// Application health metrics
		ApplicationUptime: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "k8s_diff_informer_uptime_seconds",
				Help: "Application uptime in seconds",
			},
		),

		WatchedResourcesCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "k8s_diff_informer_watched_resources_count",
				Help: "Number of resources being watched",
			},
		),

		WatchedNamespacesCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "k8s_diff_informer_watched_namespaces_count",
				Help: "Number of namespaces being watched",
			},
		),

		// Resource diff metrics
		ResourceDiffComputations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "k8s_diff_informer_diff_computations_total",
				Help: "Total number of resource diff computations",
			},
			[]string{"resource_type", "has_changes", "cluster"},
		),

		ResourceDiffDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "k8s_diff_informer_diff_computation_duration_seconds",
				Help:    "Duration of resource diff computations",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5},
			},
		),

		// Cache metrics
		ResourceCacheSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "k8s_diff_informer_cache_size",
				Help: "Size of resource cache",
			},
			[]string{"resource_type", "cluster"},
		),

		ResourceCacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "k8s_diff_informer_cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"resource_type", "cluster"},
		),

		ResourceCacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "k8s_diff_informer_cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"resource_type", "cluster"},
		),
	}
}

// RecordResourceEvent records a Kubernetes resource event
func (m *Metrics) RecordResourceEvent(resourceType, resourceName, eventType, namespace, cluster string) {
	m.ResourceEventsTotal.WithLabelValues(resourceType, resourceName, eventType, namespace, cluster).Inc()
}

// RecordSlackNotification records a Slack notification attempt
func (m *Metrics) RecordSlackNotification(status, cluster string, duration time.Duration) {
	m.SlackNotificationsTotal.WithLabelValues(status, cluster).Inc()
	m.SlackNotificationDuration.Observe(duration.Seconds())
}

// RecordSlackError records a Slack notification error
func (m *Metrics) RecordSlackError(errorType, cluster string) {
	m.SlackNotificationErrors.WithLabelValues(errorType, cluster).Inc()
}

// SetInformerSyncStatus sets the sync status for an informer
func (m *Metrics) SetInformerSyncStatus(resourceType, cluster string, synced bool) {
	value := 0.0
	if synced {
		value = 1.0
	}
	m.InformerSyncStatus.WithLabelValues(resourceType, cluster).Set(value)
}

// UpdateLastSyncTime updates the last sync time for an informer
func (m *Metrics) UpdateLastSyncTime(resourceType, cluster string) {
	m.InformerLastSyncTime.WithLabelValues(resourceType, cluster).SetToCurrentTime()
}

// RecordEventProcessed records a processed event
func (m *Metrics) RecordEventProcessed(resourceType, eventType, cluster string, duration time.Duration) {
	m.InformerEventsProcessed.WithLabelValues(resourceType, eventType, cluster).Inc()
	m.InformerProcessingDuration.Observe(duration.Seconds())
}

// SetWatchedResourcesCount sets the number of watched resources
func (m *Metrics) SetWatchedResourcesCount(count int) {
	m.WatchedResourcesCount.Set(float64(count))
}

// SetWatchedNamespacesCount sets the number of watched namespaces
func (m *Metrics) SetWatchedNamespacesCount(count int) {
	m.WatchedNamespacesCount.Set(float64(count))
}

// RecordDiffComputation records a diff computation
func (m *Metrics) RecordDiffComputation(resourceType, cluster string, hasChanges bool, duration time.Duration) {
	hasChangesStr := "false"
	if hasChanges {
		hasChangesStr = "true"
	}
	m.ResourceDiffComputations.WithLabelValues(resourceType, hasChangesStr, cluster).Inc()
	m.ResourceDiffDuration.Observe(duration.Seconds())
}

// UpdateCacheSize updates the cache size for a resource type
func (m *Metrics) UpdateCacheSize(resourceType, cluster string, size int) {
	m.ResourceCacheSize.WithLabelValues(resourceType, cluster).Set(float64(size))
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit(resourceType, cluster string) {
	m.ResourceCacheHits.WithLabelValues(resourceType, cluster).Inc()
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss(resourceType, cluster string) {
	m.ResourceCacheMisses.WithLabelValues(resourceType, cluster).Inc()
}

// StartUptimeRecorder starts recording application uptime
func (m *Metrics) StartUptimeRecorder() {
	startTime := time.Now()
	go func() {
		for {
			m.ApplicationUptime.Set(time.Since(startTime).Seconds())
			time.Sleep(10 * time.Second)
		}
	}()
}
