package metrics

import "time"

// TaskType represents the type of task being measured
type TaskType string

const (
	TaskTypeEmbedding    TaskType = "embedding"
	TaskTypeCompletion   TaskType = "completion"
	TaskTypeRerank       TaskType = "rerank"
	TaskTypeMLParse      TaskType = "ml_parse"
	TaskTypeQueryRewrite TaskType = "query_rewrite"
	TaskTypeVectorSearch TaskType = "vector_search"
	TaskTypeOCR          TaskType = "ocr"
	TaskTypeDocConv      TaskType = "docconv"
)

// TaskMetric represents a single task measurement
type TaskMetric struct {
	// Identification
	TaskID   string
	TaskType TaskType
	Provider string // "openai", "gemini", "ollama", "ml_worker", etc.

	// Timing
	Timestamp  time.Time
	DurationMs int64

	// Outcome
	Success   bool
	ErrorType string

	// Performance details
	InputTokens  int
	OutputTokens int
	InputSize    int64 // bytes
	OutputSize   int64 // bytes

	// Cost
	CostUSD float64

	// Quality (optional)
	QualityScore float64

	// Context
	UserUUID string
	Metadata map[string]interface{}
}

// ProviderMetric represents aggregated statistics for a provider
type ProviderMetric struct {
	Provider  string
	Timestamp time.Time

	// Volume
	RequestCount int64
	SuccessCount int64
	ErrorCount   int64

	// Latency statistics (milliseconds)
	AvgLatencyMs float64
	P50LatencyMs float64
	P95LatencyMs float64
	P99LatencyMs float64

	// Cost
	TotalCostUSD float64

	// Availability (percentage)
	Uptime float64
}

// DecisionMetric tracks decision-making (e.g., provider selection, feature flags)
type DecisionMetric struct {
	DecisionID   string
	DecisionType string // "provider_selection", "feature_toggle", "fallback", etc.
	Timestamp    time.Time

	// Decision details
	ChosenOption string
	Alternatives []string
	Reason       string

	// Outcome
	Success    bool
	DurationMs int64

	// Impact
	CostSavings  float64
	QualityDelta float64

	// Context
	UserUUID string
	Metadata map[string]interface{}
}

// TaskFilter for querying metrics
type TaskFilter struct {
	TaskType    *TaskType
	Provider    *string
	UserUUID    *string
	StartTime   *time.Time
	EndTime     *time.Time
	SuccessOnly bool
}

// TimeRange represents a time period for queries
type TimeRange struct {
	Start time.Time
	End   time.Time
}
