package ast

import (
	"fmt"
)

// Package ast provides the Abstract Syntax Tree (AST) components for LangSpace.
// It defines the core entity types and their behaviors, supporting the language's
// type system and validation rules.

// Entity represents a LangSpace entity, which is the fundamental building block
// of the language. Each entity has a type and a set of properties that define
// its behavior and characteristics.
type Entity interface {
	// Type returns the entity's type identifier (e.g., "file", "agent")
	Type() string

	// Properties returns the entity's current property list
	Properties() []string

	// AddProperty adds a new property to the entity, validating it according
	// to the entity's type-specific rules
	AddProperty(prop string) error

	// GetMetadata returns the value for a metadata key, and whether it exists
	GetMetadata(key string) (string, bool)

	// SetMetadata sets a metadata key-value pair on the entity
	SetMetadata(key, value string)

	// AllMetadata returns a copy of all metadata key-value pairs
	AllMetadata() map[string]string
}

// BaseEntity provides a base implementation of Entity with common functionality
// shared across all entity types. It manages basic property storage and type
// information.
type BaseEntity struct {
	entityType string   // The type identifier for this entity
	properties []string // List of properties associated with this entity
}

// NewEntity creates a new Entity based on the provided type identifier.
// It serves as a factory function for creating type-specific entity instances.
//
// Parameters:
//   - entityType: String identifier for the desired entity type
//
// Returns:
//   - Entity: A new entity instance of the requested type
//   - error: Error if the entity type is unknown
//
// Supported entity types:
//   - "file": Creates a FileEntity for file system operations
//   - "agent": Creates an AgentEntity for automation tasks
//   - "task": Creates a TaskEntity for task management
func NewEntity(entityType string) (Entity, error) {
	switch entityType {
	case "file":
		return &FileEntity{}, nil
	case "agent":
		return &AgentEntity{}, nil
	case "task":
		return &TaskEntity{}, nil
	default:
		return nil, fmt.Errorf("unknown entity type: %s", entityType)
	}
}

// Type returns the entity type
func (e *BaseEntity) Type() string {
	return e.entityType
}

// Properties returns the entity properties
func (e *BaseEntity) Properties() []string {
	return e.properties
}

// FileEntity represents a file entity in LangSpace, used to define and
// manipulate file system resources. It supports two property types:
//   - path: The file system path
//   - contents: The file contents as a string
type FileEntity struct {
	Path     string            // File system path
	Property string            // Property type (either "path" or "contents")
	metadata map[string]string // Key-value metadata storage
}

// Type returns the type of the entity
func (f *FileEntity) Type() string {
	return "file"
}

// Properties returns the properties of the entity
func (f *FileEntity) Properties() []string {
	return []string{f.Path, f.Property}
}

// AddProperty adds a property to the entity
func (f *FileEntity) AddProperty(prop string) error {
	if f.Path == "" {
		f.Path = prop
		return nil
	}
	if f.Property == "" {
		f.Property = prop
		return nil
	}
	return fmt.Errorf("file entity already has all properties set")
}

// GetMetadata returns the value for a metadata key, and whether it exists
func (f *FileEntity) GetMetadata(key string) (string, bool) {
	if f.metadata == nil {
		return "", false
	}
	val, ok := f.metadata[key]
	return val, ok
}

// SetMetadata sets a metadata key-value pair on the entity
func (f *FileEntity) SetMetadata(key, value string) {
	if f.metadata == nil {
		f.metadata = make(map[string]string)
	}
	f.metadata[key] = value
}

// AllMetadata returns a copy of all metadata key-value pairs
func (f *FileEntity) AllMetadata() map[string]string {
	result := make(map[string]string)
	for k, v := range f.metadata {
		result[k] = v
	}
	return result
}

// AgentEntity represents an agent entity in LangSpace, used to define
// automated tasks and validations. Agents can interact with other entities
// and perform operations based on their instructions.
type AgentEntity struct {
	Name     string            // Agent identifier
	Property string            // Property type (e.g., "instruction")
	metadata map[string]string // Key-value metadata storage
}

// Type returns the type of the entity
func (a *AgentEntity) Type() string {
	return "agent"
}

// Properties returns the properties of the entity
func (a *AgentEntity) Properties() []string {
	return []string{a.Name, a.Property}
}

// AddProperty adds a property to the entity
func (a *AgentEntity) AddProperty(prop string) error {
	if a.Name == "" {
		a.Name = prop
		return nil
	}
	if a.Property == "" {
		a.Property = prop
		return nil
	}
	return fmt.Errorf("agent entity already has all properties set")
}

// GetMetadata returns the value for a metadata key, and whether it exists
func (a *AgentEntity) GetMetadata(key string) (string, bool) {
	if a.metadata == nil {
		return "", false
	}
	val, ok := a.metadata[key]
	return val, ok
}

// SetMetadata sets a metadata key-value pair on the entity
func (a *AgentEntity) SetMetadata(key, value string) {
	if a.metadata == nil {
		a.metadata = make(map[string]string)
	}
	a.metadata[key] = value
}

// AllMetadata returns a copy of all metadata key-value pairs
func (a *AgentEntity) AllMetadata() map[string]string {
	result := make(map[string]string)
	for k, v := range a.metadata {
		result[k] = v
	}
	return result
}

// TaskEntity represents a task entity in LangSpace, used to define
// and manage tasks within the virtual workspace. Tasks can have
// instructions and can be automated based on triggers.
type TaskEntity struct {
	Name     string            // Task identifier
	Property string            // Property type (e.g., "instruction", "schedule")
	metadata map[string]string // Key-value metadata storage
}

// Type returns the type of the entity
func (t *TaskEntity) Type() string {
	return "task"
}

// Properties returns the properties of the entity
func (t *TaskEntity) Properties() []string {
	return []string{t.Name, t.Property}
}

// AddProperty adds a property to the entity
func (t *TaskEntity) AddProperty(prop string) error {
	if t.Name == "" {
		t.Name = prop
		return nil
	}
	if t.Property == "" {
		t.Property = prop
		return nil
	}
	return fmt.Errorf("task entity already has all properties set")
}

// GetMetadata returns the value for a metadata key, and whether it exists
func (t *TaskEntity) GetMetadata(key string) (string, bool) {
	if t.metadata == nil {
		return "", false
	}
	val, ok := t.metadata[key]
	return val, ok
}

// SetMetadata sets a metadata key-value pair on the entity
func (t *TaskEntity) SetMetadata(key, value string) {
	if t.metadata == nil {
		t.metadata = make(map[string]string)
	}
	t.metadata[key] = value
}

// AllMetadata returns a copy of all metadata key-value pairs
func (t *TaskEntity) AllMetadata() map[string]string {
	result := make(map[string]string)
	for k, v := range t.metadata {
		result[k] = v
	}
	return result
}
