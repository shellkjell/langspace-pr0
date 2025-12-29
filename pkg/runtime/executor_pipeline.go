package runtime

import (
	"fmt"
	"sync"
	"time"

	"github.com/shellkjell/langspace/pkg/ast"
)

// executePipeline executes a pipeline entity.
func (r *Runtime) executePipeline(ctx *ExecutionContext, entity ast.Entity) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Metadata:    make(map[string]string),
		StepResults: make(map[string]*StepResult),
	}
	startTime := time.Now()

	// Initialize step outputs map
	ctx.StepOutputs = make(map[string]interface{})

	// Emit start event
	ctx.EmitProgress(ProgressEvent{
		Type:    ProgressTypeStart,
		Message: fmt.Sprintf("Executing pipeline: %s", entity.Name()),
	})

	resolver := NewResolver(ctx)

	// Get steps from the pipeline
	pipeline, ok := entity.(*ast.PipelineEntity)
	if !ok {
		return nil, fmt.Errorf("entity is not a pipeline")
	}

	// Execute each step
	totalSteps := len(pipeline.Steps)
	for i, step := range pipeline.Steps {
		stepResult, err := r.executeStep(ctx, step, resolver, i+1, totalSteps)
		result.StepResults[step.Name()] = stepResult

		if err != nil {
			result.Error = fmt.Errorf("step %q failed: %w", step.Name(), err)
			ctx.EmitProgress(ProgressEvent{
				Type:    ProgressTypeError,
				Message: err.Error(),
				Step:    step.Name(),
			})
			return result, result.Error
		}

		// Update token usage
		if stepResult.Output != nil {
			if usage, ok := stepResult.Output.(TokenUsage); ok {
				result.TokensUsed.Add(usage)
			}
		}
	}

	// Handle parallel blocks in properties
	for key, value := range entity.Properties() {
		if key == "parallel" {
			if nested, ok := value.(ast.NestedEntityValue); ok {
				if err := r.executeParallelBlock(ctx, nested.Entity, resolver, result); err != nil {
					result.Error = err
					r.handleLifecycleEvent(ctx, entity, "on_failure", resolver)
					r.handleLifecycleEvent(ctx, entity, "on_complete", resolver)
					return result, err
				}
			}
		}

		// Handle branch blocks
		if key == "branch" {
			if branchVal, ok := value.(ast.BranchValue); ok {
				if err := r.executeBranchBlock(ctx, branchVal, resolver, result); err != nil {
					result.Error = err
					r.handleLifecycleEvent(ctx, entity, "on_failure", resolver)
					r.handleLifecycleEvent(ctx, entity, "on_complete", resolver)
					return result, err
				}
			}
		}

		// Handle loop blocks
		if key == "loop" {
			if loopVal, ok := value.(ast.LoopValue); ok {
				if err := r.executeLoopBlock(ctx, loopVal, resolver, result); err != nil {
					result.Error = err
					r.handleLifecycleEvent(ctx, entity, "on_failure", resolver)
					r.handleLifecycleEvent(ctx, entity, "on_complete", resolver)
					return result, err
				}
			}
		}
	}

	// Get the final output
	if outputProp, ok := entity.GetProperty("output"); ok {
		output, err := resolver.Resolve(outputProp)
		if err != nil {
			result.Error = fmt.Errorf("failed to resolve output: %w", err)
			r.handleLifecycleEvent(ctx, entity, "on_failure", resolver)
			r.handleLifecycleEvent(ctx, entity, "on_complete", resolver)
			return result, result.Error
		}
		result.Output = output
	} else if len(pipeline.Steps) > 0 {
		// Default to the last step's output
		lastStep := pipeline.Steps[len(pipeline.Steps)-1]
		if output, ok := ctx.GetStepOutput(lastStep.Name()); ok {
			result.Output = output
		}
	}

	result.Success = true
	result.Duration = time.Since(startTime)

	// Handle success lifecycle events
	r.handleLifecycleEvent(ctx, entity, "on_success", resolver)
	r.handleLifecycleEvent(ctx, entity, "on_complete", resolver)

	// Emit completion event
	ctx.EmitProgress(ProgressEvent{
		Type:     ProgressTypeComplete,
		Message:  fmt.Sprintf("Pipeline completed: %s", entity.Name()),
		Progress: 100,
		Metadata: map[string]string{
			"steps_executed": fmt.Sprintf("%d", len(result.StepResults)),
			"duration":       result.Duration.String(),
		},
	})

	return result, nil
}

// executeStep executes a single step in a pipeline.
func (r *Runtime) executeStep(ctx *ExecutionContext, step *ast.StepEntity, resolver *Resolver, stepNum, totalSteps int) (*StepResult, error) {
	stepResult := &StepResult{
		Name:      step.Name(),
		StartTime: time.Now(),
	}

	// Emit step start event
	progress := (stepNum * 100) / (totalSteps + 1)
	ctx.EmitProgress(ProgressEvent{
		Type:     ProgressTypeStep,
		Message:  fmt.Sprintf("Executing step: %s", step.Name()),
		Step:     step.Name(),
		Progress: progress,
	})

	// Get the agent to use
	agent, err := r.resolveStepAgent(ctx, step, resolver)
	if err != nil {
		stepResult.Error = err
		stepResult.EndTime = time.Now()
		stepResult.Duration = stepResult.EndTime.Sub(stepResult.StartTime)
		return stepResult, err
	}

	// Build the prompt for this step
	prompt, err := r.buildStepPrompt(ctx, step, resolver)
	if err != nil {
		stepResult.Error = err
		stepResult.EndTime = time.Now()
		stepResult.Duration = stepResult.EndTime.Sub(stepResult.StartTime)
		return stepResult, err
	}

	// Get system prompt from agent
	systemPrompt, err := r.getAgentSystemPrompt(agent, resolver)
	if err != nil {
		stepResult.Error = err
		stepResult.EndTime = time.Now()
		stepResult.Duration = stepResult.EndTime.Sub(stepResult.StartTime)
		return stepResult, err
	}

	// Add step-specific instruction if provided
	if instruction, ok := step.GetProperty("instruction"); ok {
		instructionStr, err := resolver.ResolveString(instruction)
		if err == nil && instructionStr != "" {
			systemPrompt += "\n\n" + instructionStr
		}
	}

	// Get model and temperature
	model := r.getAgentModel(agent)
	temperature := r.getAgentTemperature(agent)

	// Get provider
	provider, err := r.getProviderForModel(model)
	if err != nil {
		stepResult.Error = err
		stepResult.EndTime = time.Now()
		stepResult.Duration = stepResult.EndTime.Sub(stepResult.StartTime)
		return stepResult, err
	}

	// Build request
	req := &CompletionRequest{
		Model:        model,
		SystemPrompt: systemPrompt,
		Messages: []Message{
			{Role: RoleUser, Content: prompt},
		},
		Temperature: temperature,
	}

	// Execute
	var resp *CompletionResponse
	if ctx.Handler != nil && r.config.EnableStreaming {
		resp, err = provider.CompleteStream(ctx.Context, req, ctx.Handler)
	} else {
		resp, err = provider.Complete(ctx.Context, req)
	}

	stepResult.EndTime = time.Now()
	stepResult.Duration = stepResult.EndTime.Sub(stepResult.StartTime)

	if err != nil {
		stepResult.Error = err
		return stepResult, err
	}

	// Store the step output
	stepResult.Success = true
	stepResult.Output = resp.Content
	ctx.SetStepOutput(step.Name(), resp.Content)

	// Also store in a structured format for property access
	ctx.SetStepOutput(step.Name()+".output", resp.Content)
	ctx.SetStepOutput(step.Name()+".tokens", resp.Usage)

	return stepResult, nil
}

// resolveStepAgent resolves the agent for a step.
func (r *Runtime) resolveStepAgent(ctx *ExecutionContext, step *ast.StepEntity, resolver *Resolver) (ast.Entity, error) {
	useProp, ok := step.GetProperty("use")
	if !ok {
		return nil, fmt.Errorf("step %q has no 'use' property", step.Name())
	}

	switch v := useProp.(type) {
	case ast.ReferenceValue:
		if v.Type != "agent" {
			return nil, fmt.Errorf("expected agent reference, got %s", v.Type)
		}
		return resolver.workspace.GetAgent(v.Name)

	case ast.StringValue:
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

// buildStepPrompt builds the prompt for a pipeline step.
func (r *Runtime) buildStepPrompt(ctx *ExecutionContext, step *ast.StepEntity, resolver *Resolver) (string, error) {
	var promptParts []string

	// Get input
	if inputProp, ok := step.GetProperty("input"); ok {
		inputContent, err := r.resolveStepInput(ctx, inputProp, resolver)
		if err != nil {
			return "", fmt.Errorf("failed to resolve input: %w", err)
		}
		if inputContent != "" {
			promptParts = append(promptParts, "## Input\n\n"+inputContent)
		}
	}

	// Get context
	if contextProp, ok := step.GetProperty("context"); ok {
		contextContent, err := r.resolveContextContent(contextProp, resolver)
		if err != nil {
			return "", fmt.Errorf("failed to resolve context: %w", err)
		}
		if contextContent != "" {
			promptParts = append(promptParts, "## Context\n\n"+contextContent)
		}
	}

	// Get explicit prompt if provided
	if promptProp, ok := step.GetProperty("prompt"); ok {
		promptStr, err := resolver.ResolveString(promptProp)
		if err != nil {
			return "", fmt.Errorf("failed to resolve prompt: %w", err)
		}
		promptParts = append(promptParts, promptStr)
	}

	if len(promptParts) == 0 {
		return fmt.Sprintf("Please help me with step: %s", step.Name()), nil
	}

	return joinNonEmpty(promptParts, "\n\n"), nil
}

// resolveStepInput resolves the input for a step, which may reference previous step outputs.
func (r *Runtime) resolveStepInput(ctx *ExecutionContext, input ast.Value, resolver *Resolver) (string, error) {
	resolved, err := resolver.Resolve(input)
	if err != nil {
		return "", err
	}
	return formatContent(resolved), nil
}

// executeParallelBlock executes entities in parallel.
func (r *Runtime) executeParallelBlock(ctx *ExecutionContext, entity ast.Entity, resolver *Resolver, result *ExecutionResult) error {
	var entities []ast.Entity

	// If it's a ParallelEntity, use its steps
	if parallel, ok := entity.(*ast.ParallelEntity); ok {
		for _, step := range parallel.Steps {
			entities = append(entities, step)
		}
	} else {
		// Fallback: Collect all nested entities from properties
		for _, value := range entity.Properties() {
			if nested, ok := value.(ast.NestedEntityValue); ok {
				entities = append(entities, nested.Entity)
			}
		}
	}

	if len(entities) == 0 {
		return nil
	}

	// Execute entities in parallel
	var wg sync.WaitGroup
	execResults := make([]*ExecutionResult, len(entities))
	errors := make([]error, len(entities))

	for i, ent := range entities {
		wg.Add(1)
		go func(idx int, e ast.Entity) {
			defer wg.Done()
			res, err := r.Execute(ctx.Context, e)
			execResults[idx] = res
			errors[idx] = err
		}(i, ent)
	}

	wg.Wait()

	// Collect results
	for i, ent := range entities {
		if errors[i] != nil {
			return fmt.Errorf("parallel entity %q failed: %w", ent.Name(), errors[i])
		}

		// If it was a step, add to step results
		if stepEntity, ok := ent.(*ast.StepEntity); ok {
			result.StepResults[stepEntity.Name()] = &StepResult{
				Name:     stepEntity.Name(),
				Success:  execResults[i].Success,
				Output:   execResults[i].Output,
				Duration: execResults[i].Duration,
			}
		}
	}

	return nil
}

// executeBranchBlock executes a branch based on a condition.
func (r *Runtime) executeBranchBlock(ctx *ExecutionContext, branch ast.BranchValue, resolver *Resolver, result *ExecutionResult) error {
	// Resolve the condition
	conditionValue, err := resolver.Resolve(branch.Condition)
	if err != nil {
		return fmt.Errorf("failed to resolve branch condition: %w", err)
	}

	conditionStr := toString(conditionValue)

	// Find the matching case
	caseEntity, ok := branch.Cases[conditionStr]
	if !ok {
		// Check for default case
		caseEntity, ok = branch.Cases["default"]
		if !ok {
			// No matching case, skip
			return nil
		}
	}

	// Execute the matched case
	execResult, err := r.Execute(ctx.Context, caseEntity.Entity)
	if err != nil {
		return fmt.Errorf("branch case %q failed: %w", conditionStr, err)
	}

	// Store branch output
	ctx.SetVariable("branch", map[string]interface{}{
		"output":  execResult.Output,
		"case":    conditionStr,
		"success": execResult.Success,
	})

	// If it was a step, add to step results
	if stepEntity, ok := caseEntity.Entity.(*ast.StepEntity); ok {
		result.StepResults[stepEntity.Name()] = &StepResult{
			Name:     stepEntity.Name(),
			Success:  execResult.Success,
			Output:   execResult.Output,
			Duration: execResult.Duration,
		}
	}

	return nil
}

// executeLoopBlock executes a loop.
func (r *Runtime) executeLoopBlock(ctx *ExecutionContext, loop ast.LoopValue, resolver *Resolver, result *ExecutionResult) error {
	maxIterations := loop.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 10 // Default maximum
	}

	// Initialize current variable
	if _, ok := ctx.GetVariable("current"); !ok {
		// Use input as initial current value
		if input, ok := ctx.GetVariable("input"); ok {
			ctx.SetVariable("current", input)
		}
	}

	for i := 0; i < maxIterations; i++ {
		ctx.SetVariable("iteration", i+1)

		// Execute loop body entities
		for _, nestedEntity := range loop.Body {
			execResult, err := r.Execute(ctx.Context, nestedEntity.Entity)
			if err != nil {
				return fmt.Errorf("loop iteration %d, entity %q failed: %w", i+1, nestedEntity.Entity.Name(), err)
			}

			// If it was a step, add to step results with iteration suffix
			if stepEntity, ok := nestedEntity.Entity.(*ast.StepEntity); ok {
				stepName := fmt.Sprintf("%s_iter%d", stepEntity.Name(), i+1)
				result.StepResults[stepName] = &StepResult{
					Name:     stepName,
					Success:  execResult.Success,
					Output:   execResult.Output,
					Duration: execResult.Duration,
				}
			}
		}

		// Check break condition
		if loop.BreakCondition != nil {
			breakValue, err := resolver.Resolve(loop.BreakCondition)
			if err != nil {
				return fmt.Errorf("failed to evaluate break condition: %w", err)
			}

			if shouldBreak, ok := breakValue.(bool); ok && shouldBreak {
				break
			}
		}
	}

	return nil
}

// joinNonEmpty joins non-empty strings with a separator.
func joinNonEmpty(parts []string, sep string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	if len(nonEmpty) == 0 {
		return ""
	}
	result := nonEmpty[0]
	for i := 1; i < len(nonEmpty); i++ {
		result += sep + nonEmpty[i]
	}
	return result
}
