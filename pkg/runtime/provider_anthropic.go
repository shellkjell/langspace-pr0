package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// AnthropicProvider implements LLMProvider for the Anthropic API.
type AnthropicProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	version    string
}

// AnthropicOption is a functional option for configuring AnthropicProvider.
type AnthropicOption func(*AnthropicProvider)

// WithAnthropicAPIKey sets the API key.
func WithAnthropicAPIKey(key string) AnthropicOption {
	return func(p *AnthropicProvider) {
		p.apiKey = key
	}
}

// WithAnthropicBaseURL sets a custom base URL.
func WithAnthropicBaseURL(url string) AnthropicOption {
	return func(p *AnthropicProvider) {
		p.baseURL = url
	}
}

// WithAnthropicHTTPClient sets a custom HTTP client.
func WithAnthropicHTTPClient(client *http.Client) AnthropicOption {
	return func(p *AnthropicProvider) {
		p.httpClient = client
	}
}

// NewAnthropicProvider creates a new Anthropic provider.
func NewAnthropicProvider(opts ...AnthropicOption) *AnthropicProvider {
	p := &AnthropicProvider{
		baseURL:    "https://api.anthropic.com",
		httpClient: http.DefaultClient,
		version:    "2023-06-01",
	}

	// Check for API key in environment
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		p.apiKey = key
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// anthropicRequest is the request format for Anthropic's API.
type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
	Tools       []anthropicTool    `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content,omitempty"`
}

type anthropicContentBlock struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
	Content   string                 `json:"content,omitempty"`
}

type anthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// anthropicResponse is the response format from Anthropic's API.
type anthropicResponse struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Role       string                  `json:"role"`
	Content    []anthropicContentBlock `json:"content"`
	Model      string                  `json:"model"`
	StopReason string                  `json:"stop_reason"`
	Usage      anthropicUsage          `json:"usage"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (p *AnthropicProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("anthropic API key not set")
	}

	// Convert messages to Anthropic format
	anthropicMsgs := make([]anthropicMessage, 0, len(req.Messages))
	for _, msg := range req.Messages {
		role := string(msg.Role)
		if role == "tool" {
			role = "user" // Tool results come from user role in Anthropic
		}

		content := []anthropicContentBlock{}

		// Handle regular text content
		if msg.Content != "" {
			if msg.ToolCallID != "" {
				// This is a tool result
				content = append(content, anthropicContentBlock{
					Type:      "tool_result",
					ToolUseID: msg.ToolCallID,
					Content:   msg.Content,
				})
			} else {
				content = append(content, anthropicContentBlock{
					Type: "text",
					Text: msg.Content,
				})
			}
		}

		// Handle tool calls
		for _, tc := range msg.ToolCalls {
			content = append(content, anthropicContentBlock{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  tc.Name,
				Input: tc.Arguments,
			})
		}

		anthropicMsgs = append(anthropicMsgs, anthropicMessage{
			Role:    role,
			Content: content,
		})
	}

	// Convert tools to Anthropic format
	var anthropicTools []anthropicTool
	for _, tool := range req.Tools {
		anthropicTools = append(anthropicTools, anthropicTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.Parameters,
		})
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	anthropicReq := anthropicRequest{
		Model:       req.Model,
		Messages:    anthropicMsgs,
		System:      req.SystemPrompt,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
		Tools:       anthropicTools,
	}

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", p.version)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var anthropicResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return p.convertResponse(&anthropicResp), nil
}

func (p *AnthropicProvider) convertResponse(resp *anthropicResponse) *CompletionResponse {
	result := &CompletionResponse{
		Model: resp.Model,
		Usage: TokenUsage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
			TotalTokens:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}

	// Extract content and tool calls
	var contentParts []string
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			contentParts = append(contentParts, block.Text)
		case "tool_use":
			result.ToolCalls = append(result.ToolCalls, ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: block.Input,
			})
		}
	}
	result.Content = strings.Join(contentParts, "")

	// Convert stop reason
	switch resp.StopReason {
	case "end_turn", "stop_sequence":
		result.FinishReason = FinishReasonStop
	case "max_tokens":
		result.FinishReason = FinishReasonLength
	case "tool_use":
		result.FinishReason = FinishReasonToolUse
	default:
		result.FinishReason = FinishReasonStop
	}

	return result
}

func (p *AnthropicProvider) CompleteStream(ctx context.Context, req *CompletionRequest, handler StreamHandler) (*CompletionResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("anthropic API key not set")
	}

	// Convert messages to Anthropic format
	anthropicMsgs := make([]anthropicMessage, 0, len(req.Messages))
	for _, msg := range req.Messages {
		role := string(msg.Role)
		if role == "tool" {
			role = "user"
		}

		content := []anthropicContentBlock{}
		if msg.Content != "" {
			if msg.ToolCallID != "" {
				content = append(content, anthropicContentBlock{
					Type:      "tool_result",
					ToolUseID: msg.ToolCallID,
					Content:   msg.Content,
				})
			} else {
				content = append(content, anthropicContentBlock{
					Type: "text",
					Text: msg.Content,
				})
			}
		}

		anthropicMsgs = append(anthropicMsgs, anthropicMessage{
			Role:    role,
			Content: content,
		})
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	anthropicReq := anthropicRequest{
		Model:       req.Model,
		Messages:    anthropicMsgs,
		System:      req.SystemPrompt,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
		Stream:      true,
	}

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", p.version)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return p.handleStream(resp.Body, handler)
}

func (p *AnthropicProvider) handleStream(body io.Reader, handler StreamHandler) (*CompletionResponse, error) {
	result := &CompletionResponse{}
	var contentBuilder strings.Builder
	chunkIndex := 0

	reader := NewSSEReader(body)
	for {
		event, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			handler.OnError(err)
			return nil, err
		}

		switch event.Event {
		case "content_block_delta":
			var delta struct {
				Type  string `json:"type"`
				Index int    `json:"index"`
				Delta struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(event.Data), &delta); err != nil {
				continue
			}

			if delta.Delta.Type == "text_delta" {
				contentBuilder.WriteString(delta.Delta.Text)
				handler.OnChunk(StreamChunk{
					Content: delta.Delta.Text,
					Type:    ChunkTypeContent,
					Index:   chunkIndex,
				})
				chunkIndex++
			}

		case "message_delta":
			var delta struct {
				Delta struct {
					StopReason string `json:"stop_reason"`
				} `json:"delta"`
				Usage anthropicUsage `json:"usage"`
			}
			if err := json.Unmarshal([]byte(event.Data), &delta); err != nil {
				continue
			}

			result.Usage.OutputTokens = delta.Usage.OutputTokens
			result.Usage.TotalTokens = result.Usage.InputTokens + result.Usage.OutputTokens

			switch delta.Delta.StopReason {
			case "end_turn", "stop_sequence":
				result.FinishReason = FinishReasonStop
			case "max_tokens":
				result.FinishReason = FinishReasonLength
			case "tool_use":
				result.FinishReason = FinishReasonToolUse
			}

		case "message_start":
			var msg struct {
				Message struct {
					Model string         `json:"model"`
					Usage anthropicUsage `json:"usage"`
				} `json:"message"`
			}
			if err := json.Unmarshal([]byte(event.Data), &msg); err != nil {
				continue
			}
			result.Model = msg.Message.Model
			result.Usage.InputTokens = msg.Message.Usage.InputTokens

		case "message_stop":
			// Stream complete
		}
	}

	result.Content = contentBuilder.String()
	handler.OnComplete(result)
	return result, nil
}

func (p *AnthropicProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	// Anthropic doesn't have a models endpoint, return known models
	return []ModelInfo{
		{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", Provider: "anthropic", MaxTokens: 200000},
		{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", Provider: "anthropic", MaxTokens: 200000},
		{ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", Provider: "anthropic", MaxTokens: 200000},
		{ID: "claude-3-haiku-20240307", Name: "Claude 3 Haiku", Provider: "anthropic", MaxTokens: 200000},
	}, nil
}

// SSEReader reads Server-Sent Events from an io.Reader.
type SSEReader struct {
	reader *bufioReader
}

type bufioReader struct {
	r   io.Reader
	buf []byte
	pos int
	end int
}

func newBufioReader(r io.Reader) *bufioReader {
	return &bufioReader{r: r, buf: make([]byte, 4096)}
}

func (b *bufioReader) ReadLine() (string, error) {
	var line []byte
	for {
		// Look for newline in buffer
		for i := b.pos; i < b.end; i++ {
			if b.buf[i] == '\n' {
				line = append(line, b.buf[b.pos:i]...)
				b.pos = i + 1
				// Remove trailing \r if present
				if len(line) > 0 && line[len(line)-1] == '\r' {
					line = line[:len(line)-1]
				}
				return string(line), nil
			}
		}

		// Save remaining buffer
		if b.pos < b.end {
			line = append(line, b.buf[b.pos:b.end]...)
		}

		// Read more data
		n, err := b.r.Read(b.buf)
		b.pos = 0
		b.end = n
		if err != nil {
			if len(line) > 0 {
				return string(line), nil
			}
			return "", err
		}
	}
}

// SSEEvent represents a Server-Sent Event.
type SSEEvent struct {
	Event string
	Data  string
}

// NewSSEReader creates a new SSE reader.
func NewSSEReader(r io.Reader) *SSEReader {
	return &SSEReader{reader: newBufioReader(r)}
}

// Next reads the next SSE event.
func (r *SSEReader) Next() (*SSEEvent, error) {
	event := &SSEEvent{}

	for {
		line, err := r.reader.ReadLine()
		if err != nil {
			return nil, err
		}

		// Empty line marks end of event
		if line == "" {
			if event.Event != "" || event.Data != "" {
				return event, nil
			}
			continue
		}

		// Parse field
		if strings.HasPrefix(line, "event:") {
			event.Event = strings.TrimSpace(line[6:])
		} else if strings.HasPrefix(line, "data:") {
			if event.Data != "" {
				event.Data += "\n"
			}
			event.Data += strings.TrimSpace(line[5:])
		}
		// Ignore other fields (id, retry, comments)
	}
}
