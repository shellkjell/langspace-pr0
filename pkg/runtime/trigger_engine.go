package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shellkjell/langspace/pkg/ast"
)

// TriggerEngine manages and executes triggers.
type TriggerEngine struct {
	runtime *Runtime
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.RWMutex
	active  bool
}

// NewTriggerEngine creates a new trigger engine.
func NewTriggerEngine(r *Runtime) *TriggerEngine {
	return &TriggerEngine{
		runtime: r,
	}
}

// Start starts the trigger engine.
func (e *TriggerEngine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.active {
		return fmt.Errorf("trigger engine already active")
	}

	e.ctx, e.cancel = context.WithCancel(ctx)
	e.active = true

	go e.run()

	return nil
}

// Stop stops the trigger engine.
func (e *TriggerEngine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.active {
		return nil
	}

	e.cancel()
	e.active = false

	return nil
}

// run is the main loop for the trigger engine.
func (e *TriggerEngine) run() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.checkTriggers()
		}
	}
}

// checkTriggers checks all triggers in the workspace and executes those that match.
func (e *TriggerEngine) checkTriggers() {
	triggers := e.runtime.workspace.GetEntitiesByType("trigger")
	
	for _, t := range triggers {
		if schedule, ok := t.GetProperty("schedule"); ok {
			if e.shouldRunSchedule(toString(schedule)) {
				go e.executeTrigger(t)
			}
		}
	}
}

// shouldRunSchedule checks if a cron-like schedule should run now.
func (e *TriggerEngine) shouldRunSchedule(schedule string) bool {
	// Simple implementation: if schedule is "* * * * *", it runs every minute.
	// In production, use a proper cron parser.
	return schedule == "* * * * *"
}

// executeTrigger executes the action associated with a trigger.
func (e *TriggerEngine) executeTrigger(trigger ast.Entity) {
	// Get the action to execute (intent or pipeline)
	actionProp, ok := trigger.GetProperty("use")
	if !ok {
		return
	}

	var entityType, entityName string
	if ref, ok := actionProp.(ast.ReferenceValue); ok {
		entityType = ref.Type
		entityName = ref.Name
	} else {
		return
	}

	_, err := e.runtime.ExecuteByName(context.Background(), entityType, entityName)
	if err != nil {
		fmt.Printf("Trigger execution failed: %v\n", err)
	}
}
