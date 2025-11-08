package metrics

import "time"

// NoOpCollector is a metrics collector that does nothing
// Used when metrics collection is disabled
type NoOpCollector struct{}

// NewNoOpCollector creates a new no-op metrics collector
func NewNoOpCollector() *NoOpCollector {
	return &NoOpCollector{}
}

// RecordTask does nothing
func (n *NoOpCollector) RecordTask(metric TaskMetric) error {
	return nil
}

// RecordDecision does nothing
func (n *NoOpCollector) RecordDecision(metric DecisionMetric) error {
	return nil
}

// GetTaskMetrics returns an empty slice
func (n *NoOpCollector) GetTaskMetrics(filter TaskFilter) ([]TaskMetric, error) {
	return nil, nil
}

// GetProviderStats returns nil
func (n *NoOpCollector) GetProviderStats(provider string, timeRange TimeRange) (*ProviderMetric, error) {
	return nil, nil
}

// GetAllProviderStats returns an empty map
func (n *NoOpCollector) GetAllProviderStats(timeRange TimeRange) (map[string]*ProviderMetric, error) {
	return nil, nil
}

// ExportToCSV does nothing
func (n *NoOpCollector) ExportToCSV(filename string, timeRange TimeRange) error {
	return nil
}

// ExportToJSON does nothing
func (n *NoOpCollector) ExportToJSON(filename string, timeRange TimeRange) error {
	return nil
}

// PurgeOldMetrics does nothing
func (n *NoOpCollector) PurgeOldMetrics(olderThan time.Duration) error {
	return nil
}
