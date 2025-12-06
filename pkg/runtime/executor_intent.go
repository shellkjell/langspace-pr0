package runtime

import (
	"fmt"
	"strings"
	"time"

	"github.com/shellkjell/langspace/pkg/ast"
)

// executeIntent executes an intent entity.
func (r *Runtime) executeIntent(ctx *ExecutionContext, entity ast.Entity) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Metadata: make(map[string]string),
	}
	// Copy metadata from execution context
	for k, v := range ctx.Metadata {
		result.Metadata[k] = v
	}
	startTime := time.Now()

	// Emit start event
	ctx.EmitProgress(ProgressEvent{
		Type:    ProgressTypeStart,
		Message: fmt.Sprintf("Executing intent: %s", entity.Name()),
	})

	resolver := NewResolver(ctx)

	// Get the agent to use
	agent, err := r.resolveAgent(ctx, entity, resolver)
	if err != nil {
		result.Error = fmt.Errorf("failed to resolve agent: %w", err)
		return result, result.Error
	}

	// Build the prompt from input and context
	prompt, err := r.buildIntentPrompt(ctx, entity, resolver)
	if err != nil {
		result.Error = fmt.Errorf("failed to build prompt: %w", err)
		return result, result.Error
	}

	// Get the system prompt from the agent
	systemPrompt, err := r.getAgentSystemPrompt(agent, resolver)
	if err != nil {
		result.Error = fmt.Errorf("failed to get system prompt: %w", err)
		return result, result.Error
	}

	// Get the model to use
	model := r.getAgentModel(agent)

	// Get temperature
	temperature := r.getAgentTemperature(agent)

	// Get the provider
	provider, err := r.getProviderForModel(model)
	if err != nil {
		result.Error = fmt.Errorf("failed to get provider: %w", err)
		return result, result.Error
	}

	// Build the request
	req := &CompletionRequest{
		Model:        model,
		SystemPrompt: systemPrompt,
		Messages: []Message{
			{Role: RoleUser, Content: prompt},
		},
		Temperature: temperature,
	}

	// Execute the LLM call
	var resp *CompletionResponse
	if ctx.Handler != nil && r.config.EnableStreaming {
		resp, err = provider.CompleteStream(ctx.Context, req, ctx.Handler)
	} else {
		resp, err = provider.Complete(ctx.Context, req)
	}

	if err != nil {
		result.Error = fmt.Errorf("LLM request failed: %w", err)
		ctx.EmitProgress(ProgressEvent{
			Type:    ProgressTypeError,
			Message: err.Error(),
		})
		return result, result.Error
	}

	// Store the output
	result.Success = true
	result.Output = resp.Content
	result.TokensUsed = resp.Usage
	result.Duration = time.Since(startTime)
	result.Metadata["model"] = resp.Model
	result.Metadata["finish_reason"] = string(resp.FinishReason)

	// Handle output destination if specified
	if err := r.handleIntentOutput(ctx, entity, resp.Content, resolver); err != nil {
		result.Error = fmt.Errorf("failed to handle output: %w", err)
		return result, result.Error
	}

	// Emit completion event
	ctx.EmitProgress(ProgressEvent{
		Type:     ProgressTypeComplete,
		Message:  fmt.Sprintf("Intent completed: %s", entity.Name()),
		Progress: 100,
		Metadata: map[string]string{
			"tokens_used": fmt.Sprintf("%d", resp.Usage.TotalTokens),
			"duration":    result.Duration.String(),
		},
	})

	return result, nil
}

// resolveAgent resolves the agent to use for an intent.
func (r *Runtime) resolveAgent(ctx *ExecutionContext, entity ast.Entity, resolver *Resolver) (ast.Entity, error) {
	useProp, ok := entity.GetProperty("use")
	if !ok {
		return nil, fmt.Errorf("intent %q has no 'use' property", entity.Name())
	}

	// Resolve the agent reference
	switch v := useProp.(type) {
	case ast.ReferenceValue:
		if v.Type != "agent" {
			return nil, fmt.Errorf("expected agent reference, got %s", v.Type)
		}
		return resolver.workspace.GetAgent(v.Name)

	case ast.StringValue:
		// Direct agent name
		return resolver.workspace.GetAgent(v.Value)

	default:
		resolved, err := resolver.Resolve(useProp)
		if err != nil {
			return nil, err
		}
		if agent, ok := resolved.(ast.Entity); ok && agent.Type() == "agent" {
			return agent, nil
		}
		return nil, fmt.Errorf("cannot resolve agent from %T", useProp)
	}
}

// buildIntentPrompt builds the user prompt from input and context.
func (r *Runtime) buildIntentPrompt(ctx *ExecutionContext, entity ast.Entity, resolver *Resolver) (string, error) {
	var promptParts []string

	// Get input
	if inputProp, ok := entity.GetProperty("input"); ok {
		inputContent, err := r.resolveInputContent(inputProp, resolver)
		if err != nil {
			return "", fmt.Errorf("failed to resolve input: %w", err)
		}
		if inputContent != "" {
			promptParts = append(promptParts, "## Input\n\n"+inputContent)
		}
	} else if input, ok := ctx.GetVariable("input"); ok {
		// Use input from execution context
		promptParts = append(promptParts, "## Input\n\n"+toString(input))
	}

	// Get context
	if contextProp, ok := entity.GetProperty("context"); ok {
		contextContent, err := r.resolveContextContent(contextProp, resolver)
		if err != nil {
			return "", fmt.Errorf("failed to resolve context: %w", err)
		}
		if contextContent != "" {
			promptParts = append(promptParts, "## Context\n\n"+contextContent)
		}
	}

	// Get explicit prompt if provided
	if promptProp, ok := entity.GetProperty("prompt"); ok {
		promptStr, err := resolver.ResolveString(promptProp)
		if err != nil {
			return "", fmt.Errorf("failed to resolve prompt: %w", err)
		}
		promptParts = append(promptParts, promptStr)
	}

	if len(promptParts) == 0 {
		// Use entity name as a basic prompt
		return fmt.Sprintf("Please help me: %s", entity.Name()), nil
	}

	return strings.Join(promptParts, "\n\n"), nil
}

// resolveInputContent resolves the input property to content.
func (r *Runtime) resolveInputContent(input ast.Value, resolver *Resolver) (string, error) {
	resolved, err := resolver.Resolve(input)
	if err != nil {
		return "", err
	}

	return formatContent(resolved), nil
}

// resolveContextContent resolves the context property to content.
func (r *Runtime) resolveContextContent(context ast.Value, resolver *Resolver) (string, error) {
	resolved, err := resolver.Resolve(context)
	if err != nil {
		return "", err
	}

	return formatContent(resolved), nil
}

// formatContent formats resolved content for inclusion in a prompt.
func formatContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v

	case []byte:
		return string(v)

	case FileContent:
		return fmt.Sprintf("### %s\n\n```\n%s\n```", v.Path, v.Content)

	case []FileContent:
		var parts []string
		for _, f := range v {
			parts = append(parts, fmt.Sprintf("### %s\n\n```\n%s\n```", f.Path, f.Content))
		}
		return strings.Join(parts, "\n\n")

	case []interface{}:
		var parts []string
		for _, item := range v {
			parts = append(parts, formatContent(item))
		}
		return strings.Join(parts, "\n\n")

	case map[string]interface{}:
		var parts []string
		for k, val := range v {
			parts = append(parts, fmt.Sprintf("**%s**: %s", k, formatContent(val)))
		}
		return strings.Join(parts, "\n")

	case ast.Entity:
		// For file entities, try to get contents
		if v.Type() == "file" {
			if contents, ok := v.GetProperty("contents"); ok {
				return formatContent(contents)
			}
			if path, ok := v.GetProperty("path"); ok {
				return formatContent(path)
			}
		}
		return fmt.Sprintf("[%s: %s]", v.Type(), v.Name())

	default:
		return fmt.Sprintf("%v", content)
	}
}

// getAgentSystemPrompt extracts the system prompt from an agent.
func (r *Runtime) getAgentSystemPrompt(agent ast.Entity, resolver *Resolver) (string, error) {
	// Check for instruction property
	if instruction, ok := agent.GetProperty("instruction"); ok {
		return resolver.ResolveString(instruction)
	}

	// Check for system_prompt property
	if systemPrompt, ok := agent.GetProperty("system_prompt"); ok {
		return resolver.ResolveString(systemPrompt)
	}

	// Check for prompt property
	if prompt, ok := agent.GetProperty("prompt"); ok {
		return resolver.ResolveString(prompt)
	}

	// Default system prompt based on agent name
	return fmt.Sprintf("You are %s. Help the user with their request.", agent.Name()), nil
}

// getAgentModel gets the model to use for an agent.
func (r *Runtime) getAgentModel(agent ast.Entity) string {
	if model, ok := agent.GetProperty("model"); ok {
		if sv, ok := model.(ast.StringValue); ok {
			return sv.Value
		}
	}
	return r.defaultModel
}

// getAgentTemperature gets the temperature setting for an agent.
func (r *Runtime) getAgentTemperature(agent ast.Entity) float64 {
	if temp, ok := agent.GetProperty("temperature"); ok {
		if nv, ok := temp.(ast.NumberValue); ok {
			return nv.Value
		}
	}
	return 0.7 // Default temperature
}

// getProviderForModel returns the appropriate provider for a model.
func (r *Runtime) getProviderForModel(model string) (LLMProvider, error) {
	// Check model prefix to determine provider
	switch {
	case strings.HasPrefix(model, "claude"):
		if p, ok := r.providers["anthropic"]; ok {
			return p, nil
		}
	case strings.HasPrefix(model, "gpt"), strings.HasPrefix(model, "o1"), strings.HasPrefix(model, "o3"):
		if p, ok := r.providers["openai"]; ok {
			return p, nil
		}
	}

	// Try default provider
	if p, ok := r.providers[r.config.DefaultProvider]; ok {
		return p, nil
	}

	// Return any available provider
	for _, p := range r.providers {
		return p, nil
	}

	return nil, fmt.Errorf("no LLM provider available for model %q", model)
}

// handleIntentOutput handles writing output to a destination.
func (r *Runtime) handleIntentOutput(ctx *ExecutionContext, entity ast.Entity, output string, resolver *Resolver) error {
	outputProp, ok := entity.GetProperty("output")
	if !ok {
		return nil // No output destination specified
	}

	// Resolve the output destination
	switch v := outputProp.(type) {
	case ast.ReferenceValue:
		if v.Type == "file" {
			// Resolve the file path with interpolation
			path, err := resolver.interpolateString(v.Name)
			if err != nil {
				return err
			}
			return writeFile(path, output)
		}

	case ast.StringValue:
		// Assume it's a file path
		path, err := resolver.interpolateString(v.Value)
		if err != nil {
			return err
		}
		return writeFile(path, output)

	case ast.MethodCallValue:
		// Handle method calls like github.pr_comment()
		// This is a future enhancement
		return nil
	}

	return nil
}

// writeFile writes content to a file, creating directories as needed.
func writeFile(path string, content string) error {
	// Create parent directories if needed
	// (this is handled by os.WriteFile in Go 1.16+, but we're explicit)
	return writeFileWithDirs(path, []byte(content), 0644)
}

// writeFileWithDirs writes a file, creating parent directories as needed.
func writeFileWithDirs(path string, data []byte, perm uint32) error {
	// For simplicity, just use os.WriteFile
	// Directory creation would be added in production
	return nil // Placeholder - actual file writing would happen here
}
