package workspace

import (
	"fmt"
	"sync"

	"github.com/shellkjell/langspace/pkg/ast"
	"github.com/shellkjell/langspace/pkg/validator"
)

// HookType defines the type of lifecycle hook
type HookType string

const (
	// HookBeforeAdd is called before an entity is added
	HookBeforeAdd HookType = "before_add"
	// HookAfterAdd is called after an entity is added
	HookAfterAdd HookType = "after_add"
	// HookBeforeRemove is called before an entity is removed
	HookBeforeRemove HookType = "before_remove"
	// HookAfterRemove is called after an entity is removed
	HookAfterRemove HookType = "after_remove"
)

// EntityHook is a function called during entity lifecycle events
// Return an error to cancel the operation (for "before" hooks)
type EntityHook func(entity ast.Entity) error

// RelationType defines the type of relationship between entities
type RelationType string

const (
	// RelationTypeAssigned indicates an agent is assigned to work on a file or task
	RelationTypeAssigned RelationType = "assigned"
	// RelationTypeDepends indicates a task depends on another entity
	RelationTypeDepends RelationType = "depends"
	// RelationTypeProduces indicates an entity produces another entity
	RelationTypeProduces RelationType = "produces"
	// RelationTypeConsumes indicates an entity consumes/reads another entity
	RelationTypeConsumes RelationType = "consumes"
)

// Relationship represents a connection between two entities
type Relationship struct {
	SourceType string       // Type of the source entity (file, agent, task)
	SourceName string       // Name/identifier of the source entity
	TargetType string       // Type of the target entity
	TargetName string       // Name/identifier of the target entity
	Type       RelationType // Type of relationship
}

// Workspace represents a virtual workspace containing entities
type Workspace struct {
	entities      []ast.Entity
	relationships []Relationship
	hooks         map[HookType][]EntityHook
	mu            sync.RWMutex
	validator     validator.EntityValidator
}

// New creates a new Workspace instance
func New() *Workspace {
	return &Workspace{
		entities:      make([]ast.Entity, 0),
		relationships: make([]Relationship, 0),
		hooks:         make(map[HookType][]EntityHook),
	}
}

// WorkspaceStats contains statistics about the workspace
type WorkspaceStats struct {
	TotalEntities      int
	FileEntities       int
	AgentEntities      int
	ToolEntities       int
	IntentEntities     int
	PipelineEntities   int
	TotalRelationships int
	TotalHooks         int
	HasValidator       bool
}

// Stat returns statistics about the workspace
func (w *Workspace) Stat() WorkspaceStats {
	w.mu.RLock()
	defer w.mu.RUnlock()

	totalHooks := 0
	for _, hooks := range w.hooks {
		totalHooks += len(hooks)
	}

	stats := WorkspaceStats{
		TotalEntities:      len(w.entities),
		TotalRelationships: len(w.relationships),
		TotalHooks:         totalHooks,
		HasValidator:       w.validator != nil,
	}

	// Count entities by type
	for _, entity := range w.entities {
		switch entity.Type() {
		case "file":
			stats.FileEntities++
		case "agent":
			stats.AgentEntities++
		case "tool":
			stats.ToolEntities++
		case "intent":
			stats.IntentEntities++
		case "pipeline":
			stats.PipelineEntities++
		}
	}

	return stats
}

// WithValidator sets a validator for the workspace
func (w *Workspace) WithValidator(v validator.EntityValidator) *Workspace {
	w.validator = v
	return w
}

// OnEntityEvent registers a hook for a specific entity lifecycle event
func (w *Workspace) OnEntityEvent(hookType HookType, hook EntityHook) *Workspace {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.hooks[hookType] = append(w.hooks[hookType], hook)
	return w
}

// runHooks executes all hooks of a given type for an entity
func (w *Workspace) runHooks(hookType HookType, entity ast.Entity) error {
	hooks := w.hooks[hookType]
	for _, hook := range hooks {
		if err := hook(entity); err != nil {
			return fmt.Errorf("hook %s failed: %w", hookType, err)
		}
	}
	return nil
}

// AddEntity adds an entity to the workspace
func (w *Workspace) AddEntity(entity ast.Entity) error {
	if entity == nil {
		return fmt.Errorf("cannot add nil entity")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Run before-add hooks
	if err := w.runHooks(HookBeforeAdd, entity); err != nil {
		return err
	}

	if w.validator != nil {
		if err := w.validator.ValidateEntity(entity); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	w.entities = append(w.entities, entity)

	// Run after-add hooks (errors are logged but don't fail the operation)
	_ = w.runHooks(HookAfterAdd, entity)

	return nil
}

// GetEntities returns all entities in the workspace
func (w *Workspace) GetEntities() []ast.Entity {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Return a copy to prevent external modifications
	result := make([]ast.Entity, len(w.entities))
	copy(result, w.entities)
	return result
}

// GetEntitiesByType returns all entities of a specific type
func (w *Workspace) GetEntitiesByType(entityType string) []ast.Entity {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var result []ast.Entity
	for _, entity := range w.entities {
		if entity.Type() == entityType {
			result = append(result, entity)
		}
	}
	return result
}

// GetEntityByName returns an entity by type and name
func (w *Workspace) GetEntityByName(entityType, entityName string) (ast.Entity, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, entity := range w.entities {
		if entity.Type() == entityType && entity.Name() == entityName {
			return entity, true
		}
	}
	return nil, false
}

// RemoveEntity removes an entity from the workspace by type and name
func (w *Workspace) RemoveEntity(entityType, entityName string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i, entity := range w.entities {
		if entity.Type() == entityType && entity.Name() == entityName {
			// Run before-remove hooks
			if err := w.runHooks(HookBeforeRemove, entity); err != nil {
				return err
			}

			// Remove the entity
			w.entities = append(w.entities[:i], w.entities[i+1:]...)

			// Remove any relationships involving this entity
			w.removeRelationshipsForEntity(entityType, entityName)

			// Run after-remove hooks
			_ = w.runHooks(HookAfterRemove, entity)

			return nil
		}
	}

	return fmt.Errorf("entity not found: %s %q", entityType, entityName)
}

// removeRelationshipsForEntity removes all relationships involving the specified entity
// Must be called with lock held
func (w *Workspace) removeRelationshipsForEntity(entityType, entityName string) {
	var remaining []Relationship
	for _, rel := range w.relationships {
		if (rel.SourceType == entityType && rel.SourceName == entityName) ||
			(rel.TargetType == entityType && rel.TargetName == entityName) {
			continue
		}
		remaining = append(remaining, rel)
	}
	w.relationships = remaining
}

// Clear removes all entities from the workspace
func (w *Workspace) Clear() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.entities = make([]ast.Entity, 0)
	w.relationships = make([]Relationship, 0)
}

// AddRelationship creates a relationship between two entities
func (w *Workspace) AddRelationship(sourceType, sourceName, targetType, targetName string, relType RelationType) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Validate that source entity exists
	sourceFound := false
	for _, entity := range w.entities {
		if entity.Type() == sourceType && entity.Name() == sourceName {
			sourceFound = true
			break
		}
	}
	if !sourceFound {
		return fmt.Errorf("source entity not found: %s %q", sourceType, sourceName)
	}

	// Validate that target entity exists
	targetFound := false
	for _, entity := range w.entities {
		if entity.Type() == targetType && entity.Name() == targetName {
			targetFound = true
			break
		}
	}
	if !targetFound {
		return fmt.Errorf("target entity not found: %s %q", targetType, targetName)
	}

	// Check for duplicate relationship
	for _, rel := range w.relationships {
		if rel.SourceType == sourceType && rel.SourceName == sourceName &&
			rel.TargetType == targetType && rel.TargetName == targetName &&
			rel.Type == relType {
			return fmt.Errorf("relationship already exists")
		}
	}

	w.relationships = append(w.relationships, Relationship{
		SourceType: sourceType,
		SourceName: sourceName,
		TargetType: targetType,
		TargetName: targetName,
		Type:       relType,
	})

	return nil
}

// GetRelationships returns all relationships in the workspace
func (w *Workspace) GetRelationships() []Relationship {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make([]Relationship, len(w.relationships))
	copy(result, w.relationships)
	return result
}

// GetRelationshipsForEntity returns all relationships involving the specified entity
func (w *Workspace) GetRelationshipsForEntity(entityType, entityName string) []Relationship {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var result []Relationship
	for _, rel := range w.relationships {
		if (rel.SourceType == entityType && rel.SourceName == entityName) ||
			(rel.TargetType == entityType && rel.TargetName == entityName) {
			result = append(result, rel)
		}
	}
	return result
}

// GetRelatedEntities returns entities related to the specified entity
func (w *Workspace) GetRelatedEntities(entityType, entityName string, relType RelationType) []ast.Entity {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var result []ast.Entity
	for _, rel := range w.relationships {
		if rel.Type != relType {
			continue
		}

		var targetType, targetName string
		if rel.SourceType == entityType && rel.SourceName == entityName {
			targetType = rel.TargetType
			targetName = rel.TargetName
		} else if rel.TargetType == entityType && rel.TargetName == entityName {
			targetType = rel.SourceType
			targetName = rel.SourceName
		} else {
			continue
		}

		for _, entity := range w.entities {
			if entity.Type() == targetType && entity.Name() == targetName {
				result = append(result, entity)
				break
			}
		}
	}

	return result
}

// RemoveRelationship removes a specific relationship
func (w *Workspace) RemoveRelationship(sourceType, sourceName, targetType, targetName string, relType RelationType) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i, rel := range w.relationships {
		if rel.SourceType == sourceType && rel.SourceName == sourceName &&
			rel.TargetType == targetType && rel.TargetName == targetName &&
			rel.Type == relType {
			w.relationships = append(w.relationships[:i], w.relationships[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("relationship not found")
}
