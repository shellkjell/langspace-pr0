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
				props:     []string{"test.txt", "path"},
			},
			wantError: false,
		},
		{
			name: "valid file entity with contents",
			entity: &testEntity{
				entityType: "file",
				props:     []string{"test.txt", "contents"},
			},
			wantError: false,
		},
		{
			name: "file entity with invalid property",
			entity: &testEntity{
				entityType: "file",
				props:     []string{"test.txt", "invalid"},
			},
			wantError: true,
			errorMsg:  "invalid file property: invalid",
		},
		{
			name: "file entity with empty path",
			entity: &testEntity{
				entityType: "file",
				props:     []string{"", "path"},
			},
			wantError: true,
			errorMsg:  "file path cannot be empty",
		},
		{
			name: "file entity with too many properties",
			entity: &testEntity{
				entityType: "file",
				props:     []string{"test.txt", "path", "extra"},
			},
			wantError: true,
			errorMsg:  "file entity must have exactly 2 properties",
		},
		{
			name: "valid agent entity",
			entity: &testEntity{
				entityType: "agent",
				props:     []string{"validator", "check(test.txt)"},
			},
			wantError: false,
		},
		{
			name: "agent entity with empty name",
			entity: &testEntity{
				entityType: "agent",
				props:     []string{"", "check(test.txt)"},
			},
			wantError: true,
			errorMsg:  "agent name cannot be empty",
		},
		{
			name: "agent entity with invalid property format",
			entity: &testEntity{
				entityType: "agent",
				props:     []string{"validator", "invalid"},
			},
			wantError: true,
			errorMsg:  "invalid agent property format",
		},
		{
			name: "unknown entity type",
			entity: &testEntity{
				entityType: "unknown",
				props:     []string{"test", "prop"},
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
	props     []string
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
