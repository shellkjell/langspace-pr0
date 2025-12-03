package validator

import (
	"strings"
	"testing"

	"github.com/shellkjell/langspace/pkg/ast"
)

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
			name: "valid file entity with path",
			entity: &testEntity{
				entityType: "file",
				props:      []string{"test.txt", "path"},
			},
			wantError: false,
		},
		{
			name: "valid file entity with contents",
			entity: &testEntity{
				entityType: "file",
				props:      []string{"test.txt", "contents"},
			},
			wantError: false,
		},
		{
			name: "file entity with invalid property",
			entity: &testEntity{
				entityType: "file",
				props:      []string{"test.txt", "invalid"},
			},
			wantError: true,
			errorMsg:  "invalid file property: invalid",
		},
		{
			name: "file entity with empty path",
			entity: &testEntity{
				entityType: "file",
				props:      []string{"", "path"},
			},
			wantError: true,
			errorMsg:  "file path cannot be empty",
		},
		{
			name: "file entity with too many properties",
			entity: &testEntity{
				entityType: "file",
				props:      []string{"test.txt", "path", "extra"},
			},
			wantError: true,
			errorMsg:  "file entity must have exactly 2 properties",
		},
		{
			name: "valid agent entity with instruction",
			entity: &testEntity{
				entityType: "agent",
				props:      []string{"validator", "instruction"},
			},
			wantError: false,
		},
		{
			name: "valid agent entity with model",
			entity: &testEntity{
				entityType: "agent",
				props:      []string{"gpt-4", "model"},
			},
			wantError: false,
		},
		{
			name: "valid agent entity with check",
			entity: &testEntity{
				entityType: "agent",
				props:      []string{"validator", "check(test.txt)"},
			},
			wantError: false,
		},
		{
			name: "agent entity with empty name",
			entity: &testEntity{
				entityType: "agent",
				props:      []string{"", "instruction"},
			},
			wantError: true,
			errorMsg:  "agent name cannot be empty",
		},
		{
			name: "agent entity with invalid property",
			entity: &testEntity{
				entityType: "agent",
				props:      []string{"validator", "invalid"},
			},
			wantError: true,
			errorMsg:  "invalid agent property",
		},
		{
			name: "valid task entity with instruction",
			entity: &testEntity{
				entityType: "task",
				props:      []string{"build", "instruction"},
			},
			wantError: false,
		},
		{
			name: "valid task entity with schedule",
			entity: &testEntity{
				entityType: "task",
				props:      []string{"backup", "schedule"},
			},
			wantError: false,
		},
		{
			name: "valid task entity with priority",
			entity: &testEntity{
				entityType: "task",
				props:      []string{"urgent", "priority"},
			},
			wantError: false,
		},
		{
			name: "task entity with empty name",
			entity: &testEntity{
				entityType: "task",
				props:      []string{"", "instruction"},
			},
			wantError: true,
			errorMsg:  "task name cannot be empty",
		},
		{
			name: "task entity with invalid property",
			entity: &testEntity{
				entityType: "task",
				props:      []string{"build", "invalid"},
			},
			wantError: true,
			errorMsg:  "invalid task property",
		},
		{
			name: "unknown entity type",
			entity: &testEntity{
				entityType: "unknown",
				props:      []string{"test", "prop"},
			},
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

// testEntity is a mock implementation of ast.Entity for testing
type testEntity struct {
	entityType string
	props      []string
	metadata   map[string]string
}

func (e *testEntity) Type() string {
	return e.entityType
}

func (e *testEntity) Properties() []string {
	return e.props
}

func (e *testEntity) AddProperty(prop string) error {
	e.props = append(e.props, prop)
	return nil
}

func (e *testEntity) GetMetadata(key string) (string, bool) {
	if e.metadata == nil {
		return "", false
	}
	val, ok := e.metadata[key]
	return val, ok
}

func (e *testEntity) SetMetadata(key, value string) {
	if e.metadata == nil {
		e.metadata = make(map[string]string)
	}
	e.metadata[key] = value
}

func (e *testEntity) AllMetadata() map[string]string {
	result := make(map[string]string)
	for k, v := range e.metadata {
		result[k] = v
	}
	return result
}
