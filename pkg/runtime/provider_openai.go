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

// OpenAIProvider implements LLMProvider for the OpenAI API.
type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// OpenAIOption is a functional option for configuring OpenAIProvider.
type OpenAIOption func(*OpenAIProvider)

// WithOpenAIAPIKey sets the API key.
func WithOpenAIAPIKey(key string) OpenAIOption {
	return func(p *OpenAIProvider) {
		p.apiKey = key
	}
}

// WithOpenAIBaseURL sets a custom base URL (e.g., for Azure OpenAI).
func WithOpenAIBaseURL(url string) OpenAIOption {
	return func(p *OpenAIProvider) {
		p.baseURL = url
	}
}

// WithOpenAIHTTPClient sets a custom HTTP client.
func WithOpenAIHTTPClient(client *http.Client) OpenAIOption {
	return func(p *OpenAIProvider) {
		p.httpClient = client
	}
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(opts ...OpenAIOption) *OpenAIProvider {
	p := &OpenAIProvider{
		baseURL:    "https://api.openai.com",
		httpClient: http.DefaultClient,
	}

	// Check for API key in environment
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		p.apiKey = key
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

// openaiRequest is the request format for OpenAI's API.
type openaiRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Tools       []openaiTool    `json:"tools,omitempty"`
}

type openaiMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openaiTool struct {
	Type     string         `json:"type"`
	Function openaiFunction `json:"function"`
}

type openaiFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type openaiToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// openaiResponse is the response format from OpenAI's API.
type openaiResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int           `json:"index"`
		Message      openaiMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (p *OpenAIProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("openai API key not set")
	}

	// Convert messages to OpenAI format
	openaiMsgs := make([]openaiMessage, 0, len(req.Messages)+1)

	// Add system message if provided
	if req.SystemPrompt != "" {
		openaiMsgs = append(openaiMsgs, openaiMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	for _, msg := range req.Messages {
		oaiMsg := openaiMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}

		if msg.ToolCallID != "" {
			oaiMsg.ToolCallID = msg.ToolCallID
		}

		// Convert tool calls
		for _, tc := range msg.ToolCalls {
			argsJSON, _ := json.Marshal(tc.Arguments)
			oaiMsg.ToolCalls = append(oaiMsg.ToolCalls, openaiToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      tc.Name,
					Arguments: string(argsJSON),
				},
			})
		}

		openaiMsgs = append(openaiMsgs, oaiMsg)
	}

	// Convert tools to OpenAI format
	var openaiTools []openaiTool
	for _, tool := range req.Tools {
		openaiTools = append(openaiTools, openaiTool{
			Type: "function",
			Function: openaiFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		})
	}

	openaiReq := openaiRequest{
		Model:       req.Model,
		Messages:    openaiMsgs,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Tools:       openaiTools,
	}

	body, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var openaiResp openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return p.convertResponse(&openaiResp), nil
}

func (p *OpenAIProvider) convertResponse(resp *openaiResponse) *CompletionResponse {
	result := &CompletionResponse{
		Model: resp.Model,
		Usage: TokenUsage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
	}

	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		result.Content = choice.Message.Content

		// Convert tool calls
		for _, tc := range choice.Message.ToolCalls {
			var args map[string]interface{}
			_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
			result.ToolCalls = append(result.ToolCalls, ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: args,
			})
		}

		// Convert finish reason
		switch choice.FinishReason {
		case "stop":
			result.FinishReason = FinishReasonStop
		case "length":
			result.FinishReason = FinishReasonLength
		case "tool_calls":
			result.FinishReason = FinishReasonToolUse
		default:
			result.FinishReason = FinishReasonStop
		}
	}

	return result
}

func (p *OpenAIProvider) CompleteStream(ctx context.Context, req *CompletionRequest, handler StreamHandler) (*CompletionResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("openai API key not set")
	}

	// Convert messages to OpenAI format
	openaiMsgs := make([]openaiMessage, 0, len(req.Messages)+1)

	if req.SystemPrompt != "" {
		openaiMsgs = append(openaiMsgs, openaiMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	for _, msg := range req.Messages {
		openaiMsgs = append(openaiMsgs, openaiMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	openaiReq := openaiRequest{
		Model:       req.Model,
		Messages:    openaiMsgs,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      true,
	}

	body, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

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

func (p *OpenAIProvider) handleStream(body io.Reader, handler StreamHandler) (*CompletionResponse, error) {
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

		if event.Data == "[DONE]" {
			break
		}

		var chunk struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Model   string `json:"model"`
			Choices []struct {
				Index int `json:"index"`
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(event.Data), &chunk); err != nil {
			continue
		}

		result.Model = chunk.Model

		if len(chunk.Choices) > 0 {
			choice := chunk.Choices[0]
			if choice.Delta.Content != "" {
				contentBuilder.WriteString(choice.Delta.Content)
				handler.OnChunk(StreamChunk{
					Content: choice.Delta.Content,
					Type:    ChunkTypeContent,
					Index:   chunkIndex,
				})
				chunkIndex++
			}

			if choice.FinishReason != "" {
				switch choice.FinishReason {
				case "stop":
					result.FinishReason = FinishReasonStop
				case "length":
					result.FinishReason = FinishReasonLength
				case "tool_calls":
					result.FinishReason = FinishReasonToolUse
				}
			}
		}
	}

	result.Content = contentBuilder.String()
	handler.OnComplete(result)
	return result, nil
}

func (p *OpenAIProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("openai API key not set")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var listResp struct {
		Data []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]ModelInfo, 0, len(listResp.Data))
	for _, m := range listResp.Data {
		models = append(models, ModelInfo{
			ID:       m.ID,
			Name:     m.ID,
			Provider: "openai",
		})
	}

	return models, nil
}
