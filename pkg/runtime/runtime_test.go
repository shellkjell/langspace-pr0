package runtime

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/shellkjell/langspace/pkg/ast"
	"github.com/shellkjell/langspace/pkg/parser"
	"github.com/shellkjell/langspace/pkg/workspace"
)

// parseSource is a helper to parse source code in tests.
func parseSource(t *testing.T, source string) []ast.Entity {
	t.Helper()
	p := parser.New(source)
	result := p.ParseWithRecovery()
	if result.HasErrors() {
		t.Fatalf("parse error: %s", result.ErrorString())
	}
	return result.Entities
}

// addEntities is a helper to add entities to workspace in tests.
func addEntities(t *testing.T, ws *workspace.Workspace, entities []ast.Entity) {
	t.Helper()
	for _, e := range entities {
		if err := ws.AddEntity(e); err != nil {
			t.Fatalf("add entity error: %v", err)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DefaultModel != "claude-sonnet-4-20250514" {
		t.Errorf("expected default model claude-sonnet-4-20250514, got %s", cfg.DefaultModel)
	}
	if cfg.DefaultProvider != "anthropic" {
		t.Errorf("expected default provider anthropic, got %s", cfg.DefaultProvider)
	}
	if cfg.Timeout != 5*time.Minute {
		t.Errorf("expected timeout 5m, got %v", cfg.Timeout)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("expected max retries 3, got %d", cfg.MaxRetries)
	}
	if !cfg.EnableStreaming {
		t.Error("expected streaming to be enabled by default")
	}
	if cfg.Environment == nil {
		t.Error("expected environment map to be initialized")
	}
}

func TestNew(t *testing.T) {
	ws := workspace.New()
	rt := New(ws)

	if rt.workspace != ws {
		t.Error("expected workspace to be set")
	}
	if rt.providers == nil {
		t.Error("expected providers map to be initialized")
	}
	if rt.config == nil {
		t.Error("expected config to be initialized")
	}
}

func TestNewWithOptions(t *testing.T) {
	ws := workspace.New()
	mockProvider := NewMockProvider()

	cfg := &Config{
		DefaultModel:    "gpt-4",
		DefaultProvider: "openai",
		Timeout:         10 * time.Minute,
	}

	rt := New(ws,
		WithConfig(cfg),
		WithProvider("mock", mockProvider),
	)

	if rt.defaultModel != "gpt-4" {
		t.Errorf("expected default model gpt-4, got %s", rt.defaultModel)
	}

	if p, ok := rt.GetProvider("mock"); !ok {
		t.Error("expected mock provider to be registered")
	} else if p != mockProvider {
		t.Error("expected mock provider to match")
	}
}

func TestRegisterProvider(t *testing.T) {
	ws := workspace.New()
	rt := New(ws)
	mockProvider := NewMockProvider()

	rt.RegisterProvider("test", mockProvider)

	p, ok := rt.GetProvider("test")
	if !ok {
		t.Error("expected provider to be found")
	}
	if p != mockProvider {
		t.Error("expected provider to match")
	}
}

func TestGetProviderNotFound(t *testing.T) {
	ws := workspace.New()
	rt := New(ws)

	_, ok := rt.GetProvider("nonexistent")
	if ok {
		t.Error("expected provider not to be found")
	}
}

func TestTokenUsage_Add(t *testing.T) {
	usage := TokenUsage{
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	}

	other := TokenUsage{
		InputTokens:  200,
		OutputTokens: 100,
		TotalTokens:  300,
	}

	usage.Add(other)

	if usage.InputTokens != 300 {
		t.Errorf("expected input tokens 300, got %d", usage.InputTokens)
	}
	if usage.OutputTokens != 150 {
		t.Errorf("expected output tokens 150, got %d", usage.OutputTokens)
	}
	if usage.TotalTokens != 450 {
		t.Errorf("expected total tokens 450, got %d", usage.TotalTokens)
	}
}

func TestExecutionContext_Variables(t *testing.T) {
	ctx := &ExecutionContext{
		Variables: make(map[string]interface{}),
	}

	ctx.SetVariable("test", "value")

	v, ok := ctx.GetVariable("test")
	if !ok {
		t.Error("expected variable to be found")
	}
	if v != "value" {
		t.Errorf("expected value 'value', got %v", v)
	}

	_, ok = ctx.GetVariable("nonexistent")
	if ok {
		t.Error("expected variable not to be found")
	}
}

func TestExecutionContext_StepOutputs(t *testing.T) {
	ctx := &ExecutionContext{}

	_, ok := ctx.GetStepOutput("step1")
	if ok {
		t.Error("expected step output not to be found initially")
	}

	ctx.SetStepOutput("step1", "output1")

	v, ok := ctx.GetStepOutput("step1")
	if !ok {
		t.Error("expected step output to be found")
	}
	if v != "output1" {
		t.Errorf("expected output 'output1', got %v", v)
	}
}

func TestExecuteByName_NotFound(t *testing.T) {
	ws := workspace.New()
	rt := New(ws)

	_, err := rt.ExecuteByName(context.Background(), "intent", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent entity")
	}
	if !strings.Contains(err.Error(), "entity not found") {
		t.Errorf("expected 'entity not found' error, got: %v", err)
	}
}

func TestExecute_UnsupportedEntityType(t *testing.T) {
	source := `agent "test-agent" {
	model: "gpt-4"
	instruction: "Test"
}`
	entities := parseSource(t, source)
	ws := workspace.New()
	addEntities(t, ws, entities)

	rt := New(ws)

	_, err := rt.Execute(context.Background(), entities[0])
	if err == nil {
		t.Error("expected error for unsupported entity type")
	}
	if !strings.Contains(err.Error(), "cannot execute entity of type") {
		t.Errorf("expected 'cannot execute entity of type' error, got: %v", err)
	}
}

func TestExecute_IntentWithMockProvider(t *testing.T) {
	source := `
agent "test-agent" {
	model: "mock-model"
	instruction: "You are a helpful assistant"
}

intent "test-intent" {
	use: agent("test-agent")
	prompt: "Hello, world!"
}
`
	entities := parseSource(t, source)
	ws := workspace.New()
	addEntities(t, ws, entities)

	mockProvider := NewMockProvider(WithMockResponses(MockResponse{
		Content:      "Hello! How can I help you?",
		FinishReason: FinishReasonStop,
	}))
	rt := New(ws, WithProvider("mock", mockProvider))

	intent, found := ws.GetEntityByName("intent", "test-intent")
	if !found {
		t.Fatal("intent not found")
	}

	result, err := rt.Execute(context.Background(), intent)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if !result.Success {
		t.Error("expected execution to succeed")
	}
	if result.Output != "Hello! How can I help you?" {
		t.Errorf("unexpected output: %v", result.Output)
	}
}

func TestExecute_IntentWithInput(t *testing.T) {
	source := `
agent "echo-agent" {
	model: "echo-model"
	instruction: "Echo the input"
}

intent "echo-intent" {
	use: agent("echo-agent")
	prompt: "$input"
}
`
	entities := parseSource(t, source)
	ws := workspace.New()
	addEntities(t, ws, entities)

	echoProvider := NewEchoProvider()
	rt := New(ws, WithProvider("mock", echoProvider))

	intent, found := ws.GetEntityByName("intent", "echo-intent")
	if !found {
		t.Fatal("intent not found")
	}

	result, err := rt.Execute(context.Background(), intent, WithInput("Test message"))
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if !result.Success {
		t.Error("expected execution to succeed")
	}
	output := result.Output.(string)
	if !strings.Contains(output, "Test message") {
		t.Errorf("expected output to contain 'Test message', got: %s", output)
	}
}

func TestExecute_IntentWithStreaming(t *testing.T) {
	source := `
agent "stream-agent" {
	model: "mock-model"
	instruction: "Stream response"
}

intent "stream-intent" {
	use: agent("stream-agent")
	prompt: "Stream test"
}
`
	entities := parseSource(t, source)
	ws := workspace.New()
	addEntities(t, ws, entities)

	mockProvider := NewMockProvider(WithMockResponses(MockResponse{
		Content:      "Streamed response",
		FinishReason: FinishReasonStop,
	}))
	rt := New(ws, WithProvider("mock", mockProvider))

	intent, found := ws.GetEntityByName("intent", "stream-intent")
	if !found {
		t.Fatal("intent not found")
	}

	var chunks []StreamChunk
	handler := &CallbackStreamHandler{
		ChunkFunc:    func(chunk StreamChunk) { chunks = append(chunks, chunk) },
		CompleteFunc: func(response *CompletionResponse) {},
		ErrorFunc:    func(err error) {},
		ProgressFunc: func(event ProgressEvent) {},
	}

	result, err := rt.Execute(context.Background(), intent, WithStreamHandler(handler))
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if !result.Success {
		t.Error("expected execution to succeed")
	}
}

func TestExecute_PipelineWithSteps(t *testing.T) {
	source := `
agent "step-agent" {
	model: "mock-model"
	instruction: "Process step"
}

pipeline "test-pipeline" {
	step "first" {
		use: agent("step-agent")
		prompt: "Step 1"
	}
	step "second" {
		use: agent("step-agent")
		prompt: step("first").output
	}
	output: step("second").output
}
`
	entities := parseSource(t, source)
	ws := workspace.New()
	addEntities(t, ws, entities)

	sequenceProvider := NewSequenceProvider("First step output", "Second step output")
	rt := New(ws, WithProvider("mock", sequenceProvider))

	pipeline, found := ws.GetEntityByName("pipeline", "test-pipeline")
	if !found {
		t.Fatal("pipeline not found")
	}

	result, err := rt.Execute(context.Background(), pipeline)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if !result.Success {
		t.Errorf("expected execution to succeed, error: %v", result.Error)
	}
	if len(result.StepResults) != 2 {
		t.Errorf("expected 2 step results, got %d", len(result.StepResults))
	}
	if result.Output != "Second step output" {
		t.Errorf("expected output 'Second step output', got: %v", result.Output)
	}
}

func TestExecute_WithTimeout(t *testing.T) {
	source := `
agent "slow-agent" {
	model: "mock-model"
	instruction: "Slow response"
}

intent "slow-intent" {
	use: agent("slow-agent")
	prompt: "Slow test"
}
`
	entities := parseSource(t, source)
	ws := workspace.New()
	addEntities(t, ws, entities)

	mockProvider := NewMockProvider(WithMockResponses(MockResponse{
		Content:      "Quick response",
		FinishReason: FinishReasonStop,
	}))
	rt := New(ws, WithProvider("mock", mockProvider))

	intent, found := ws.GetEntityByName("intent", "slow-intent")
	if !found {
		t.Fatal("intent not found")
	}

	result, err := rt.Execute(context.Background(), intent, WithTimeout(1*time.Second))
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if !result.Success {
		t.Error("expected execution to succeed with sufficient timeout")
	}
}

func TestExecute_WithMetadata(t *testing.T) {
	source := `
agent "meta-agent" {
	model: "mock-model"
	instruction: "Metadata test"
}

intent "meta-intent" {
	use: agent("meta-agent")
	prompt: "Test"
}
`
	entities := parseSource(t, source)
	ws := workspace.New()
	addEntities(t, ws, entities)

	mockProvider := NewMockProvider(WithMockResponses(MockResponse{
		Content:      "Response",
		FinishReason: FinishReasonStop,
	}))
	rt := New(ws, WithProvider("mock", mockProvider))

	intent, found := ws.GetEntityByName("intent", "meta-intent")
	if !found {
		t.Fatal("intent not found")
	}

	result, err := rt.Execute(context.Background(), intent,
		WithMetadata("request_id", "12345"),
		WithMetadata("user_id", "user-1"),
	)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if !result.Success {
		t.Error("expected execution to succeed")
	}
	if result.Metadata["request_id"] != "12345" {
		t.Errorf("expected request_id metadata, got: %v", result.Metadata)
	}
}
