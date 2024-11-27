package ast

import (
	"fmt"
)

// Entity represents a LangSpace entity
type Entity interface {
	Type() string
	Properties() []string
	AddProperty(prop string) error
}

// BaseEntity provides a base implementation of Entity
type BaseEntity struct {
	entityType  string
	properties []string
}

// NewEntity creates a new Entity based on the type
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

// FileEntity represents a file entity in LangSpace
type FileEntity struct {
	Path     string
	Property string
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

// AgentEntity represents an agent entity in LangSpace
type AgentEntity struct {
	Name     string
	Property string
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
