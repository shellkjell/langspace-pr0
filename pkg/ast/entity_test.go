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
			name:      "task entity",
			entType:   "task",
			wantType:  "task",
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
			properties: []string{"validator", "instruction"},
			wantError:  false,
		},
		{
			name:       "too many properties",
			properties: []string{"validator", "instruction", "extra"},
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

func TestTaskEntity_AddProperty(t *testing.T) {
	tests := []struct {
		name       string
		properties []string
		wantError  bool
	}{
		{
			name:       "valid properties",
			properties: []string{"build", "instruction"},
			wantError:  false,
		},
		{
			name:       "too many properties",
			properties: []string{"build", "instruction", "extra"},
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, _ := NewEntity("task")
			var err error
			for _, prop := range tt.properties {
				err = entity.AddProperty(prop)
				if err != nil {
					break
				}
			}
			if (err != nil) != tt.wantError {
				t.Errorf("TaskEntity.AddProperty() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestBaseEntity_Properties(t *testing.T) {
	entity := &BaseEntity{
		entityType: "test",
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

func TestEntity_Metadata(t *testing.T) {
	entityTypes := []string{"file", "agent", "task"}

	for _, entType := range entityTypes {
		t.Run(entType+"_metadata", func(t *testing.T) {
			entity, err := NewEntity(entType)
			if err != nil {
				t.Fatalf("NewEntity(%s) error = %v", entType, err)
			}

			// Test GetMetadata on empty entity
			val, ok := entity.GetMetadata("nonexistent")
			if ok {
				t.Error("GetMetadata should return false for nonexistent key")
			}
			if val != "" {
				t.Errorf("GetMetadata should return empty string, got %q", val)
			}

			// Test SetMetadata and GetMetadata
			entity.SetMetadata("author", "test-user")
			entity.SetMetadata("version", "1.0")

			val, ok = entity.GetMetadata("author")
			if !ok {
				t.Error("GetMetadata should return true for existing key")
			}
			if val != "test-user" {
				t.Errorf("GetMetadata(author) = %q, want %q", val, "test-user")
			}

			val, ok = entity.GetMetadata("version")
			if !ok {
				t.Error("GetMetadata should return true for existing key")
			}
			if val != "1.0" {
				t.Errorf("GetMetadata(version) = %q, want %q", val, "1.0")
			}

			// Test AllMetadata
			all := entity.AllMetadata()
			if len(all) != 2 {
				t.Errorf("AllMetadata() returned %d items, want 2", len(all))
			}
			if all["author"] != "test-user" || all["version"] != "1.0" {
				t.Errorf("AllMetadata() = %v, want map with author and version", all)
			}

			// Test that AllMetadata returns a copy (modifying doesn't affect original)
			all["new-key"] = "new-value"
			_, ok = entity.GetMetadata("new-key")
			if ok {
				t.Error("Modifying AllMetadata result should not affect entity")
			}

			// Test overwriting metadata
			entity.SetMetadata("author", "updated-user")
			val, _ = entity.GetMetadata("author")
			if val != "updated-user" {
				t.Errorf("SetMetadata overwrite failed, got %q, want %q", val, "updated-user")
			}
		})
	}
}
