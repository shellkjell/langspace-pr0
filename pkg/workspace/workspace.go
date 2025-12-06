package workspace

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/shellkjell/langspace/pkg/ast"
	"github.com/shellkjell/langspace/pkg/slices"
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
	// HookBeforeUpdate is called before an entity is updated
	HookBeforeUpdate HookType = "before_update"
	// HookAfterUpdate is called after an entity is updated
	HookAfterUpdate HookType = "after_update"
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

// EventType defines the type of workspace event
type EventType string

const (
	// EventEntityAdded is emitted when an entity is added
	EventEntityAdded EventType = "entity_added"
	// EventEntityRemoved is emitted when an entity is removed
	EventEntityRemoved EventType = "entity_removed"
	// EventEntityUpdated is emitted when an entity is updated
	EventEntityUpdated EventType = "entity_updated"
	// EventRelationshipAdded is emitted when a relationship is added
	EventRelationshipAdded EventType = "relationship_added"
	// EventRelationshipRemoved is emitted when a relationship is removed
	EventRelationshipRemoved EventType = "relationship_removed"
	// EventWorkspaceCleared is emitted when the workspace is cleared
	EventWorkspaceCleared EventType = "workspace_cleared"
)

// Event represents a workspace event
type Event struct {
	Type         EventType     // The type of event
	Entity       ast.Entity    // The entity involved (for entity events)
	Relationship *Relationship // The relationship involved (for relationship events)
}

// EventHandler is a function that handles workspace events
type EventHandler func(event Event)

// EntityVersion represents a historical version of an entity
type EntityVersion struct {
	Version   int        // Version number (starts at 1)
	Entity    ast.Entity // The entity at this version
	Timestamp int64      // Unix timestamp when this version was created
}

// entityKey creates a unique key for an entity based on type and name
func entityKey(entityType, entityName string) string {
	return entityType + ":" + entityName
}

// Workspace represents a virtual workspace containing entities
type Workspace struct {
	entities          []ast.Entity
	relationships     []Relationship
	hooks             map[HookType][]EntityHook
	eventHandlers     []EventHandler
	entityVersions    map[string][]EntityVersion // Maps entity key to version history
	versioningEnabled bool
	mu                sync.RWMutex
	validator         validator.EntityValidator
}

// New creates a new Workspace instance
func New() *Workspace {
	return &Workspace{
		entities:          make([]ast.Entity, 0),
		relationships:     make([]Relationship, 0),
		hooks:             make(map[HookType][]EntityHook),
		eventHandlers:     make([]EventHandler, 0),
		entityVersions:    make(map[string][]EntityVersion),
		versioningEnabled: false,
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
	ScriptEntities     int
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
		case "script":
			stats.ScriptEntities++
		}
	}

	return stats
}

// WithValidator sets a validator for the workspace
func (w *Workspace) WithValidator(v validator.EntityValidator) *Workspace {
	w.validator = v
	return w
}

// WithVersioning enables entity version tracking.
// When enabled, the workspace maintains a history of all entity versions,
// allowing you to retrieve previous states of entities.
func (w *Workspace) WithVersioning() *Workspace {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.versioningEnabled = true
	return w
}

// recordVersion stores a new version of an entity if versioning is enabled.
// Must be called with lock held.
func (w *Workspace) recordVersion(entity ast.Entity) {
	if !w.versioningEnabled {
		return
	}
	key := entityKey(entity.Type(), entity.Name())
	versions := w.entityVersions[key]
	newVersion := EntityVersion{
		Version:   len(versions) + 1,
		Entity:    entity,
		Timestamp: time.Now().Unix(),
	}
	w.entityVersions[key] = append(versions, newVersion)
}

// GetEntityVersion returns a specific version of an entity.
// Version numbers start at 1. Returns nil if the version doesn't exist.
func (w *Workspace) GetEntityVersion(entityType, entityName string, version int) (ast.Entity, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	key := entityKey(entityType, entityName)
	versions, ok := w.entityVersions[key]
	if !ok || version < 1 || version > len(versions) {
		return nil, false
	}
	return versions[version-1].Entity, true
}

// GetEntityVersionCount returns the number of versions for an entity.
func (w *Workspace) GetEntityVersionCount(entityType, entityName string) int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	key := entityKey(entityType, entityName)
	return len(w.entityVersions[key])
}

// GetEntityHistory returns all versions of an entity in chronological order.
func (w *Workspace) GetEntityHistory(entityType, entityName string) []EntityVersion {
	w.mu.RLock()
	defer w.mu.RUnlock()

	key := entityKey(entityType, entityName)
	versions := w.entityVersions[key]
	if versions == nil {
		return nil
	}
	// Return a copy
	result := make([]EntityVersion, len(versions))
	copy(result, versions)
	return result
}

// OnEntityEvent registers a hook for a specific entity lifecycle event
func (w *Workspace) OnEntityEvent(hookType HookType, hook EntityHook) *Workspace {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.hooks[hookType] = append(w.hooks[hookType], hook)
	return w
}

// OnEvent registers an event handler for workspace events.
// Unlike hooks, event handlers cannot block operations and are called
// after the operation is complete.
func (w *Workspace) OnEvent(handler EventHandler) *Workspace {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.eventHandlers = append(w.eventHandlers, handler)
	return w
}

// emit sends an event to all registered event handlers.
// This is called after operations complete successfully.
func (w *Workspace) emit(event Event) {
	for _, handler := range w.eventHandlers {
		handler(event)
	}
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

	// Record version if versioning is enabled
	w.recordVersion(entity)

	// Run after-add hooks (errors are logged but don't fail the operation)
	_ = w.runHooks(HookAfterAdd, entity)

	// Emit entity added event
	w.emit(Event{Type: EventEntityAdded, Entity: entity})

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

	return slices.Filter(w.entities, func(e ast.Entity) bool {
		return e.Type() == entityType
	})
}

// GetEntityByName returns an entity by type and name
func (w *Workspace) GetEntityByName(entityType, entityName string) (ast.Entity, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return slices.Find(w.entities, func(e ast.Entity) bool {
		return e.Type() == entityType && e.Name() == entityName
	})
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

			// Emit entity removed event
			w.emit(Event{Type: EventEntityRemoved, Entity: entity})

			return nil
		}
	}

	return fmt.Errorf("entity not found: %s %q", entityType, entityName)
}

// UpdateEntity replaces an existing entity with a new version.
// The entity must already exist in the workspace.
// Hooks are called before and after the update.
func (w *Workspace) UpdateEntity(entity ast.Entity) error {
	if entity == nil {
		return fmt.Errorf("cannot update with nil entity")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Find the existing entity
	idx := slices.FindIndex(w.entities, func(e ast.Entity) bool {
		return e.Type() == entity.Type() && e.Name() == entity.Name()
	})
	if idx == -1 {
		return fmt.Errorf("entity not found: %s %q", entity.Type(), entity.Name())
	}

	// Run before-update hooks
	if err := w.runHooks(HookBeforeUpdate, entity); err != nil {
		return err
	}

	// Validate the new entity if validator is set
	if w.validator != nil {
		if err := w.validator.ValidateEntity(entity); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	// Replace the entity
	w.entities[idx] = entity

	// Record version if versioning is enabled
	w.recordVersion(entity)

	// Run after-update hooks
	_ = w.runHooks(HookAfterUpdate, entity)

	// Emit entity updated event
	w.emit(Event{Type: EventEntityUpdated, Entity: entity})

	return nil
}

// UpsertEntity adds an entity if it doesn't exist, or updates it if it does.
// This is a convenience method combining AddEntity and UpdateEntity behavior.
func (w *Workspace) UpsertEntity(entity ast.Entity) error {
	if entity == nil {
		return fmt.Errorf("cannot upsert nil entity")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if entity already exists
	idx := slices.FindIndex(w.entities, func(e ast.Entity) bool {
		return e.Type() == entity.Type() && e.Name() == entity.Name()
	})

	if idx >= 0 {
		// Update existing entity
		if err := w.runHooks(HookBeforeUpdate, entity); err != nil {
			return err
		}

		if w.validator != nil {
			if err := w.validator.ValidateEntity(entity); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
		}

		w.entities[idx] = entity
		w.recordVersion(entity)
		_ = w.runHooks(HookAfterUpdate, entity)
		w.emit(Event{Type: EventEntityUpdated, Entity: entity})
	} else {
		// Add new entity
		if err := w.runHooks(HookBeforeAdd, entity); err != nil {
			return err
		}

		if w.validator != nil {
			if err := w.validator.ValidateEntity(entity); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
		}

		w.entities = append(w.entities, entity)
		w.recordVersion(entity)
		_ = w.runHooks(HookAfterAdd, entity)
		w.emit(Event{Type: EventEntityAdded, Entity: entity})
	}

	return nil
}

// removeRelationshipsForEntity removes all relationships involving the specified entity
// Must be called with lock held
func (w *Workspace) removeRelationshipsForEntity(entityType, entityName string) {
	w.relationships = slices.Filter(w.relationships, func(rel Relationship) bool {
		return !((rel.SourceType == entityType && rel.SourceName == entityName) ||
			(rel.TargetType == entityType && rel.TargetName == entityName))
	})
}

// Clear removes all entities from the workspace
func (w *Workspace) Clear() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.entities = make([]ast.Entity, 0)
	w.relationships = make([]Relationship, 0)

	// Emit workspace cleared event
	w.emit(Event{Type: EventWorkspaceCleared})
}

// AddRelationship creates a relationship between two entities
func (w *Workspace) AddRelationship(sourceType, sourceName, targetType, targetName string, relType RelationType) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Validate that source entity exists
	sourceExists := slices.Any(w.entities, func(e ast.Entity) bool {
		return e.Type() == sourceType && e.Name() == sourceName
	})
	if !sourceExists {
		return fmt.Errorf("source entity not found: %s %q", sourceType, sourceName)
	}

	// Validate that target entity exists
	targetExists := slices.Any(w.entities, func(e ast.Entity) bool {
		return e.Type() == targetType && e.Name() == targetName
	})
	if !targetExists {
		return fmt.Errorf("target entity not found: %s %q", targetType, targetName)
	}

	// Check for duplicate relationship
	isDuplicate := slices.Any(w.relationships, func(rel Relationship) bool {
		return rel.SourceType == sourceType && rel.SourceName == sourceName &&
			rel.TargetType == targetType && rel.TargetName == targetName &&
			rel.Type == relType
	})
	if isDuplicate {
		return fmt.Errorf("relationship already exists")
	}

	rel := Relationship{
		SourceType: sourceType,
		SourceName: sourceName,
		TargetType: targetType,
		TargetName: targetName,
		Type:       relType,
	}
	w.relationships = append(w.relationships, rel)

	// Emit relationship added event
	w.emit(Event{Type: EventRelationshipAdded, Relationship: &rel})

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

	return slices.Filter(w.relationships, func(rel Relationship) bool {
		return (rel.SourceType == entityType && rel.SourceName == entityName) ||
			(rel.TargetType == entityType && rel.TargetName == entityName)
	})
}

// GetRelatedEntities returns entities related to the specified entity
func (w *Workspace) GetRelatedEntities(entityType, entityName string, relType RelationType) []ast.Entity {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Find matching relationships
	matchingRels := slices.Filter(w.relationships, func(rel Relationship) bool {
		if rel.Type != relType {
			return false
		}
		return (rel.SourceType == entityType && rel.SourceName == entityName) ||
			(rel.TargetType == entityType && rel.TargetName == entityName)
	})

	// Collect related entities
	var result []ast.Entity
	for _, rel := range matchingRels {
		var targetType, targetName string
		if rel.SourceType == entityType && rel.SourceName == entityName {
			targetType = rel.TargetType
			targetName = rel.TargetName
		} else {
			targetType = rel.SourceType
			targetName = rel.SourceName
		}

		if entity, found := slices.Find(w.entities, func(e ast.Entity) bool {
			return e.Type() == targetType && e.Name() == targetName
		}); found {
			result = append(result, entity)
		}
	}

	return result
}

// RemoveRelationship removes a specific relationship
func (w *Workspace) RemoveRelationship(sourceType, sourceName, targetType, targetName string, relType RelationType) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	idx := slices.FindIndex(w.relationships, func(rel Relationship) bool {
		return rel.SourceType == sourceType && rel.SourceName == sourceName &&
			rel.TargetType == targetType && rel.TargetName == targetName &&
			rel.Type == relType
	})
	if idx == -1 {
		return fmt.Errorf("relationship not found")
	}

	removedRel := w.relationships[idx]
	w.relationships = append(w.relationships[:idx], w.relationships[idx+1:]...)

	// Emit relationship removed event
	w.emit(Event{Type: EventRelationshipRemoved, Relationship: &removedRel})

	return nil
}

// SerializedProperty represents a property for JSON serialization
type SerializedProperty struct {
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

// SerializedEntity represents an entity for JSON serialization
type SerializedEntity struct {
	Type       string                        `json:"type"`
	Name       string                        `json:"name"`
	Properties map[string]SerializedProperty `json:"properties"`
	Metadata   map[string]string             `json:"metadata,omitempty"`
	Line       int                           `json:"line,omitempty"`
	Column     int                           `json:"column,omitempty"`
}

// SerializedRelationship represents a relationship for JSON serialization
type SerializedRelationship struct {
	SourceType string       `json:"source_type"`
	SourceName string       `json:"source_name"`
	TargetType string       `json:"target_type"`
	TargetName string       `json:"target_name"`
	Type       RelationType `json:"type"`
}

// SerializedWorkspace represents a workspace for JSON serialization
type SerializedWorkspace struct {
	Version       int                      `json:"version"`
	Entities      []SerializedEntity       `json:"entities"`
	Relationships []SerializedRelationship `json:"relationships"`
}

// serializeValue converts an ast.Value to a SerializedProperty
func serializeValue(v ast.Value) (SerializedProperty, error) {
	var prop SerializedProperty
	switch val := v.(type) {
	case ast.StringValue:
		prop.Type = "string"
		data, _ := json.Marshal(val.Value)
		prop.Value = data
	case ast.NumberValue:
		prop.Type = "number"
		data, _ := json.Marshal(val.Value)
		prop.Value = data
	case ast.BoolValue:
		prop.Type = "bool"
		data, _ := json.Marshal(val.Value)
		prop.Value = data
	case ast.ArrayValue:
		prop.Type = "array"
		var elements []SerializedProperty
		for _, elem := range val.Elements {
			se, err := serializeValue(elem)
			if err != nil {
				return prop, err
			}
			elements = append(elements, se)
		}
		data, _ := json.Marshal(elements)
		prop.Value = data
	case ast.ReferenceValue:
		prop.Type = "reference"
		data, _ := json.Marshal(map[string]interface{}{
			"type": val.Type,
			"name": val.Name,
			"path": val.Path,
		})
		prop.Value = data
	case ast.VariableValue:
		prop.Type = "variable"
		data, _ := json.Marshal(val.Name)
		prop.Value = data
	default:
		return prop, fmt.Errorf("unsupported value type: %T", v)
	}
	return prop, nil
}

// deserializeValue converts a SerializedProperty back to an ast.Value
func deserializeValue(prop SerializedProperty) (ast.Value, error) {
	switch prop.Type {
	case "string":
		var s string
		if err := json.Unmarshal(prop.Value, &s); err != nil {
			return nil, err
		}
		return ast.StringValue{Value: s}, nil
	case "number":
		var n float64
		if err := json.Unmarshal(prop.Value, &n); err != nil {
			return nil, err
		}
		return ast.NumberValue{Value: n}, nil
	case "bool":
		var b bool
		if err := json.Unmarshal(prop.Value, &b); err != nil {
			return nil, err
		}
		return ast.BoolValue{Value: b}, nil
	case "array":
		var elements []SerializedProperty
		if err := json.Unmarshal(prop.Value, &elements); err != nil {
			return nil, err
		}
		values := make([]ast.Value, len(elements))
		for i, elem := range elements {
			v, err := deserializeValue(elem)
			if err != nil {
				return nil, err
			}
			values[i] = v
		}
		return ast.ArrayValue{Elements: values}, nil
	case "reference":
		var ref struct {
			Type string   `json:"type"`
			Name string   `json:"name"`
			Path []string `json:"path"`
		}
		if err := json.Unmarshal(prop.Value, &ref); err != nil {
			return nil, err
		}
		return ast.ReferenceValue{Type: ref.Type, Name: ref.Name, Path: ref.Path}, nil
	case "variable":
		var name string
		if err := json.Unmarshal(prop.Value, &name); err != nil {
			return nil, err
		}
		return ast.VariableValue{Name: name}, nil
	default:
		return nil, fmt.Errorf("unsupported value type: %s", prop.Type)
	}
}

// Serialize converts the workspace to a SerializedWorkspace for JSON export.
func (w *Workspace) Serialize() (*SerializedWorkspace, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	sw := &SerializedWorkspace{
		Version:       1,
		Entities:      make([]SerializedEntity, 0, len(w.entities)),
		Relationships: make([]SerializedRelationship, 0, len(w.relationships)),
	}

	for _, entity := range w.entities {
		se := SerializedEntity{
			Type:       entity.Type(),
			Name:       entity.Name(),
			Properties: make(map[string]SerializedProperty),
			Metadata:   entity.AllMetadata(),
			Line:       entity.Line(),
			Column:     entity.Column(),
		}

		for key, val := range entity.Properties() {
			prop, err := serializeValue(val)
			if err != nil {
				return nil, fmt.Errorf("failed to serialize property %s: %w", key, err)
			}
			se.Properties[key] = prop
		}

		sw.Entities = append(sw.Entities, se)
	}

	for _, rel := range w.relationships {
		sw.Relationships = append(sw.Relationships, SerializedRelationship{
			SourceType: rel.SourceType,
			SourceName: rel.SourceName,
			TargetType: rel.TargetType,
			TargetName: rel.TargetName,
			Type:       rel.Type,
		})
	}

	return sw, nil
}

// SaveTo writes the workspace to an io.Writer in JSON format.
func (w *Workspace) SaveTo(writer io.Writer) error {
	sw, err := w.Serialize()
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(sw)
}

// SaveToFile writes the workspace to a file in JSON format.
func (w *Workspace) SaveToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.SaveTo(file)
}

// LoadFrom loads entities and relationships from an io.Reader containing JSON.
// This clears the existing workspace and replaces it with the loaded data.
// Hooks and event handlers are NOT loaded - they must be re-registered.
func (w *Workspace) LoadFrom(reader io.Reader) error {
	var sw SerializedWorkspace
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&sw); err != nil {
		return fmt.Errorf("failed to decode workspace: %w", err)
	}

	return w.loadFromSerialized(&sw)
}

// LoadFromFile loads entities and relationships from a JSON file.
func (w *Workspace) LoadFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return w.LoadFrom(file)
}

// loadFromSerialized populates the workspace from a SerializedWorkspace.
func (w *Workspace) loadFromSerialized(sw *SerializedWorkspace) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Clear existing data (but keep hooks and event handlers)
	w.entities = make([]ast.Entity, 0, len(sw.Entities))
	w.relationships = make([]Relationship, 0, len(sw.Relationships))
	w.entityVersions = make(map[string][]EntityVersion)

	// Load entities
	for _, se := range sw.Entities {
		entity, err := ast.NewEntity(se.Type, se.Name)
		if err != nil {
			return fmt.Errorf("failed to create entity %s/%s: %w", se.Type, se.Name, err)
		}

		// Set properties
		for key, prop := range se.Properties {
			val, err := deserializeValue(prop)
			if err != nil {
				return fmt.Errorf("failed to deserialize property %s: %w", key, err)
			}
			entity.SetProperty(key, val)
		}

		// Set metadata
		for key, value := range se.Metadata {
			entity.SetMetadata(key, value)
		}

		// Set location
		entity.SetLocation(se.Line, se.Column)

		w.entities = append(w.entities, entity)

		// Record version if versioning is enabled
		if w.versioningEnabled {
			key := entityKey(entity.Type(), entity.Name())
			w.entityVersions[key] = []EntityVersion{{
				Version:   1,
				Entity:    entity,
				Timestamp: time.Now().Unix(),
			}}
		}
	}

	// Load relationships
	for _, sr := range sw.Relationships {
		w.relationships = append(w.relationships, Relationship{
			SourceType: sr.SourceType,
			SourceName: sr.SourceName,
			TargetType: sr.TargetType,
			TargetName: sr.TargetName,
			Type:       sr.Type,
		})
	}

	return nil
}

// Snapshot represents a point-in-time snapshot of the workspace state.
type Snapshot struct {
	ID        string               // Unique identifier for the snapshot
	Timestamp int64                // Unix timestamp when snapshot was created
	Data      *SerializedWorkspace // The serialized workspace state
}

// CreateSnapshot creates a point-in-time snapshot of the current workspace state.
// The snapshot can be restored later to return the workspace to this state.
func (w *Workspace) CreateSnapshot(id string) (*Snapshot, error) {
	sw, err := w.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize workspace: %w", err)
	}

	return &Snapshot{
		ID:        id,
		Timestamp: time.Now().Unix(),
		Data:      sw,
	}, nil
}

// RestoreSnapshot restores the workspace to a previous snapshot state.
// This clears the current workspace and replaces it with the snapshot data.
// Hooks and event handlers are preserved.
func (w *Workspace) RestoreSnapshot(snapshot *Snapshot) error {
	if snapshot == nil {
		return fmt.Errorf("cannot restore nil snapshot")
	}
	if snapshot.Data == nil {
		return fmt.Errorf("snapshot contains no data")
	}
	return w.loadFromSerialized(snapshot.Data)
}

// SnapshotStore manages multiple snapshots for a workspace.
type SnapshotStore struct {
	snapshots map[string]*Snapshot
	mu        sync.RWMutex
}

// NewSnapshotStore creates a new snapshot store.
func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{
		snapshots: make(map[string]*Snapshot),
	}
}

// Save stores a snapshot in the store.
func (ss *SnapshotStore) Save(snapshot *Snapshot) error {
	if snapshot == nil {
		return fmt.Errorf("cannot save nil snapshot")
	}
	if snapshot.ID == "" {
		return fmt.Errorf("snapshot must have an ID")
	}

	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.snapshots[snapshot.ID] = snapshot
	return nil
}

// Get retrieves a snapshot by ID.
func (ss *SnapshotStore) Get(id string) (*Snapshot, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	snapshot, ok := ss.snapshots[id]
	return snapshot, ok
}

// Delete removes a snapshot from the store.
func (ss *SnapshotStore) Delete(id string) bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	_, exists := ss.snapshots[id]
	if exists {
		delete(ss.snapshots, id)
	}
	return exists
}

// List returns all snapshot IDs in the store.
func (ss *SnapshotStore) List() []string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	ids := make([]string, 0, len(ss.snapshots))
	for id := range ss.snapshots {
		ids = append(ids, id)
	}
	return ids
}

// Count returns the number of snapshots in the store.
func (ss *SnapshotStore) Count() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return len(ss.snapshots)
}
