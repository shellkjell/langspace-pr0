package ast

import (
	"testing"
)

func TestNewEntity(t *testing.T) {
	tests := []struct {
		name      string
		entType   string
		wantType  string
		wantError bool
	}{
		{
			name:      "file entity",
			entType:   "file",
			wantType:  "file",
			wantError: false,
		},
		{
			name:      "agent entity",
			entType:   "agent",
			wantType:  "agent",
			wantError: false,
		},
		{
			name:      "unknown entity",
			entType:   "unknown",
			wantType:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEntity(tt.entType)
			if (err != nil) != tt.wantError {
				t.Errorf("NewEntity() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && got.Type() != tt.wantType {
				t.Errorf("NewEntity().Type() = %v, want %v", got.Type(), tt.wantType)
			}
		})
	}
}

func TestFileEntity_AddProperty(t *testing.T) {
	tests := []struct {
		name       string
		properties []string
		wantError  bool
	}{
		{
			name:       "valid properties",
			properties: []string{"test.txt", "content"},
			wantError:  false,
		},
		{
			name:       "too many properties",
			properties: []string{"test.txt", "content", "extra"},
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, _ := NewEntity("file")
			var err error
			for _, prop := range tt.properties {
				err = entity.AddProperty(prop)
				if err != nil {
					break
				}
			}
			if (err != nil) != tt.wantError {
				t.Errorf("FileEntity.AddProperty() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAgentEntity_AddProperty(t *testing.T) {
	tests := []struct {
		name       string
		properties []string
		wantError  bool
	}{
		{
			name:       "valid properties",
			properties: []string{"validator", "check(file.txt)"},
			wantError:  false,
		},
		{
			name:       "too many properties",
			properties: []string{"validator", "check(file.txt)", "extra"},
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, _ := NewEntity("agent")
			var err error
			for _, prop := range tt.properties {
				err = entity.AddProperty(prop)
				if err != nil {
					break
				}
			}
			if (err != nil) != tt.wantError {
				t.Errorf("AgentEntity.AddProperty() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestBaseEntity_Properties(t *testing.T) {
	entity := &BaseEntity{
		entityType:  "test",
		properties: []string{"prop1", "prop2"},
	}

	props := entity.Properties()
	if len(props) != 2 {
		t.Errorf("BaseEntity.Properties() returned %d properties, want 2", len(props))
	}
	if props[0] != "prop1" || props[1] != "prop2" {
		t.Errorf("BaseEntity.Properties() = %v, want [prop1 prop2]", props)
	}
}

func BenchmarkNewEntity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewEntity("file")
	}
}

func BenchmarkAddProperty(b *testing.B) {
	entity, _ := NewEntity("file")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entity.AddProperty("test.txt")
	}
}
