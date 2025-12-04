package ast

import (
	"fmt"
)

// Package ast provides the Abstract Syntax Tree (AST) components for LangSpace.
// It defines the core entity types and their behaviors, supporting the language's
// type system and validation rules.

// Value represents any value in the AST
type Value interface {
	isValue()
}

// StringValue represents a string literal
type StringValue struct {
	Value string
}

func (s StringValue) isValue() {}

// NumberValue represents a numeric literal
type NumberValue struct {
	Value float64
}

func (n NumberValue) isValue() {}

// BoolValue represents a boolean literal
type BoolValue struct {
	Value bool
}

func (b BoolValue) isValue() {}

// ArrayValue represents an array of values
type ArrayValue struct {
	Elements []Value
}

func (a ArrayValue) isValue() {}

// ObjectValue represents an object/map of key-value pairs
type ObjectValue struct {
	Properties map[string]Value
}

func (o ObjectValue) isValue() {}

// ReferenceValue represents a reference to another entity (e.g., agent("name"))
type ReferenceValue struct {
	Type string // "agent", "file", "tool", "step", etc.
	Name string
	Path []string // For dot access, e.g., step("x").output
}

func (r ReferenceValue) isValue() {}

// VariableValue represents a variable reference (e.g., $input)
type VariableValue struct {
	Name string
}

func (v VariableValue) isValue() {}

// Entity represents a LangSpace entity, which is the fundamental building block
// of the language. Each entity has a type and a set of properties that define
// its behavior and characteristics.
type Entity interface {
	// Type returns the entity's type identifier (e.g., "file", "agent")
	Type() string

	// Name returns the entity's name/identifier
	Name() string

	// Properties returns the entity's property map
	Properties() map[string]Value

	// GetProperty returns a property value by key
	GetProperty(key string) (Value, bool)

	// SetProperty sets a property value
	SetProperty(key string, value Value)

	// GetMetadata returns the value for a metadata key, and whether it exists
	GetMetadata(key string) (string, bool)

	// SetMetadata sets a metadata key-value pair on the entity
	SetMetadata(key, value string)

	// AllMetadata returns a copy of all metadata key-value pairs
	AllMetadata() map[string]string

	// Line returns the source line where this entity was defined
	Line() int

	// Column returns the source column where this entity was defined
	Column() int

	// SetLocation sets the source location
	SetLocation(line, column int)
}

// BaseEntity provides a base implementation of Entity with common functionality
type BaseEntity struct {
	entityType string
	name       string
	properties map[string]Value
	metadata   map[string]string
	line       int
	column     int
}

// NewBaseEntity creates a new BaseEntity
func NewBaseEntity(entityType, name string) *BaseEntity {
	return &BaseEntity{
		entityType: entityType,
		name:       name,
		properties: make(map[string]Value),
		metadata:   make(map[string]string),
	}
}

func (e *BaseEntity) Type() string                      { return e.entityType }
func (e *BaseEntity) Name() string                      { return e.name }
func (e *BaseEntity) Properties() map[string]Value      { return e.properties }
func (e *BaseEntity) Line() int                         { return e.line }
func (e *BaseEntity) Column() int                       { return e.column }
func (e *BaseEntity) SetLocation(line, column int)      { e.line = line; e.column = column }
func (e *BaseEntity) SetProperty(key string, val Value) { e.properties[key] = val }

func (e *BaseEntity) GetProperty(key string) (Value, bool) {
	v, ok := e.properties[key]
	return v, ok
}

func (e *BaseEntity) GetMetadata(key string) (string, bool) {
	if e.metadata == nil {
		return "", false
	}
	val, ok := e.metadata[key]
	return val, ok
}

func (e *BaseEntity) SetMetadata(key, value string) {
	if e.metadata == nil {
		e.metadata = make(map[string]string)
	}
	e.metadata[key] = value
}

func (e *BaseEntity) AllMetadata() map[string]string {
	result := make(map[string]string)
	for k, v := range e.metadata {
		result[k] = v
	}
	return result
}

// FileEntity represents a file entity in LangSpace
type FileEntity struct {
	*BaseEntity
}

// NewFileEntity creates a new file entity
func NewFileEntity(name string) *FileEntity {
	return &FileEntity{BaseEntity: NewBaseEntity("file", name)}
}

// AgentEntity represents an agent entity in LangSpace
type AgentEntity struct {
	*BaseEntity
}

// NewAgentEntity creates a new agent entity
func NewAgentEntity(name string) *AgentEntity {
	return &AgentEntity{BaseEntity: NewBaseEntity("agent", name)}
}

// ToolEntity represents a tool entity in LangSpace
type ToolEntity struct {
	*BaseEntity
}

// NewToolEntity creates a new tool entity
func NewToolEntity(name string) *ToolEntity {
	return &ToolEntity{BaseEntity: NewBaseEntity("tool", name)}
}

// IntentEntity represents an intent entity in LangSpace
type IntentEntity struct {
	*BaseEntity
}

// NewIntentEntity creates a new intent entity
func NewIntentEntity(name string) *IntentEntity {
	return &IntentEntity{BaseEntity: NewBaseEntity("intent", name)}
}

// PipelineEntity represents a pipeline entity in LangSpace
type PipelineEntity struct {
	*BaseEntity
	Steps []*StepEntity
}

// NewPipelineEntity creates a new pipeline entity
func NewPipelineEntity(name string) *PipelineEntity {
	return &PipelineEntity{
		BaseEntity: NewBaseEntity("pipeline", name),
		Steps:      make([]*StepEntity, 0),
	}
}

// AddStep adds a step to the pipeline
func (p *PipelineEntity) AddStep(step *StepEntity) {
	p.Steps = append(p.Steps, step)
}

// StepEntity represents a step within a pipeline
type StepEntity struct {
	*BaseEntity
}

// NewStepEntity creates a new step entity
func NewStepEntity(name string) *StepEntity {
	return &StepEntity{BaseEntity: NewBaseEntity("step", name)}
}

// TriggerEntity represents a trigger entity in LangSpace
type TriggerEntity struct {
	*BaseEntity
}

// NewTriggerEntity creates a new trigger entity
func NewTriggerEntity(name string) *TriggerEntity {
	return &TriggerEntity{BaseEntity: NewBaseEntity("trigger", name)}
}

// ConfigEntity represents a config block in LangSpace
type ConfigEntity struct {
	*BaseEntity
}

// NewConfigEntity creates a new config entity
func NewConfigEntity() *ConfigEntity {
	return &ConfigEntity{BaseEntity: NewBaseEntity("config", "")}
}

// MCPEntity represents an MCP server connection
type MCPEntity struct {
	*BaseEntity
}

// NewMCPEntity creates a new MCP entity
func NewMCPEntity(name string) *MCPEntity {
	return &MCPEntity{BaseEntity: NewBaseEntity("mcp", name)}
}

// ScriptEntity represents a script entity in LangSpace.
// Scripts enable code-first agent actions — a more efficient alternative to
// multiple tool calls. Instead of loading full data into the context window
// through repeated tool invocations, agents write executable code that performs
// complex operations in a single execution, returning only the results.
//
// Key properties:
//   - language: The programming language (python, javascript, bash, sql)
//   - runtime: The runtime/interpreter to use (python3, node, bash, postgresql)
//   - code: The script source code (inline or file reference)
//   - parameters: Input parameters passed to the script
//   - capabilities: What the script can access (database, filesystem, network)
//   - timeout: Maximum execution time
//   - limits: Resource constraints (memory, cpu)
//   - sandbox: Security restrictions (allowed_modules, network access)
type ScriptEntity struct {
	*BaseEntity
}

// NewScriptEntity creates a new script entity
func NewScriptEntity(name string) *ScriptEntity {
	return &ScriptEntity{BaseEntity: NewBaseEntity("script", name)}
}

// NewEntity creates a new Entity based on the provided type identifier.
func NewEntity(entityType string, name string) (Entity, error) {
	switch entityType {
	case "file":
		return NewFileEntity(name), nil
	case "agent":
		return NewAgentEntity(name), nil
	case "tool":
		return NewToolEntity(name), nil
	case "intent":
		return NewIntentEntity(name), nil
	case "pipeline":
		return NewPipelineEntity(name), nil
	case "step":
		return NewStepEntity(name), nil
	case "trigger":
		return NewTriggerEntity(name), nil
	case "config":
		return NewConfigEntity(), nil
	case "mcp":
		return NewMCPEntity(name), nil
	case "script":
		return NewScriptEntity(name), nil
	default:
		return nil, fmt.Errorf("unknown entity type: %s", entityType)
	}
}

// Legacy support - keeping old function signature for compatibility
// Deprecated: Use NewEntity(entityType, name) instead
func NewEntityLegacy(entityType string) (Entity, error) {
	return NewEntity(entityType, "")
}
