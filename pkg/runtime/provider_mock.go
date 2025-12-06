package runtime

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MockProvider implements LLMProvider for testing purposes.
// It allows setting up canned responses and recording requests.
type MockProvider struct {
	name           string
	responses      []MockResponse
	responseIdx    int
	requests       []CompletionRequest
	streamDelay    time.Duration
	chunkSize      int
	errorOnRequest error
	mu             sync.Mutex
}

// MockResponse represents a canned response for the mock provider.
type MockResponse struct {
	Content      string
	ToolCalls    []ToolCall
	FinishReason FinishReason
	Usage        TokenUsage
	Error        error
}

// MockProviderOption is a functional option for configuring MockProvider.
type MockProviderOption func(*MockProvider)

// WithMockName sets the provider name.
func WithMockName(name string) MockProviderOption {
	return func(p *MockProvider) {
		p.name = name
	}
}

// WithMockResponses sets the canned responses.
func WithMockResponses(responses ...MockResponse) MockProviderOption {
	return func(p *MockProvider) {
		p.responses = responses
	}
}

// WithMockStreamDelay sets the delay between stream chunks.
func WithMockStreamDelay(delay time.Duration) MockProviderOption {
	return func(p *MockProvider) {
		p.streamDelay = delay
	}
}

// WithMockChunkSize sets the size of stream chunks.
func WithMockChunkSize(size int) MockProviderOption {
	return func(p *MockProvider) {
		p.chunkSize = size
	}
}

// WithMockError sets an error to return on all requests.
func WithMockError(err error) MockProviderOption {
	return func(p *MockProvider) {
		p.errorOnRequest = err
	}
}

// NewMockProvider creates a new mock provider.
func NewMockProvider(opts ...MockProviderOption) *MockProvider {
	p := &MockProvider{
		name:        "mock",
		responses:   []MockResponse{},
		requests:    []CompletionRequest{},
		streamDelay: 10 * time.Millisecond,
		chunkSize:   10,
	}

	for _, opt := range opts {
		opt(p)
	}

	// Add a default response if none provided
	if len(p.responses) == 0 {
		p.responses = []MockResponse{{
			Content:      "This is a mock response.",
			FinishReason: FinishReasonStop,
			Usage: TokenUsage{
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  150,
			},
		}}
	}

	return p
}

func (p *MockProvider) Name() string {
	return p.name
}

// AddResponse adds a response to the queue.
func (p *MockProvider) AddResponse(resp MockResponse) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.responses = append(p.responses, resp)
}

// SetResponses replaces all responses.
func (p *MockProvider) SetResponses(responses ...MockResponse) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.responses = responses
	p.responseIdx = 0
}

// GetRequests returns all recorded requests.
func (p *MockProvider) GetRequests() []CompletionRequest {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make([]CompletionRequest, len(p.requests))
	copy(result, p.requests)
	return result
}

// LastRequest returns the most recent request.
func (p *MockProvider) LastRequest() *CompletionRequest {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.requests) == 0 {
		return nil
	}
	req := p.requests[len(p.requests)-1]
	return &req
}

// Reset clears all recorded requests and resets response index.
func (p *MockProvider) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.requests = []CompletionRequest{}
	p.responseIdx = 0
}

func (p *MockProvider) getNextResponse() (MockResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.errorOnRequest != nil {
		return MockResponse{}, p.errorOnRequest
	}

	if len(p.responses) == 0 {
		return MockResponse{}, fmt.Errorf("no mock responses configured")
	}

	resp := p.responses[p.responseIdx]
	p.responseIdx = (p.responseIdx + 1) % len(p.responses)
	return resp, nil
}

func (p *MockProvider) recordRequest(req *CompletionRequest) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.requests = append(p.requests, *req)
}

func (p *MockProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	p.recordRequest(req)

	resp, err := p.getNextResponse()
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	return &CompletionResponse{
		Content:      resp.Content,
		ToolCalls:    resp.ToolCalls,
		FinishReason: resp.FinishReason,
		Usage:        resp.Usage,
		Model:        req.Model,
	}, nil
}

func (p *MockProvider) CompleteStream(ctx context.Context, req *CompletionRequest, handler StreamHandler) (*CompletionResponse, error) {
	p.recordRequest(req)

	resp, err := p.getNextResponse()
	if err != nil {
		handler.OnError(err)
		return nil, err
	}

	if resp.Error != nil {
		handler.OnError(resp.Error)
		return nil, resp.Error
	}

	// Simulate streaming by sending chunks
	content := resp.Content
	chunkSize := p.chunkSize
	if chunkSize <= 0 {
		chunkSize = 10
	}

	var sentContent strings.Builder
	chunkIndex := 0

	for i := 0; i < len(content); i += chunkSize {
		select {
		case <-ctx.Done():
			handler.OnError(ctx.Err())
			return nil, ctx.Err()
		default:
		}

		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}

		chunk := content[i:end]
		sentContent.WriteString(chunk)

		handler.OnChunk(StreamChunk{
			Content: chunk,
			Type:    ChunkTypeContent,
			Index:   chunkIndex,
		})
		chunkIndex++

		if p.streamDelay > 0 {
			time.Sleep(p.streamDelay)
		}
	}

	result := &CompletionResponse{
		Content:      sentContent.String(),
		ToolCalls:    resp.ToolCalls,
		FinishReason: resp.FinishReason,
		Usage:        resp.Usage,
		Model:        req.Model,
	}

	handler.OnComplete(result)
	return result, nil
}

func (p *MockProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	return []ModelInfo{
		{ID: "mock-model", Name: "Mock Model", Provider: "mock", MaxTokens: 4096},
	}, nil
}

// EchoProvider is a mock provider that echoes back the input.
// Useful for testing variable resolution and message formatting.
type EchoProvider struct {
	*MockProvider
}

// NewEchoProvider creates an echo provider.
func NewEchoProvider() *EchoProvider {
	return &EchoProvider{
		MockProvider: NewMockProvider(WithMockName("echo")),
	}
}

func (p *EchoProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	p.recordRequest(req)

	// Echo back the last user message
	var lastUserMessage string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == RoleUser {
			lastUserMessage = req.Messages[i].Content
			break
		}
	}

	response := fmt.Sprintf("Echo: %s", lastUserMessage)
	if req.SystemPrompt != "" {
		response = fmt.Sprintf("[System: %s]\n%s", req.SystemPrompt, response)
	}

	return &CompletionResponse{
		Content:      response,
		FinishReason: FinishReasonStop,
		Model:        req.Model,
		Usage: TokenUsage{
			InputTokens:  len(lastUserMessage) / 4,
			OutputTokens: len(response) / 4,
			TotalTokens:  (len(lastUserMessage) + len(response)) / 4,
		},
	}, nil
}

// SequenceProvider returns responses in sequence, useful for multi-turn conversations.
type SequenceProvider struct {
	*MockProvider
}

// NewSequenceProvider creates a sequence provider with the given responses.
func NewSequenceProvider(responses ...string) *SequenceProvider {
	mockResponses := make([]MockResponse, len(responses))
	for i, content := range responses {
		mockResponses[i] = MockResponse{
			Content:      content,
			FinishReason: FinishReasonStop,
			Usage: TokenUsage{
				InputTokens:  100,
				OutputTokens: len(content) / 4,
				TotalTokens:  100 + len(content)/4,
			},
		}
	}

	return &SequenceProvider{
		MockProvider: NewMockProvider(
			WithMockName("sequence"),
			WithMockResponses(mockResponses...),
		),
	}
}
