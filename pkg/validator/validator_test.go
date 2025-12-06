package validator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/shellkjell/langspace/pkg/ast"
)

// Helper functions to create test entities - use typed constructors for simplicity
func createFileEntity(name string) ast.Entity {
	entity := ast.NewFileEntity(name)
	// File entity needs path or contents
	entity.SetProperty("path", ast.StringValue{Value: "/tmp/test"})
	return entity
}

func createAgentEntity(name string) ast.Entity {
	entity := ast.NewAgentEntity(name)
	entity.SetProperty("model", ast.StringValue{Value: "gpt-4"})
	entity.SetProperty("instruction", ast.StringValue{Value: "You are an assistant"})
	return entity
}

func createToolEntity(name string) ast.Entity {
	entity := ast.NewToolEntity(name)
	entity.SetProperty("command", ast.StringValue{Value: "echo hello"})
	return entity
}

func createIntentEntity(name string) ast.Entity {
	entity := ast.NewIntentEntity(name)
	entity.SetProperty("use", ast.ReferenceValue{Type: "agent", Name: "test_agent"})
	return entity
}

func createPipelineEntity(name string) ast.Entity {
	return ast.NewPipelineEntity(name)
}

func createTriggerEntity(name string) ast.Entity {
	entity := ast.NewTriggerEntity(name)
	entity.SetProperty("event", ast.StringValue{Value: "on_start"})
	return entity
}

func createConfigEntity() ast.Entity {
	entity := ast.NewConfigEntity()
	entity.SetProperty("default_model", ast.StringValue{Value: "gpt-4"})
	return entity
}

func createMCPEntity(name string) ast.Entity {
	entity := ast.NewMCPEntity(name)
	entity.SetProperty("command", ast.StringValue{Value: "node server.js"})
	return entity
}

func createStepEntity(name string) ast.Entity {
	entity := ast.NewStepEntity(name)
	entity.SetProperty("use", ast.ReferenceValue{Type: "agent", Name: "test"})
	return entity
}

func createScriptEntity(name string) ast.Entity {
	entity := ast.NewScriptEntity(name)
	entity.SetProperty("language", ast.StringValue{Value: "python"})
	entity.SetProperty("code", ast.StringValue{Value: "print('hello')"})
	return entity
}

func TestValidator_ValidateEntity(t *testing.T) {
	tests := []struct {
		name      string
		entity    ast.Entity
		wantError bool
		errorMsg  string
	}{
		{
			name:      "nil entity",
			entity:    nil,
			wantError: true,
			errorMsg:  "entity cannot be nil",
		},
		{
			name:      "valid file entity",
			entity:    createFileEntity("test.txt"),
			wantError: false,
		},
		{
			name:      "file entity with empty name",
			entity:    createFileEntity(""),
			wantError: true,
			errorMsg:  "file entity must have a name",
		},
		{
			name: "file entity without path or contents",
			entity: func() ast.Entity {
				return ast.NewFileEntity("test.txt")
			}(),
			wantError: true,
			errorMsg:  "file entity must have either 'path' or 'contents' property",
		},
		{
			name:      "valid agent entity",
			entity:    createAgentEntity("assistant"),
			wantError: false,
		},
		{
			name:      "agent entity with empty name",
			entity:    createAgentEntity(""),
			wantError: true,
			errorMsg:  "agent entity must have a name",
		},
		{
			name: "agent entity without model",
			entity: func() ast.Entity {
				e := ast.NewAgentEntity("test")
				e.SetProperty("instruction", ast.StringValue{Value: "test"})
				return e
			}(),
			wantError: true,
			errorMsg:  "agent entity must have 'model' property",
		},
		{
			name:      "valid tool entity",
			entity:    createToolEntity("calculator"),
			wantError: false,
		},
		{
			name:      "tool entity with empty name",
			entity:    createToolEntity(""),
			wantError: true,
			errorMsg:  "tool entity must have a name",
		},
		{
			name: "tool entity without command or function",
			entity: func() ast.Entity {
				return ast.NewToolEntity("test")
			}(),
			wantError: true,
			errorMsg:  "tool entity must have either 'command' or 'function' property",
		},
		{
			name:      "valid intent entity",
			entity:    createIntentEntity("analyze"),
			wantError: false,
		},
		{
			name:      "intent entity with empty name",
			entity:    createIntentEntity(""),
			wantError: true,
			errorMsg:  "intent entity must have a name",
		},
		{
			name: "intent entity without use",
			entity: func() ast.Entity {
				return ast.NewIntentEntity("test")
			}(),
			wantError: true,
			errorMsg:  "intent entity must have 'use' property",
		},
		{
			name:      "valid pipeline entity",
			entity:    createPipelineEntity("build"),
			wantError: false,
		},
		{
			name:      "pipeline entity with empty name",
			entity:    createPipelineEntity(""),
			wantError: true,
			errorMsg:  "pipeline entity must have a name",
		},
		{
			name:      "valid trigger entity",
			entity:    createTriggerEntity("startup"),
			wantError: false,
		},
		{
			name:      "trigger entity with empty name",
			entity:    createTriggerEntity(""),
			wantError: true,
			errorMsg:  "trigger entity must have a name",
		},
		{
			name: "trigger entity without event or schedule",
			entity: func() ast.Entity {
				return ast.NewTriggerEntity("test")
			}(),
			wantError: true,
			errorMsg:  "trigger entity must have 'event' or 'schedule' property",
		},
		{
			name:      "valid config entity",
			entity:    createConfigEntity(),
			wantError: false,
		},
		{
			name: "config entity without properties",
			entity: func() ast.Entity {
				return ast.NewConfigEntity()
			}(),
			wantError: true,
			errorMsg:  "config entity must have at least one property",
		},
		{
			name:      "valid mcp entity",
			entity:    createMCPEntity("server"),
			wantError: false,
		},
		{
			name:      "mcp entity with empty name",
			entity:    createMCPEntity(""),
			wantError: true,
			errorMsg:  "mcp entity must have a name",
		},
		{
			name: "mcp entity without command",
			entity: func() ast.Entity {
				return ast.NewMCPEntity("test")
			}(),
			wantError: true,
			errorMsg:  "mcp entity must have 'command' property",
		},
		{
			name:      "valid step entity",
			entity:    createStepEntity("process"),
			wantError: false,
		},
		{
			name: "step entity with empty name",
			entity: func() ast.Entity {
				return ast.NewStepEntity("")
			}(),
			wantError: true,
			errorMsg:  "step entity must have a name",
		},
		{
			name: "step entity without use",
			entity: func() ast.Entity {
				return ast.NewStepEntity("test")
			}(),
			wantError: true,
			errorMsg:  "step entity must have 'use' property",
		},
		{
			name:      "valid script entity",
			entity:    createScriptEntity("update-db"),
			wantError: false,
		},
		{
			name: "script entity with empty name",
			entity: func() ast.Entity {
				e := ast.NewScriptEntity("")
				e.SetProperty("language", ast.StringValue{Value: "python"})
				e.SetProperty("code", ast.StringValue{Value: "print('hi')"})
				return e
			}(),
			wantError: true,
			errorMsg:  "script entity must have a name",
		},
		{
			name: "script entity without language",
			entity: func() ast.Entity {
				e := ast.NewScriptEntity("test")
				e.SetProperty("code", ast.StringValue{Value: "print('hi')"})
				return e
			}(),
			wantError: true,
			errorMsg:  "script entity must have 'language' property",
		},
		{
			name: "script entity without code or path",
			entity: func() ast.Entity {
				e := ast.NewScriptEntity("test")
				e.SetProperty("language", ast.StringValue{Value: "python"})
				return e
			}(),
			wantError: true,
			errorMsg:  "script entity must have 'code' or 'path' property",
		},
		{
			name: "script entity with path instead of code",
			entity: func() ast.Entity {
				e := ast.NewScriptEntity("test")
				e.SetProperty("language", ast.StringValue{Value: "python"})
				e.SetProperty("path", ast.StringValue{Value: "/scripts/test.py"})
				return e
			}(),
			wantError: false,
		},
		{
			name: "unknown entity type",
			entity: func() ast.Entity {
				return &unknownEntity{name: "test"}
			}(),
			wantError: true,
			errorMsg:  "unknown entity type: unknown",
		},
	}

	v := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateEntity(tt.entity)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateEntity() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError && err != nil && tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("ValidateEntity() error = %v, want error containing %v", err, tt.errorMsg)
			}
		})
	}
}

// unknownEntity is a mock implementation for testing unknown entity types
type unknownEntity struct {
	name       string
	properties map[string]ast.Value
	metadata   map[string]string
	line       int
	column     int
}

func (e *unknownEntity) Type() string {
	return "unknown"
}

func (e *unknownEntity) Name() string {
	return e.name
}

func (e *unknownEntity) Properties() map[string]ast.Value {
	if e.properties == nil {
		return make(map[string]ast.Value)
	}
	return e.properties
}

func (e *unknownEntity) SetProperty(key string, val ast.Value) {
	if e.properties == nil {
		e.properties = make(map[string]ast.Value)
	}
	e.properties[key] = val
}

func (e *unknownEntity) GetProperty(key string) (ast.Value, bool) {
	if e.properties == nil {
		return nil, false
	}
	val, ok := e.properties[key]
	return val, ok
}

func (e *unknownEntity) GetMetadata(key string) (string, bool) {
	if e.metadata == nil {
		return "", false
	}
	val, ok := e.metadata[key]
	return val, ok
}

func (e *unknownEntity) SetMetadata(key, value string) {
	if e.metadata == nil {
		e.metadata = make(map[string]string)
	}
	e.metadata[key] = value
}

func (e *unknownEntity) AllMetadata() map[string]string {
	result := make(map[string]string)
	for k, v := range e.metadata {
		result[k] = v
	}
	return result
}

func (e *unknownEntity) Line() int {
	return e.line
}

func (e *unknownEntity) Column() int {
	return e.column
}

func (e *unknownEntity) SetLocation(line, column int) {
	e.line = line
	e.column = column
}

func TestValidator_RegisterValidator(t *testing.T) {
	v := New()

	// Register a custom validator for the unknown entity type
	customValidatorCalled := false
	v.RegisterValidator("unknown", func(entity ast.Entity) error {
		customValidatorCalled = true
		if entity.Name() == "" {
			return fmt.Errorf("custom: entity must have a name")
		}
		return nil
	})

	// Test that custom validator is called
	entity := &unknownEntity{name: "test"}
	err := v.ValidateEntity(entity)
	if err != nil {
		t.Errorf("ValidateEntity() unexpected error = %v", err)
	}
	if !customValidatorCalled {
		t.Error("custom validator was not called")
	}

	// Test custom validator error
	entity2 := &unknownEntity{name: ""}
	err = v.ValidateEntity(entity2)
	if err == nil {
		t.Error("expected error from custom validator")
	}
	if !strings.Contains(err.Error(), "custom: entity must have a name") {
		t.Errorf("expected custom error message, got: %v", err)
	}
}

func TestValidator_CustomValidatorOverridesBuiltin(t *testing.T) {
	v := New()

	// Override the built-in file validator
	v.RegisterValidator("file", func(entity ast.Entity) error {
		// Custom validation: require a "custom_prop" property
		if _, ok := entity.GetProperty("custom_prop"); !ok {
			return fmt.Errorf("file entity must have 'custom_prop' property")
		}
		return nil
	})

	// File with path but no custom_prop should now fail
	entity := ast.NewFileEntity("test.txt")
	entity.SetProperty("path", ast.StringValue{Value: "/tmp/test"})

	err := v.ValidateEntity(entity)
	if err == nil {
		t.Error("expected error from custom validator")
	}

	// File with custom_prop should pass
	entity.SetProperty("custom_prop", ast.StringValue{Value: "value"})
	err = v.ValidateEntity(entity)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
