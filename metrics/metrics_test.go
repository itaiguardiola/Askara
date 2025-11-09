package metrics

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// Helper function to create a pointer to TaskType
func ptrTaskType(t TaskType) *TaskType {
	return &t
}

// Helper function to create a pointer to string
func ptrString(s string) *string {
	return &s
}

// Helper function to create a pointer to time.Time
func ptrTime(t time.Time) *time.Time {
	return &t
}

func TestInMemoryCollector_RecordTask(t *testing.T) {
	collector := NewInMemoryCollector()

	task := TaskMetric{
		TaskID:     "task1",
		TaskType:   TaskTypeEmbedding,
		Provider:   "openai",
		DurationMs: 150,
		Success:    true,
		CostUSD:    0.001,
	}

	err := collector.RecordTask(task)
	if err != nil {
		t.Errorf("Failed to record task: %v", err)
	}

	// Verify timestamp was set
	tasks, _ := collector.GetTaskMetrics(TaskFilter{})
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	if tasks[0].Timestamp.IsZero() {
		t.Error("Timestamp should be set automatically")
	}
}

func TestInMemoryCollector_GetTaskMetrics_FilterByType(t *testing.T) {
	collector := NewInMemoryCollector()

	// Add tasks of different types
	collector.RecordTask(TaskMetric{
		TaskID:   "task1",
		TaskType: TaskTypeEmbedding,
		Provider: "openai",
		Success:  true,
	})

	collector.RecordTask(TaskMetric{
		TaskID:   "task2",
		TaskType: TaskTypeCompletion,
		Provider: "openai",
		Success:  true,
	})

	collector.RecordTask(TaskMetric{
		TaskID:   "task3",
		TaskType: TaskTypeEmbedding,
		Provider: "gemini",
		Success:  true,
	})

	// Filter by type
	filter := TaskFilter{
		TaskType: ptrTaskType(TaskTypeEmbedding),
	}

	tasks, err := collector.GetTaskMetrics(filter)
	if err != nil {
		t.Errorf("Failed to retrieve tasks: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("Expected 2 embedding tasks, got %d", len(tasks))
	}

	for _, task := range tasks {
		if task.TaskType != TaskTypeEmbedding {
			t.Errorf("Expected TaskTypeEmbedding, got %s", task.TaskType)
		}
	}
}

func TestInMemoryCollector_GetTaskMetrics_FilterByProvider(t *testing.T) {
	collector := NewInMemoryCollector()

	collector.RecordTask(TaskMetric{
		TaskID:   "task1",
		TaskType: TaskTypeEmbedding,
		Provider: "openai",
		Success:  true,
	})

	collector.RecordTask(TaskMetric{
		TaskID:   "task2",
		TaskType: TaskTypeEmbedding,
		Provider: "gemini",
		Success:  true,
	})

	// Filter by provider
	filter := TaskFilter{
		Provider: ptrString("openai"),
	}

	tasks, err := collector.GetTaskMetrics(filter)
	if err != nil {
		t.Errorf("Failed to retrieve tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	if tasks[0].Provider != "openai" {
		t.Errorf("Expected openai provider, got %s", tasks[0].Provider)
	}
}

func TestInMemoryCollector_GetTaskMetrics_FilterByTimeRange(t *testing.T) {
	collector := NewInMemoryCollector()

	now := time.Now()
	past := now.Add(-2 * time.Hour)

	collector.RecordTask(TaskMetric{
		TaskID:    "task1",
		TaskType:  TaskTypeEmbedding,
		Provider:  "openai",
		Timestamp: past,
		Success:   true,
	})

	collector.RecordTask(TaskMetric{
		TaskID:    "task2",
		TaskType:  TaskTypeEmbedding,
		Provider:  "openai",
		Timestamp: now,
		Success:   true,
	})

	// Filter by time range (last hour only)
	filter := TaskFilter{
		StartTime: ptrTime(now.Add(-1 * time.Hour)),
	}

	tasks, err := collector.GetTaskMetrics(filter)
	if err != nil {
		t.Errorf("Failed to retrieve tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task in time range, got %d", len(tasks))
	}

	if tasks[0].TaskID != "task2" {
		t.Errorf("Expected task2, got %s", tasks[0].TaskID)
	}
}

func TestInMemoryCollector_GetTaskMetrics_FilterSuccessOnly(t *testing.T) {
	collector := NewInMemoryCollector()

	collector.RecordTask(TaskMetric{
		TaskID:   "task1",
		TaskType: TaskTypeEmbedding,
		Provider: "openai",
		Success:  true,
	})

	collector.RecordTask(TaskMetric{
		TaskID:    "task2",
		TaskType:  TaskTypeEmbedding,
		Provider:  "openai",
		Success:   false,
		ErrorType: "rate_limit",
	})

	// Filter success only
	filter := TaskFilter{
		SuccessOnly: true,
	}

	tasks, err := collector.GetTaskMetrics(filter)
	if err != nil {
		t.Errorf("Failed to retrieve tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 successful task, got %d", len(tasks))
	}

	if !tasks[0].Success {
		t.Error("Expected success=true")
	}
}

func TestInMemoryCollector_GetProviderStats(t *testing.T) {
	collector := NewInMemoryCollector()

	// Add multiple tasks with varying latencies
	latencies := []int64{100, 150, 200, 250, 300, 350, 400, 450, 500, 1000}

	for i, latency := range latencies {
		collector.RecordTask(TaskMetric{
			TaskID:     fmt.Sprintf("task%d", i),
			TaskType:   TaskTypeCompletion,
			Provider:   "openai",
			DurationMs: latency,
			Success:    i < 9, // Last one fails
			CostUSD:    0.01,
		})
	}

	timeRange := TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now().Add(1 * time.Hour),
	}

	stats, err := collector.GetProviderStats("openai", timeRange)
	if err != nil {
		t.Errorf("Failed to get provider stats: %v", err)
	}

	// Verify counts
	if stats.RequestCount != 10 {
		t.Errorf("Expected 10 requests, got %d", stats.RequestCount)
	}

	if stats.SuccessCount != 9 {
		t.Errorf("Expected 9 successes, got %d", stats.SuccessCount)
	}

	if stats.ErrorCount != 1 {
		t.Errorf("Expected 1 error, got %d", stats.ErrorCount)
	}

	// Verify uptime
	expectedUptime := 90.0
	if stats.Uptime != expectedUptime {
		t.Errorf("Expected %.2f%% uptime, got %.2f%%", expectedUptime, stats.Uptime)
	}

	// Verify average latency
	expectedAvg := 370.0 // Sum of latencies / 10
	if stats.AvgLatencyMs != expectedAvg {
		t.Errorf("Expected avg latency %.2f, got %.2f", expectedAvg, stats.AvgLatencyMs)
	}

	// Verify total cost (use epsilon for float comparison)
	expectedCost := 0.10 // 10 * 0.01
	epsilon := 0.00001
	if (stats.TotalCostUSD - expectedCost) > epsilon || (expectedCost - stats.TotalCostUSD) > epsilon {
		t.Errorf("Expected total cost %.4f, got %.4f", expectedCost, stats.TotalCostUSD)
	}

	// Verify percentiles
	// With 10 elements [100, 150, 200, 250, 300, 350, 400, 450, 500, 1000]
	// P50 at index 5 = 350
	if stats.P50LatencyMs != 350.0 {
		t.Errorf("Expected P50 latency 350ms, got %.2f", stats.P50LatencyMs)
	}

	// P95 at index 9 (0.95 * 10 = 9.5, floored to 9) = 1000
	if stats.P95LatencyMs != 1000.0 {
		t.Errorf("Expected P95 latency 1000ms, got %.2f", stats.P95LatencyMs)
	}
}

func TestInMemoryCollector_GetAllProviderStats(t *testing.T) {
	collector := NewInMemoryCollector()

	// Add tasks for multiple providers
	collector.RecordTask(TaskMetric{
		TaskID:     "task1",
		TaskType:   TaskTypeEmbedding,
		Provider:   "openai",
		DurationMs: 150,
		Success:    true,
		CostUSD:    0.001,
	})

	collector.RecordTask(TaskMetric{
		TaskID:     "task2",
		TaskType:   TaskTypeEmbedding,
		Provider:   "gemini",
		DurationMs: 100,
		Success:    true,
		CostUSD:    0.0001,
	})

	collector.RecordTask(TaskMetric{
		TaskID:     "task3",
		TaskType:   TaskTypeEmbedding,
		Provider:   "gemini",
		DurationMs: 120,
		Success:    true,
		CostUSD:    0.0001,
	})

	timeRange := TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now().Add(1 * time.Hour),
	}

	stats, err := collector.GetAllProviderStats(timeRange)
	if err != nil {
		t.Errorf("Failed to get all provider stats: %v", err)
	}

	if len(stats) != 2 {
		t.Errorf("Expected stats for 2 providers, got %d", len(stats))
	}

	// Verify OpenAI stats
	openaiStats, ok := stats["openai"]
	if !ok {
		t.Error("Expected openai stats")
	}
	if openaiStats.RequestCount != 1 {
		t.Errorf("Expected 1 openai request, got %d", openaiStats.RequestCount)
	}

	// Verify Gemini stats
	geminiStats, ok := stats["gemini"]
	if !ok {
		t.Error("Expected gemini stats")
	}
	if geminiStats.RequestCount != 2 {
		t.Errorf("Expected 2 gemini requests, got %d", geminiStats.RequestCount)
	}

	// Verify Gemini is cheaper
	if geminiStats.TotalCostUSD >= openaiStats.TotalCostUSD {
		t.Errorf("Expected Gemini to be cheaper: Gemini $%.6f vs OpenAI $%.6f",
			geminiStats.TotalCostUSD, openaiStats.TotalCostUSD)
	}
}

func TestInMemoryCollector_RecordDecision(t *testing.T) {
	collector := NewInMemoryCollector()

	decision := DecisionMetric{
		DecisionID:   "decision1",
		DecisionType: "provider_selection",
		ChosenOption: "gemini",
		Alternatives: []string{"openai", "ollama"},
		Reason:       "cost_optimization",
		Success:      true,
		CostSavings:  0.0009,
	}

	err := collector.RecordDecision(decision)
	if err != nil {
		t.Errorf("Failed to record decision: %v", err)
	}

	_, decisions := collector.GetMetricsCount()
	if decisions != 1 {
		t.Errorf("Expected 1 decision, got %d", decisions)
	}
}

func TestInMemoryCollector_ExportToCSV(t *testing.T) {
	collector := NewInMemoryCollector()

	collector.RecordTask(TaskMetric{
		TaskID:       "export_test",
		TaskType:     TaskTypeEmbedding,
		Provider:     "gemini",
		DurationMs:   100,
		Success:      true,
		CostUSD:      0.0001,
		InputTokens:  100,
		OutputTokens: 0,
		UserUUID:     "test-user",
	})

	tmpFile := "/tmp/metrics_test.csv"
	defer os.Remove(tmpFile)

	timeRange := TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now().Add(1 * time.Hour),
	}

	err := collector.ExportToCSV(tmpFile, timeRange)
	if err != nil {
		t.Errorf("Failed to export CSV: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("CSV file was not created")
	}

	// Read file and verify content
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Errorf("Failed to read CSV file: %v", err)
	}

	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Error("CSV file is empty")
	}

	// Verify header exists
	if !contains(contentStr, "TaskID") || !contains(contentStr, "Provider") {
		t.Error("CSV file missing expected headers")
	}

	// Verify data exists
	if !contains(contentStr, "export_test") || !contains(contentStr, "gemini") {
		t.Error("CSV file missing expected data")
	}
}

func TestInMemoryCollector_ExportToJSON(t *testing.T) {
	collector := NewInMemoryCollector()

	collector.RecordTask(TaskMetric{
		TaskID:     "json_test",
		TaskType:   TaskTypeCompletion,
		Provider:   "openai",
		DurationMs: 200,
		Success:    true,
		CostUSD:    0.002,
	})

	tmpFile := "/tmp/metrics_test.json"
	defer os.Remove(tmpFile)

	timeRange := TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now().Add(1 * time.Hour),
	}

	err := collector.ExportToJSON(tmpFile, timeRange)
	if err != nil {
		t.Errorf("Failed to export JSON: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("JSON file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Errorf("Failed to read JSON file: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "json_test") || !contains(contentStr, "openai") {
		t.Error("JSON file missing expected data")
	}
}

func TestInMemoryCollector_PurgeOldMetrics(t *testing.T) {
	collector := NewInMemoryCollector()

	now := time.Now()
	old := now.Add(-48 * time.Hour)

	// Add old metric
	collector.RecordTask(TaskMetric{
		TaskID:    "old_task",
		TaskType:  TaskTypeEmbedding,
		Provider:  "openai",
		Timestamp: old,
		Success:   true,
	})

	// Add recent metric
	collector.RecordTask(TaskMetric{
		TaskID:    "recent_task",
		TaskType:  TaskTypeEmbedding,
		Provider:  "openai",
		Timestamp: now,
		Success:   true,
	})

	// Purge metrics older than 24 hours
	err := collector.PurgeOldMetrics(24 * time.Hour)
	if err != nil {
		t.Errorf("Failed to purge old metrics: %v", err)
	}

	// Verify only recent metric remains
	tasks, _ := collector.GetTaskMetrics(TaskFilter{})
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task after purge, got %d", len(tasks))
	}

	if tasks[0].TaskID != "recent_task" {
		t.Errorf("Expected recent_task to remain, got %s", tasks[0].TaskID)
	}
}

func TestNoOpCollector(t *testing.T) {
	collector := NewNoOpCollector()

	// All operations should succeed without error
	err := collector.RecordTask(TaskMetric{})
	if err != nil {
		t.Errorf("NoOpCollector.RecordTask should not return error: %v", err)
	}

	err = collector.RecordDecision(DecisionMetric{})
	if err != nil {
		t.Errorf("NoOpCollector.RecordDecision should not return error: %v", err)
	}

	tasks, err := collector.GetTaskMetrics(TaskFilter{})
	if err != nil || tasks != nil {
		t.Error("NoOpCollector.GetTaskMetrics should return nil, nil")
	}

	stats, err := collector.GetProviderStats("any", TimeRange{})
	if err != nil || stats != nil {
		t.Error("NoOpCollector.GetProviderStats should return nil, nil")
	}

	allStats, err := collector.GetAllProviderStats(TimeRange{})
	if err != nil || allStats != nil {
		t.Error("NoOpCollector.GetAllProviderStats should return nil, nil")
	}

	err = collector.ExportToCSV("dummy.csv", TimeRange{})
	if err != nil {
		t.Errorf("NoOpCollector.ExportToCSV should not return error: %v", err)
	}

	err = collector.ExportToJSON("dummy.json", TimeRange{})
	if err != nil {
		t.Errorf("NoOpCollector.ExportToJSON should not return error: %v", err)
	}

	err = collector.PurgeOldMetrics(24 * time.Hour)
	if err != nil {
		t.Errorf("NoOpCollector.PurgeOldMetrics should not return error: %v", err)
	}
}

func TestInMemoryCollector_ConcurrentAccess(t *testing.T) {
	collector := NewInMemoryCollector()

	// Simulate concurrent writes
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				collector.RecordTask(TaskMetric{
					TaskID:   fmt.Sprintf("task-%d-%d", id, j),
					TaskType: TaskTypeEmbedding,
					Provider: "openai",
					Success:  true,
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all tasks recorded
	tasks, _ := collector.GetMetricsCount()
	expected := 1000
	if tasks != expected {
		t.Errorf("Expected %d tasks after concurrent writes, got %d", expected, tasks)
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
