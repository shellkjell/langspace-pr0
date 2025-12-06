package runtime

import (
	"context"
)

// LLMProvider defines the interface for LLM providers.
// Implementations handle communication with specific LLM APIs.
type LLMProvider interface {
	// Name returns the provider name (e.g., "anthropic", "openai")
	Name() string

	// Complete sends a completion request and returns the response.
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// CompleteStream sends a completion request and streams the response.
	CompleteStream(ctx context.Context, req *CompletionRequest, handler StreamHandler) (*CompletionResponse, error)

	// ListModels returns available models for this provider.
	ListModels(ctx context.Context) ([]ModelInfo, error)
}

// CompletionRequest represents a request to an LLM.
type CompletionRequest struct {
	// Model specifies which model to use
	Model string `json:"model"`

	// Messages is the conversation history
	Messages []Message `json:"messages"`

	// SystemPrompt is the system instruction
	SystemPrompt string `json:"system_prompt,omitempty"`

	// Temperature controls randomness (0-1)
	Temperature float64 `json:"temperature,omitempty"`

	// MaxTokens limits the response length
	MaxTokens int `json:"max_tokens,omitempty"`

	// Tools available to the model
	Tools []ToolDefinition `json:"tools,omitempty"`

	// StopSequences to end generation
	StopSequences []string `json:"stop_sequences,omitempty"`

	// Metadata for tracking/logging
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Message represents a message in the conversation.
type Message struct {
	Role    MessageRole `json:"role"`
	Content string      `json:"content"`

	// ToolCalls contains any tool calls made by the assistant
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID is set for tool result messages
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// MessageRole represents the role of a message sender.
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// ToolDefinition describes a tool available to the model.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ToolCall represents a tool invocation by the model.
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// CompletionResponse represents a response from an LLM.
type CompletionResponse struct {
	// Content is the generated text
	Content string `json:"content"`

	// ToolCalls contains any tool calls made
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// FinishReason indicates why generation stopped
	FinishReason FinishReason `json:"finish_reason"`

	// Usage contains token counts
	Usage TokenUsage `json:"usage"`

	// Model that was used
	Model string `json:"model"`
}

// FinishReason indicates why the model stopped generating.
type FinishReason string

const (
	FinishReasonStop      FinishReason = "stop"
	FinishReasonLength    FinishReason = "length"
	FinishReasonToolUse   FinishReason = "tool_use"
	FinishReasonError     FinishReason = "error"
	FinishReasonCancelled FinishReason = "cancelled"
)

// ModelInfo contains information about an available model.
type ModelInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Provider     string   `json:"provider"`
	MaxTokens    int      `json:"max_tokens"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// StreamHandler receives streaming events during execution.
type StreamHandler interface {
	// OnChunk is called when a new chunk of content is received
	OnChunk(chunk StreamChunk)

	// OnProgress is called for progress updates
	OnProgress(event ProgressEvent)

	// OnComplete is called when streaming is complete
	OnComplete(response *CompletionResponse)

	// OnError is called when an error occurs
	OnError(err error)
}

// StreamChunk represents a chunk of streamed content.
type StreamChunk struct {
	// Content is the text chunk
	Content string `json:"content"`

	// Type indicates the chunk type
	Type ChunkType `json:"type"`

	// Index for ordering chunks
	Index int `json:"index"`

	// Delta contains incremental token usage
	Delta *TokenUsage `json:"delta,omitempty"`
}

// ChunkType indicates the type of stream chunk.
type ChunkType string

const (
	ChunkTypeContent   ChunkType = "content"
	ChunkTypeToolStart ChunkType = "tool_start"
	ChunkTypeToolEnd   ChunkType = "tool_end"
)

// ProgressEvent represents a progress update during execution.
type ProgressEvent struct {
	// Type of progress event
	Type ProgressType `json:"type"`

	// Message describing the progress
	Message string `json:"message"`

	// Step name (for pipeline execution)
	Step string `json:"step,omitempty"`

	// Progress percentage (0-100, if applicable)
	Progress int `json:"progress,omitempty"`

	// Metadata for additional context
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ProgressType indicates the type of progress event.
type ProgressType string

const (
	ProgressTypeStart    ProgressType = "start"
	ProgressTypeStep     ProgressType = "step"
	ProgressTypeComplete ProgressType = "complete"
	ProgressTypeError    ProgressType = "error"
)

// DefaultStreamHandler provides a no-op implementation of StreamHandler.
type DefaultStreamHandler struct{}

func (h *DefaultStreamHandler) OnChunk(chunk StreamChunk)               {}
func (h *DefaultStreamHandler) OnProgress(event ProgressEvent)          {}
func (h *DefaultStreamHandler) OnComplete(response *CompletionResponse) {}
func (h *DefaultStreamHandler) OnError(err error)                       {}

// CallbackStreamHandler allows using callbacks for stream handling.
type CallbackStreamHandler struct {
	ChunkFunc    func(StreamChunk)
	ProgressFunc func(ProgressEvent)
	CompleteFunc func(*CompletionResponse)
	ErrorFunc    func(error)
}

func (h *CallbackStreamHandler) OnChunk(chunk StreamChunk) {
	if h.ChunkFunc != nil {
		h.ChunkFunc(chunk)
	}
}

func (h *CallbackStreamHandler) OnProgress(event ProgressEvent) {
	if h.ProgressFunc != nil {
		h.ProgressFunc(event)
	}
}

func (h *CallbackStreamHandler) OnComplete(response *CompletionResponse) {
	if h.CompleteFunc != nil {
		h.CompleteFunc(response)
	}
}

func (h *CallbackStreamHandler) OnError(err error) {
	if h.ErrorFunc != nil {
		h.ErrorFunc(err)
	}
}

// BufferedStreamHandler collects all chunks for later access.
type BufferedStreamHandler struct {
	Chunks   []StreamChunk
	Events   []ProgressEvent
	Response *CompletionResponse
	Err      error
}

func (h *BufferedStreamHandler) OnChunk(chunk StreamChunk) {
	h.Chunks = append(h.Chunks, chunk)
}

func (h *BufferedStreamHandler) OnProgress(event ProgressEvent) {
	h.Events = append(h.Events, event)
}

func (h *BufferedStreamHandler) OnComplete(response *CompletionResponse) {
	h.Response = response
}

func (h *BufferedStreamHandler) OnError(err error) {
	h.Err = err
}

// Content returns all collected content as a single string.
func (h *BufferedStreamHandler) Content() string {
	var content string
	for _, chunk := range h.Chunks {
		if chunk.Type == ChunkTypeContent {
			content += chunk.Content
		}
	}
	return content
}
