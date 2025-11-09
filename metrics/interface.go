package metrics

import "time"

// MetricsCollector defines the interface for collecting and retrieving metrics
type MetricsCollector interface {
	// Task tracking
	RecordTask(metric TaskMetric) error
	RecordDecision(metric DecisionMetric) error

	// Retrieval
	GetTaskMetrics(filter TaskFilter) ([]TaskMetric, error)
	GetProviderStats(provider string, timeRange TimeRange) (*ProviderMetric, error)
	GetAllProviderStats(timeRange TimeRange) (map[string]*ProviderMetric, error)

	// Export
	ExportToCSV(filename string, timeRange TimeRange) error
	ExportToJSON(filename string, timeRange TimeRange) error

	// Cleanup
	PurgeOldMetrics(olderThan time.Duration) error
}
