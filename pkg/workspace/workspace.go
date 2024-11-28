package workspace

import (
	"fmt"
	"sync"

	"github.com/shellkjell/langspace/pkg/ast"
)

// Workspace represents a virtual workspace containing entities
type Workspace struct {
	entities  []ast.Entity
	mu        sync.RWMutex
	validator Validator
}

// Validator interface defines entity validation behavior
type Validator interface {
	ValidateEntity(entity ast.Entity) error
}

// New creates a new Workspace instance
func New() *Workspace {
	return &Workspace{
		entities: make([]ast.Entity, 0),
	}
}

// WorkspaceStats contains statistics about the workspace
type WorkspaceStats struct {
	TotalEntities int
	FileEntities  int
	AgentEntities int
	HasValidator  bool
}

// Stat returns statistics about the workspace
func (w *Workspace) Stat() WorkspaceStats {

	w.mu.RLock()
	defer w.mu.RUnlock()

	stats := WorkspaceStats{
		TotalEntities: len(w.entities),
		HasValidator:  w.validator != nil,
	}

	// Count entities by type
	for _, entity := range w.entities {
		switch entity.Type() {
		case "file":
			stats.FileEntities++
		case "agent":
			stats.AgentEntities++
		}
	}

	return stats
}

// WithValidator sets a validator for the workspace
func (w *Workspace) WithValidator(v Validator) *Workspace {
	w.validator = v
	return w
}

// AddEntity adds an entity to the workspace
func (w *Workspace) AddEntity(entity ast.Entity) error {
	if entity == nil {
		return fmt.Errorf("cannot add nil entity")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.validator != nil {
		if err := w.validator.ValidateEntity(entity); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	} else {
		// Validate entity based on type
		switch entity.Type() {
		case "file":
			if len(entity.Properties()) != 2 {
				return fmt.Errorf("file entity must have exactly 2 properties (name and contents)")
			}
		case "agent":
			if len(entity.Properties()) != 2 {
				return fmt.Errorf("agent entity must have exactly 2 properties (type and instruction)")
			}
		default:
			return fmt.Errorf("unknown entity type: %s", entity.Type())
		}
	}

	w.entities = append(w.entities, entity)
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

// Clear removes all entities from the workspace
func (w *Workspace) Clear() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.entities = make([]ast.Entity, 0)
}
