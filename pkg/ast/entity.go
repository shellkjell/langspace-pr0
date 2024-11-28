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
}

// BaseEntity provides a base implementation of Entity with common functionality
// shared across all entity types. It manages basic property storage and type
// information.
type BaseEntity struct {
	entityType  string    // The type identifier for this entity
	properties []string   // List of properties associated with this entity
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
func NewEntity(entityType string) (Entity, error) {
	switch entityType {
	case "file":
		return &FileEntity{}, nil
	case "agent":
		return &AgentEntity{}, nil
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
	Path     string // File system path
	Property string // Property type (either "path" or "contents")
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

// AgentEntity represents an agent entity in LangSpace, used to define
// automated tasks and validations. Agents can interact with other entities
// and perform operations based on their instructions.
type AgentEntity struct {
	Name     string // Agent identifier
	Property string // Property type (e.g., "instruction")
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
