package llm

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// ProviderMode determines how providers are selected
type ProviderMode string

const (
	// ProviderModeSingle uses a single provider
	ProviderModeSingle ProviderMode = "single"

	// ProviderModeAuto automatically selects the best provider based on metrics
	ProviderModeAuto ProviderMode = "auto"

	// ProviderModeFailover tries providers in order until one succeeds
	ProviderModeFailover ProviderMode = "failover"
)

// ProviderManager manages multiple LLM providers with auto-selection and failover
type ProviderManager struct {
	mode          ProviderMode
	collector     *MetricsCollector
	autoSelector  *AutoSelector
	singleProvider *InstrumentedProvider
	providerType  ProviderType
	mu            sync.RWMutex

	// Configuration
	enableMetrics bool
	metricsPath   string
}

// ProviderManagerConfig holds configuration for the provider manager
type ProviderManagerConfig struct {
	Mode           ProviderMode
	EnableMetrics  bool
	MetricsPath    string
	SelectorWeights *SelectorWeights
}

// DefaultProviderManagerConfig returns default configuration
func DefaultProviderManagerConfig() *ProviderManagerConfig {
	weights := DefaultSelectorWeights()
	return &ProviderManagerConfig{
		Mode:           ProviderModeSingle,
		EnableMetrics:  true,
		MetricsPath:    "./data/llm_metrics.json",
		SelectorWeights: &weights,
	}
}

// NewProviderManager creates a new provider manager
func NewProviderManager(config *ProviderManagerConfig) *ProviderManager {
	if config == nil {
		config = DefaultProviderManagerConfig()
	}

	var collector *MetricsCollector
	if config.EnableMetrics {
		collector = NewMetricsCollector(10000, config.MetricsPath)
	}

	var autoSelector *AutoSelector
	if config.Mode == ProviderModeAuto && config.SelectorWeights != nil {
		autoSelector = NewAutoSelector(collector, *config.SelectorWeights)
	}

	return &ProviderManager{
		mode:          config.Mode,
		collector:     collector,
		autoSelector:  autoSelector,
		enableMetrics: config.EnableMetrics,
		metricsPath:   config.MetricsPath,
	}
}

// NewProviderManagerFromEnv creates a provider manager from environment variables
func NewProviderManagerFromEnv() (*ProviderManager, error) {
	// Determine mode
	modeStr := os.Getenv("LLM_PROVIDER_MODE")
	if modeStr == "" {
		modeStr = "single"
	}
	mode := ProviderMode(modeStr)

	// Check if metrics are enabled
	enableMetrics := os.Getenv("LLM_METRICS_ENABLED") != "false" // Default true

	metricsPath := os.Getenv("LLM_METRICS_PATH")
	if metricsPath == "" {
		metricsPath = "./data/llm_metrics.json"
	}

	config := &ProviderManagerConfig{
		Mode:          mode,
		EnableMetrics: enableMetrics,
		MetricsPath:   metricsPath,
	}

	if mode == ProviderModeAuto {
		weights := DefaultSelectorWeights()
		config.SelectorWeights = &weights
	}

	manager := NewProviderManager(config)

	// Load providers based on mode
	switch mode {
	case ProviderModeSingle:
		// Load single provider from LLM_PROVIDER env var
		provider, err := NewProviderFromEnv()
		if err != nil {
			return nil, fmt.Errorf("failed to create provider: %w", err)
		}

		providerType := ProviderType(os.Getenv("LLM_PROVIDER"))
		if providerType == "" {
			providerType = ProviderTypeOpenAI
		}

		model := getModelForProvider(providerType)
		if err := manager.SetSingleProvider(provider, providerType, model); err != nil {
			return nil, fmt.Errorf("failed to set single provider: %w", err)
		}

	case ProviderModeAuto, ProviderModeFailover:
		// Load multiple providers
		if err := manager.LoadMultipleProvidersFromEnv(); err != nil {
			return nil, fmt.Errorf("failed to load multiple providers: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported provider mode: %s", mode)
	}

	return manager, nil
}

// SetSingleProvider sets a single provider for the manager
func (pm *ProviderManager) SetSingleProvider(provider LLMProvider, providerType ProviderType, model string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.enableMetrics && pm.collector != nil {
		pm.singleProvider = NewInstrumentedProvider(provider, providerType, pm.collector, model)
	} else {
		// Create a dummy collector for the instrumented provider
		dummyCollector := NewMetricsCollector(0, "")
		pm.singleProvider = NewInstrumentedProvider(provider, providerType, dummyCollector, model)
	}

	pm.providerType = providerType
	pm.mode = ProviderModeSingle

	log.Printf("[ProviderManager] Set single provider: %s (%s)", providerType, model)

	return nil
}

// RegisterProvider registers a provider with the auto-selector
func (pm *ProviderManager) RegisterProvider(providerType ProviderType, provider LLMProvider, model string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.autoSelector == nil {
		weights := DefaultSelectorWeights()
		pm.autoSelector = NewAutoSelector(pm.collector, weights)
	}

	instrumentedProvider := NewInstrumentedProvider(provider, providerType, pm.collector, model)
	pm.autoSelector.RegisterProvider(providerType, instrumentedProvider)

	log.Printf("[ProviderManager] Registered provider: %s (%s)", providerType, model)

	return nil
}

// LoadMultipleProvidersFromEnv loads all configured providers from environment variables
func (pm *ProviderManager) LoadMultipleProvidersFromEnv() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.autoSelector == nil {
		weights := DefaultSelectorWeights()
		pm.autoSelector = NewAutoSelector(pm.collector, weights)
	}

	providersLoaded := 0

	// Try to load each provider type if API key is present
	providerConfigs := []struct {
		providerType ProviderType
		apiKeyEnv    string
		modelEnv     string
		defaultModel string
	}{
		{ProviderTypeOllama, "OLLAMA_HOST", "OLLAMA_MODEL", "llama2"},
		{ProviderTypeOpenAI, "OPENAI_API_KEY", "OPENAI_MODEL", "gpt-3.5-turbo"},
		{ProviderTypeClaude, "CLAUDE_API_KEY", "CLAUDE_MODEL", "claude-3-5-sonnet-20241022"},
		{ProviderTypeGemini, "GEMINI_API_KEY", "GEMINI_MODEL", "gemini-1.5-flash"},
		{ProviderTypeGroq, "GROQ_API_KEY", "GROQ_MODEL", "mixtral-8x7b-32768"},
	}

	for _, pc := range providerConfigs {
		// Check if provider is configured
		apiKey := os.Getenv(pc.apiKeyEnv)
		if apiKey == "" && pc.providerType != ProviderTypeOllama {
			continue
		}

		// Special handling for Ollama (doesn't require API key)
		if pc.providerType == ProviderTypeOllama {
			// Check if Ollama is available
			ollamaHost := os.Getenv("OLLAMA_HOST")
			if ollamaHost == "" {
				ollamaHost = os.Getenv("OLLAMA_BASE_URL")
			}
			if ollamaHost == "" {
				continue // Skip if not configured
			}
		}

		// Create provider
		config := &Config{Provider: string(pc.providerType)}

		switch pc.providerType {
		case ProviderTypeOllama:
			config.OllamaConfig = DefaultOllamaConfig()
		case ProviderTypeOpenAI:
			config.OpenAIConfig = DefaultOpenAIConfig(apiKey)
		case ProviderTypeClaude:
			config.ClaudeConfig = DefaultClaudeConfig(apiKey)
		case ProviderTypeGemini:
			config.GeminiConfig = DefaultGeminiConfig(apiKey)
		case ProviderTypeGroq:
			config.GroqConfig = DefaultGroqConfig(apiKey)
		}

		provider, err := NewProvider(config)
		if err != nil {
			log.Printf("[ProviderManager] Failed to create %s provider: %v", pc.providerType, err)
			continue
		}

		model := os.Getenv(pc.modelEnv)
		if model == "" {
			model = pc.defaultModel
		}

		instrumentedProvider := NewInstrumentedProvider(provider, pc.providerType, pm.collector, model)
		pm.autoSelector.RegisterProvider(pc.providerType, instrumentedProvider)

		log.Printf("[ProviderManager] Loaded provider: %s (%s)", pc.providerType, model)
		providersLoaded++
	}

	if providersLoaded == 0 {
		return fmt.Errorf("no providers could be loaded. Please configure at least one provider via environment variables")
	}

	log.Printf("[ProviderManager] Loaded %d providers in %s mode", providersLoaded, pm.mode)

	return nil
}

// GetProvider returns a provider based on the current mode
func (pm *ProviderManager) GetProvider() (LLMProvider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	switch pm.mode {
	case ProviderModeSingle:
		if pm.singleProvider == nil {
			return nil, fmt.Errorf("no provider configured in single mode")
		}
		return pm.singleProvider, nil

	case ProviderModeAuto:
		if pm.autoSelector == nil {
			return nil, fmt.Errorf("auto-selector not initialized")
		}
		provider, _, err := pm.autoSelector.SelectBestProvider()
		return provider, err

	case ProviderModeFailover:
		// For failover, return the first healthy provider
		if pm.autoSelector == nil {
			return nil, fmt.Errorf("auto-selector not initialized for failover mode")
		}
		provider, _, err := pm.autoSelector.SelectBestProvider()
		return provider, err

	default:
		return nil, fmt.Errorf("unsupported provider mode: %s", pm.mode)
	}
}

// GetProviderWithFallback tries to get a provider, with fallback on failure
func (pm *ProviderManager) GetProviderWithFallback() (LLMProvider, ProviderType, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	switch pm.mode {
	case ProviderModeSingle:
		if pm.singleProvider == nil {
			return nil, "", fmt.Errorf("no provider configured")
		}
		return pm.singleProvider, pm.providerType, nil

	case ProviderModeAuto, ProviderModeFailover:
		if pm.autoSelector == nil {
			return nil, "", fmt.Errorf("auto-selector not initialized")
		}
		return pm.autoSelector.SelectBestProvider()

	default:
		return nil, "", fmt.Errorf("unsupported provider mode: %s", pm.mode)
	}
}

// GetMetrics returns the metrics collector
func (pm *ProviderManager) GetMetrics() *MetricsCollector {
	return pm.collector
}

// GetStats returns statistics for all providers
func (pm *ProviderManager) GetStats() map[ProviderType]*ProviderStats {
	if pm.collector == nil {
		return nil
	}
	return pm.collector.GetAllStats()
}

// GetMode returns the current provider mode
func (pm *ProviderManager) GetMode() ProviderMode {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.mode
}

// Helper function to get model name for a provider type
func getModelForProvider(providerType ProviderType) string {
	switch providerType {
	case ProviderTypeOllama:
		return getEnvOrDefault("OLLAMA_MODEL", "llama2")
	case ProviderTypeOpenAI:
		return getEnvOrDefault("OPENAI_MODEL", "gpt-3.5-turbo")
	case ProviderTypeClaude:
		return getEnvOrDefault("CLAUDE_MODEL", "claude-3-5-sonnet-20241022")
	case ProviderTypeGemini:
		return getEnvOrDefault("GEMINI_MODEL", "gemini-1.5-flash")
	case ProviderTypeGroq:
		return getEnvOrDefault("GROQ_MODEL", "mixtral-8x7b-32768")
	default:
		return "unknown"
	}
}
