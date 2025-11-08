package metrics

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

// InMemoryCollector implements MetricsCollector using in-memory storage
// Thread-safe and suitable for single-node deployments
type InMemoryCollector struct {
	tasks     []TaskMetric
	decisions []DecisionMetric
	mu        sync.RWMutex
}

// NewInMemoryCollector creates a new in-memory metrics collector
func NewInMemoryCollector() *InMemoryCollector {
	return &InMemoryCollector{
		tasks:     make([]TaskMetric, 0, 1000),
		decisions: make([]DecisionMetric, 0, 100),
	}
}

// RecordTask records a task metric
func (c *InMemoryCollector) RecordTask(metric TaskMetric) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	c.tasks = append(c.tasks, metric)
	return nil
}

// RecordDecision records a decision metric
func (c *InMemoryCollector) RecordDecision(metric DecisionMetric) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	c.decisions = append(c.decisions, metric)
	return nil
}

// GetTaskMetrics retrieves task metrics matching the filter
func (c *InMemoryCollector) GetTaskMetrics(filter TaskFilter) ([]TaskMetric, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]TaskMetric, 0)

	for _, task := range c.tasks {
		if c.matchesFilter(task, filter) {
			result = append(result, task)
		}
	}

	return result, nil
}

// matchesFilter checks if a task matches the given filter
func (c *InMemoryCollector) matchesFilter(task TaskMetric, filter TaskFilter) bool {
	if filter.TaskType != nil && task.TaskType != *filter.TaskType {
		return false
	}
	if filter.Provider != nil && task.Provider != *filter.Provider {
		return false
	}
	if filter.UserUUID != nil && task.UserUUID != *filter.UserUUID {
		return false
	}
	if filter.StartTime != nil && task.Timestamp.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && task.Timestamp.After(*filter.EndTime) {
		return false
	}
	if filter.SuccessOnly && !task.Success {
		return false
	}
	return true
}

// GetProviderStats returns aggregated statistics for a specific provider
func (c *InMemoryCollector) GetProviderStats(provider string, timeRange TimeRange) (*ProviderMetric, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var relevantTasks []TaskMetric
	for _, task := range c.tasks {
		if task.Provider == provider &&
			!task.Timestamp.Before(timeRange.Start) &&
			!task.Timestamp.After(timeRange.End) {
			relevantTasks = append(relevantTasks, task)
		}
	}

	if len(relevantTasks) == 0 {
		return nil, fmt.Errorf("no metrics found for provider %s in time range", provider)
	}

	return c.aggregateMetrics(provider, relevantTasks), nil
}

// GetAllProviderStats returns aggregated statistics for all providers
func (c *InMemoryCollector) GetAllProviderStats(timeRange TimeRange) (map[string]*ProviderMetric, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	providerTasks := make(map[string][]TaskMetric)

	for _, task := range c.tasks {
		if !task.Timestamp.Before(timeRange.Start) && !task.Timestamp.After(timeRange.End) {
			providerTasks[task.Provider] = append(providerTasks[task.Provider], task)
		}
	}

	result := make(map[string]*ProviderMetric)
	for provider, tasks := range providerTasks {
		result[provider] = c.aggregateMetrics(provider, tasks)
	}

	return result, nil
}

// aggregateMetrics calculates aggregated statistics from task metrics
func (c *InMemoryCollector) aggregateMetrics(provider string, tasks []TaskMetric) *ProviderMetric {
	metric := &ProviderMetric{
		Provider:  provider,
		Timestamp: time.Now(),
	}

	metric.RequestCount = int64(len(tasks))

	var totalLatency int64
	var latencies []int64
	var totalCost float64

	for _, task := range tasks {
		if task.Success {
			metric.SuccessCount++
		} else {
			metric.ErrorCount++
		}

		totalLatency += task.DurationMs
		latencies = append(latencies, task.DurationMs)
		totalCost += task.CostUSD
	}

	// Calculate average
	if len(tasks) > 0 {
		metric.AvgLatencyMs = float64(totalLatency) / float64(len(tasks))
	}

	metric.TotalCostUSD = totalCost

	// Calculate uptime percentage
	if metric.RequestCount > 0 {
		metric.Uptime = (float64(metric.SuccessCount) / float64(metric.RequestCount)) * 100.0
	}

	// Calculate percentiles
	if len(latencies) > 0 {
		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

		p50Index := len(latencies) / 2
		p95Index := int(float64(len(latencies)) * 0.95)
		p99Index := int(float64(len(latencies)) * 0.99)

		// Ensure indices are within bounds
		if p50Index >= len(latencies) {
			p50Index = len(latencies) - 1
		}
		if p95Index >= len(latencies) {
			p95Index = len(latencies) - 1
		}
		if p99Index >= len(latencies) {
			p99Index = len(latencies) - 1
		}

		metric.P50LatencyMs = float64(latencies[p50Index])
		metric.P95LatencyMs = float64(latencies[p95Index])
		metric.P99LatencyMs = float64(latencies[p99Index])
	}

	return metric
}

// ExportToCSV exports metrics to a CSV file
func (c *InMemoryCollector) ExportToCSV(filename string, timeRange TimeRange) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	err = writer.Write([]string{
		"TaskID", "TaskType", "Provider", "Timestamp", "DurationMs",
		"Success", "ErrorType", "InputTokens", "OutputTokens",
		"InputSize", "OutputSize", "CostUSD", "QualityScore", "UserUUID",
	})
	if err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, task := range c.tasks {
		if !task.Timestamp.Before(timeRange.Start) && !task.Timestamp.After(timeRange.End) {
			err = writer.Write([]string{
				task.TaskID,
				string(task.TaskType),
				task.Provider,
				task.Timestamp.Format(time.RFC3339),
				fmt.Sprintf("%d", task.DurationMs),
				fmt.Sprintf("%t", task.Success),
				task.ErrorType,
				fmt.Sprintf("%d", task.InputTokens),
				fmt.Sprintf("%d", task.OutputTokens),
				fmt.Sprintf("%d", task.InputSize),
				fmt.Sprintf("%d", task.OutputSize),
				fmt.Sprintf("%.6f", task.CostUSD),
				fmt.Sprintf("%.2f", task.QualityScore),
				task.UserUUID,
			})
			if err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	}

	return nil
}

// ExportToJSON exports metrics to a JSON file
func (c *InMemoryCollector) ExportToJSON(filename string, timeRange TimeRange) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var exportTasks []TaskMetric
	for _, task := range c.tasks {
		if !task.Timestamp.Before(timeRange.Start) && !task.Timestamp.After(timeRange.End) {
			exportTasks = append(exportTasks, task)
		}
	}

	data, err := json.MarshalIndent(exportTasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

// PurgeOldMetrics removes metrics older than the specified duration
func (c *InMemoryCollector) PurgeOldMetrics(olderThan time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)

	// Filter tasks
	var filteredTasks []TaskMetric
	for _, task := range c.tasks {
		if task.Timestamp.After(cutoff) {
			filteredTasks = append(filteredTasks, task)
		}
	}
	c.tasks = filteredTasks

	// Filter decisions
	var filteredDecisions []DecisionMetric
	for _, decision := range c.decisions {
		if decision.Timestamp.After(cutoff) {
			filteredDecisions = append(filteredDecisions, decision)
		}
	}
	c.decisions = filteredDecisions

	return nil
}

// GetMetricsCount returns the current number of metrics stored
func (c *InMemoryCollector) GetMetricsCount() (tasks int, decisions int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.tasks), len(c.decisions)
}
