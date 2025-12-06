// Package runtime provides the execution runtime for LangSpace.
// It handles LLM integration, intent/pipeline execution, variable resolution,
// and streaming output.
package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shellkjell/langspace/pkg/ast"
	"github.com/shellkjell/langspace/pkg/workspace"
)

// Runtime is the main execution engine for LangSpace.
// It coordinates LLM providers, variable resolution, and workflow execution.
type Runtime struct {
	workspace    *workspace.Workspace
	providers    map[string]LLMProvider
	defaultModel string
	config       *Config
	mu           sync.RWMutex
}

// Config holds runtime configuration options.
type Config struct {
	// DefaultModel is the model to use when not specified
	DefaultModel string `json:"default_model"`

	// DefaultProvider is the provider to use when not specified
	DefaultProvider string `json:"default_provider"`

	// Timeout is the default timeout for LLM requests
	Timeout time.Duration `json:"timeout"`

	// MaxRetries is the maximum number of retries for failed requests
	MaxRetries int `json:"max_retries"`

	// EnableStreaming enables streaming responses by default
	EnableStreaming bool `json:"enable_streaming"`

	// Environment variables (can be overridden)
	Environment map[string]string `json:"environment"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		DefaultModel:    "claude-sonnet-4-20250514",
		DefaultProvider: "anthropic",
		Timeout:         5 * time.Minute,
		MaxRetries:      3,
		EnableStreaming: true,
		Environment:     make(map[string]string),
	}
}

// New creates a new Runtime with the given workspace.
func New(ws *workspace.Workspace, opts ...Option) *Runtime {
	r := &Runtime{
		workspace:    ws,
		providers:    make(map[string]LLMProvider),
		config:       DefaultConfig(),
		defaultModel: "claude-sonnet-4-20250514",
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Option is a functional option for configuring the Runtime.
type Option func(*Runtime)

// WithConfig sets the runtime configuration.
func WithConfig(cfg *Config) Option {
	return func(r *Runtime) {
		if cfg != nil {
			r.config = cfg
			if cfg.DefaultModel != "" {
				r.defaultModel = cfg.DefaultModel
			}
		}
	}
}

// WithProvider registers an LLM provider.
func WithProvider(name string, provider LLMProvider) Option {
	return func(r *Runtime) {
		r.providers[name] = provider
	}
}

// RegisterProvider registers an LLM provider by name.
func (r *Runtime) RegisterProvider(name string, provider LLMProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = provider
}

// GetProvider returns a provider by name.
func (r *Runtime) GetProvider(name string) (LLMProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

// Execute runs an entity (intent or pipeline) and returns the result.
func (r *Runtime) Execute(ctx context.Context, entity ast.Entity, opts ...ExecuteOption) (*ExecutionResult, error) {
	execOpts := &executeOptions{
		input:    nil,
		handler:  nil,
		timeout:  r.config.Timeout,
		metadata: make(map[string]string),
	}

	for _, opt := range opts {
		opt(execOpts)
	}

	// Create execution context
	execCtx := &ExecutionContext{
		Context:   ctx,
		Runtime:   r,
		Workspace: r.workspace,
		Variables: make(map[string]interface{}),
		Metadata:  execOpts.metadata,
		Handler:   execOpts.handler,
		StartTime: time.Now(),
	}

	// Set input variable if provided
	if execOpts.input != nil {
		execCtx.Variables["input"] = execOpts.input
	}

	// Apply timeout
	if execOpts.timeout > 0 {
		var cancel context.CancelFunc
		execCtx.Context, cancel = context.WithTimeout(ctx, execOpts.timeout)
		defer cancel()
	}

	// Dispatch based on entity type
	switch entity.Type() {
	case "intent":
		return r.executeIntent(execCtx, entity)
	case "pipeline":
		return r.executePipeline(execCtx, entity)
	default:
		return nil, fmt.Errorf("cannot execute entity of type %q", entity.Type())
	}
}

// ExecuteByName looks up and executes an entity by type and name.
func (r *Runtime) ExecuteByName(ctx context.Context, entityType, entityName string, opts ...ExecuteOption) (*ExecutionResult, error) {
	entity, found := r.workspace.GetEntityByName(entityType, entityName)
	if !found {
		return nil, fmt.Errorf("entity not found: %s %q", entityType, entityName)
	}
	return r.Execute(ctx, entity, opts...)
}

// executeOptions holds options for a single execution.
type executeOptions struct {
	input    interface{}
	handler  StreamHandler
	timeout  time.Duration
	metadata map[string]string
}

// ExecuteOption is a functional option for Execute.
type ExecuteOption func(*executeOptions)

// WithInput sets the input for execution.
func WithInput(input interface{}) ExecuteOption {
	return func(o *executeOptions) {
		o.input = input
	}
}

// WithStreamHandler sets the stream handler for streaming output.
func WithStreamHandler(handler StreamHandler) ExecuteOption {
	return func(o *executeOptions) {
		o.handler = handler
	}
}

// WithTimeout sets the execution timeout.
func WithTimeout(timeout time.Duration) ExecuteOption {
	return func(o *executeOptions) {
		o.timeout = timeout
	}
}

// WithMetadata sets execution metadata.
func WithMetadata(key, value string) ExecuteOption {
	return func(o *executeOptions) {
		o.metadata[key] = value
	}
}

// ExecutionContext holds the context for a single execution.
type ExecutionContext struct {
	Context   context.Context
	Runtime   *Runtime
	Workspace *workspace.Workspace
	Variables map[string]interface{}
	Metadata  map[string]string
	Handler   StreamHandler
	StartTime time.Time

	// For pipeline execution
	StepOutputs map[string]interface{}
}

// SetVariable sets a variable in the execution context.
func (ec *ExecutionContext) SetVariable(name string, value interface{}) {
	ec.Variables[name] = value
}

// GetVariable gets a variable from the execution context.
func (ec *ExecutionContext) GetVariable(name string) (interface{}, bool) {
	v, ok := ec.Variables[name]
	return v, ok
}

// SetStepOutput sets the output of a step (for pipeline execution).
func (ec *ExecutionContext) SetStepOutput(stepName string, output interface{}) {
	if ec.StepOutputs == nil {
		ec.StepOutputs = make(map[string]interface{})
	}
	ec.StepOutputs[stepName] = output
}

// GetStepOutput gets the output of a step.
func (ec *ExecutionContext) GetStepOutput(stepName string) (interface{}, bool) {
	if ec.StepOutputs == nil {
		return nil, false
	}
	v, ok := ec.StepOutputs[stepName]
	return v, ok
}

// EmitProgress sends a progress event if a handler is registered.
func (ec *ExecutionContext) EmitProgress(event ProgressEvent) {
	if ec.Handler != nil {
		ec.Handler.OnProgress(event)
	}
}

// EmitChunk sends a streaming chunk if a handler is registered.
func (ec *ExecutionContext) EmitChunk(chunk StreamChunk) {
	if ec.Handler != nil {
		ec.Handler.OnChunk(chunk)
	}
}

// ExecutionResult represents the result of executing an entity.
type ExecutionResult struct {
	// Success indicates whether execution completed successfully
	Success bool `json:"success"`

	// Output is the final output of the execution
	Output interface{} `json:"output,omitempty"`

	// Error contains any error that occurred
	Error error `json:"error,omitempty"`

	// Duration is how long execution took
	Duration time.Duration `json:"duration"`

	// StepResults contains results from pipeline steps (for pipelines)
	StepResults map[string]*StepResult `json:"step_results,omitempty"`

	// Metadata contains execution metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// TokensUsed tracks token usage
	TokensUsed TokenUsage `json:"tokens_used,omitempty"`
}

// StepResult represents the result of a single pipeline step.
type StepResult struct {
	Name      string        `json:"name"`
	Success   bool          `json:"success"`
	Output    interface{}   `json:"output,omitempty"`
	Error     error         `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
}

// TokenUsage tracks LLM token usage.
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// Add adds token usage from another TokenUsage.
func (t *TokenUsage) Add(other TokenUsage) {
	t.InputTokens += other.InputTokens
	t.OutputTokens += other.OutputTokens
	t.TotalTokens += other.TotalTokens
}
