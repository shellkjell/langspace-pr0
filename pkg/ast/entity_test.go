package ast

import (
	"testing"
)

// Value type tests

func TestStringValue(t *testing.T) {
	v := StringValue{Value: "hello"}
	// StringValue implements Value interface
	var _ Value = v
}

func TestNumberValue(t *testing.T) {
	v := NumberValue{Value: 3.14}
	// NumberValue implements Value interface
	var _ Value = v
}

func TestBoolValue(t *testing.T) {
	v1 := BoolValue{Value: true}
	v2 := BoolValue{Value: false}
	// BoolValue implements Value interface
	var _ Value = v1
	var _ Value = v2
}

func TestArrayValue(t *testing.T) {
	v := ArrayValue{Elements: []Value{
		StringValue{Value: "a"},
		StringValue{Value: "b"},
		NumberValue{Value: 3},
	}}
	// ArrayValue implements Value interface
	var _ Value = v

	if len(v.Elements) != 3 {
		t.Errorf("ArrayValue.Elements has %d items, want 3", len(v.Elements))
	}
}

func TestObjectValue(t *testing.T) {
	v := ObjectValue{Properties: map[string]Value{
		"name": StringValue{Value: "test"},
		"age":  NumberValue{Value: 25},
	}}
	// ObjectValue implements Value interface
	var _ Value = v

	if len(v.Properties) != 2 {
		t.Errorf("ObjectValue.Properties has %d items, want 2", len(v.Properties))
	}
}

func TestReferenceValue(t *testing.T) {
	v := ReferenceValue{
		Type: "agent",
		Name: "reviewer",
		Path: []string{"output"},
	}
	// ReferenceValue implements Value interface
	var _ Value = v

	if v.Type != "agent" {
		t.Errorf("ReferenceValue.Type = %q, want %q", v.Type, "agent")
	}
	if v.Name != "reviewer" {
		t.Errorf("ReferenceValue.Name = %q, want %q", v.Name, "reviewer")
	}
	if len(v.Path) != 1 || v.Path[0] != "output" {
		t.Errorf("ReferenceValue.Path = %v, want [output]", v.Path)
	}
}

func TestVariableValue(t *testing.T) {
	v := VariableValue{Name: "input"}
	// VariableValue implements Value interface
	var _ Value = v

	if v.Name != "input" {
		t.Errorf("VariableValue.Name = %q, want %q", v.Name, "input")
	}
}

// Entity type tests

func TestFileEntity(t *testing.T) {
	e := NewFileEntity("config.json")
	e.SetProperty("path", StringValue{Value: "/etc/config.json"})
	e.SetProperty("contents", StringValue{Value: "{}"})

	if e.Type() != "file" {
		t.Errorf("FileEntity.Type() = %q, want %q", e.Type(), "file")
	}
	if e.Name() != "config.json" {
		t.Errorf("FileEntity.Name() = %q, want %q", e.Name(), "config.json")
	}

	props := e.Properties()
	if len(props) != 2 {
		t.Errorf("FileEntity.Properties() has %d items, want 2", len(props))
	}

	path, ok := e.GetProperty("path")
	if !ok {
		t.Error("FileEntity.GetProperty(path) should return true")
	}
	if sv, ok := path.(StringValue); !ok || sv.Value != "/etc/config.json" {
		t.Errorf("FileEntity.GetProperty(path) = %v, want /etc/config.json", path)
	}
}

func TestAgentEntity(t *testing.T) {
	e := NewAgentEntity("reviewer")
	e.SetProperty("model", StringValue{Value: "gpt-4o"})
	e.SetProperty("temperature", NumberValue{Value: 0.7})
	e.SetProperty("tools", ArrayValue{Elements: []Value{
		StringValue{Value: "read_file"},
		StringValue{Value: "write_file"},
	}})

	if e.Type() != "agent" {
		t.Errorf("AgentEntity.Type() = %q, want %q", e.Type(), "agent")
	}
	if e.Name() != "reviewer" {
		t.Errorf("AgentEntity.Name() = %q, want %q", e.Name(), "reviewer")
	}

	model, ok := e.GetProperty("model")
	if !ok {
		t.Error("AgentEntity.GetProperty(model) should return true")
	}
	if sv, ok := model.(StringValue); !ok || sv.Value != "gpt-4o" {
		t.Errorf("AgentEntity.GetProperty(model) = %v, want gpt-4o", model)
	}

	temp, ok := e.GetProperty("temperature")
	if !ok {
		t.Error("AgentEntity.GetProperty(temperature) should return true")
	}
	if nv, ok := temp.(NumberValue); !ok || nv.Value != 0.7 {
		t.Errorf("AgentEntity.GetProperty(temperature) = %v, want 0.7", temp)
	}
}

func TestToolEntity(t *testing.T) {
	e := NewToolEntity("linter")
	e.SetProperty("command", StringValue{Value: "eslint"})

	if e.Type() != "tool" {
		t.Errorf("ToolEntity.Type() = %q, want %q", e.Type(), "tool")
	}
	if e.Name() != "linter" {
		t.Errorf("ToolEntity.Name() = %q, want %q", e.Name(), "linter")
	}
}

func TestIntentEntity(t *testing.T) {
	e := NewIntentEntity("review-code")
	e.SetProperty("use", ReferenceValue{Type: "agent", Name: "reviewer"})
	e.SetProperty("input", VariableValue{Name: "code"})

	if e.Type() != "intent" {
		t.Errorf("IntentEntity.Type() = %q, want %q", e.Type(), "intent")
	}
	if e.Name() != "review-code" {
		t.Errorf("IntentEntity.Name() = %q, want %q", e.Name(), "review-code")
	}
}

func TestPipelineEntity(t *testing.T) {
	e := NewPipelineEntity("code-review-pipeline")
	e.SetProperty("input", VariableValue{Name: "code"})

	if e.Type() != "pipeline" {
		t.Errorf("PipelineEntity.Type() = %q, want %q", e.Type(), "pipeline")
	}
	if e.Name() != "code-review-pipeline" {
		t.Errorf("PipelineEntity.Name() = %q, want %q", e.Name(), "code-review-pipeline")
	}
}

func TestPipelineEntity_AddStep(t *testing.T) {
	p := NewPipelineEntity("test-pipeline")
	step1 := NewStepEntity("step1")
	step2 := NewStepEntity("step2")

	p.AddStep(step1)
	p.AddStep(step2)

	if len(p.Steps) != 2 {
		t.Errorf("PipelineEntity.Steps has %d items, want 2", len(p.Steps))
	}
	if p.Steps[0].Name() != "step1" {
		t.Errorf("PipelineEntity.Steps[0].Name() = %q, want step1", p.Steps[0].Name())
	}
	if p.Steps[1].Name() != "step2" {
		t.Errorf("PipelineEntity.Steps[1].Name() = %q, want step2", p.Steps[1].Name())
	}
}

func TestStepEntity(t *testing.T) {
	e := NewStepEntity("analyze")
	e.SetProperty("use", ReferenceValue{Type: "agent", Name: "analyzer"})

	if e.Type() != "step" {
		t.Errorf("StepEntity.Type() = %q, want %q", e.Type(), "step")
	}
	if e.Name() != "analyze" {
		t.Errorf("StepEntity.Name() = %q, want %q", e.Name(), "analyze")
	}
}

func TestTriggerEntity(t *testing.T) {
	e := NewTriggerEntity("on-push")
	e.SetProperty("event", StringValue{Value: "push"})

	if e.Type() != "trigger" {
		t.Errorf("TriggerEntity.Type() = %q, want %q", e.Type(), "trigger")
	}
	if e.Name() != "on-push" {
		t.Errorf("TriggerEntity.Name() = %q, want %q", e.Name(), "on-push")
	}
}

func TestConfigEntity(t *testing.T) {
	e := NewConfigEntity()
	e.SetProperty("api_key", VariableValue{Name: "OPENAI_API_KEY"})

	if e.Type() != "config" {
		t.Errorf("ConfigEntity.Type() = %q, want %q", e.Type(), "config")
	}
}

func TestMCPEntity(t *testing.T) {
	e := NewMCPEntity("filesystem")
	e.SetProperty("command", StringValue{Value: "npx"})
	e.SetProperty("args", ArrayValue{Elements: []Value{
		StringValue{Value: "@modelcontextprotocol/server-filesystem"},
		StringValue{Value: "/tmp"},
	}})

	if e.Type() != "mcp" {
		t.Errorf("MCPEntity.Type() = %q, want %q", e.Type(), "mcp")
	}
	if e.Name() != "filesystem" {
		t.Errorf("MCPEntity.Name() = %q, want %q", e.Name(), "filesystem")
	}
}

func TestNewEntity(t *testing.T) {
	tests := []struct {
		name      string
		entType   string
		entName   string
		wantType  string
		wantError bool
	}{
		{
			name:      "file entity",
			entType:   "file",
			entName:   "test.txt",
			wantType:  "file",
			wantError: false,
		},
		{
			name:      "agent entity",
			entType:   "agent",
			entName:   "reviewer",
			wantType:  "agent",
			wantError: false,
		},
		{
			name:      "tool entity",
			entType:   "tool",
			entName:   "linter",
			wantType:  "tool",
			wantError: false,
		},
		{
			name:      "intent entity",
			entType:   "intent",
			entName:   "review",
			wantType:  "intent",
			wantError: false,
		},
		{
			name:      "pipeline entity",
			entType:   "pipeline",
			entName:   "build",
			wantType:  "pipeline",
			wantError: false,
		},
		{
			name:      "step entity",
			entType:   "step",
			entName:   "analyze",
			wantType:  "step",
			wantError: false,
		},
		{
			name:      "trigger entity",
			entType:   "trigger",
			entName:   "on-push",
			wantType:  "trigger",
			wantError: false,
		},
		{
			name:      "config entity",
			entType:   "config",
			entName:   "",
			wantType:  "config",
			wantError: false,
		},
		{
			name:      "mcp entity",
			entType:   "mcp",
			entName:   "fs",
			wantType:  "mcp",
			wantError: false,
		},
		{
			name:      "unknown entity",
			entType:   "unknown",
			entName:   "test",
			wantType:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEntity(tt.entType, tt.entName)
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

func TestEntityPropertyOperations(t *testing.T) {
	entities := []struct {
		name       string
		entityType string
	}{
		{"file", "file"},
		{"agent", "agent"},
		{"tool", "tool"},
		{"intent", "intent"},
		{"pipeline", "pipeline"},
		{"step", "step"},
		{"trigger", "trigger"},
		{"config", "config"},
		{"mcp", "mcp"},
	}

	for _, e := range entities {
		t.Run(e.name+"_property_operations", func(t *testing.T) {
			entity, err := NewEntity(e.entityType, "test")
			if err != nil {
				t.Fatalf("NewEntity(%s) error = %v", e.entityType, err)
			}

			// Test GetProperty on empty entity
			val, ok := entity.GetProperty("nonexistent")
			if ok {
				t.Error("GetProperty should return false for nonexistent key")
			}
			if val != nil {
				t.Error("GetProperty should return nil for nonexistent key")
			}

			// Test SetProperty and GetProperty
			entity.SetProperty("key1", StringValue{Value: "value1"})
			entity.SetProperty("key2", NumberValue{Value: 42})

			val, ok = entity.GetProperty("key1")
			if !ok {
				t.Error("GetProperty should return true for existing key")
			}
			if sv, ok := val.(StringValue); !ok || sv.Value != "value1" {
				t.Errorf("GetProperty(key1) = %v, want value1", val)
			}

			val, ok = entity.GetProperty("key2")
			if !ok {
				t.Error("GetProperty should return true for existing key")
			}
			if nv, ok := val.(NumberValue); !ok || nv.Value != 42 {
				t.Errorf("GetProperty(key2) = %v, want 42", val)
			}

			// Test Properties
			props := entity.Properties()
			if len(props) != 2 {
				t.Errorf("Properties() returned %d items, want 2", len(props))
			}

			// Test overwriting property
			entity.SetProperty("key1", StringValue{Value: "updated"})
			val, _ = entity.GetProperty("key1")
			if sv, ok := val.(StringValue); !ok || sv.Value != "updated" {
				t.Errorf("SetProperty overwrite failed, got %v, want updated", val)
			}
		})
	}
}

func TestEntity_Metadata(t *testing.T) {
	entityTypes := []string{"file", "agent", "tool", "intent", "pipeline", "step", "trigger", "config", "mcp"}

	for _, entType := range entityTypes {
		t.Run(entType+"_metadata", func(t *testing.T) {
			entity, err := NewEntity(entType, "test")
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

func TestEntity_Location(t *testing.T) {
	entity, _ := NewEntity("agent", "test")

	// Default values should be 0
	if entity.Line() != 0 {
		t.Errorf("Entity.Line() default = %d, want 0", entity.Line())
	}
	if entity.Column() != 0 {
		t.Errorf("Entity.Column() default = %d, want 0", entity.Column())
	}

	// Set and verify location
	entity.SetLocation(10, 5)
	if entity.Line() != 10 {
		t.Errorf("Entity.Line() = %d, want 10", entity.Line())
	}
	if entity.Column() != 5 {
		t.Errorf("Entity.Column() = %d, want 5", entity.Column())
	}
}

func BenchmarkNewEntity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewEntity("agent", "test")
	}
}

func BenchmarkSetProperty(b *testing.B) {
	entity, _ := NewEntity("agent", "test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entity.SetProperty("model", StringValue{Value: "gpt-4"})
	}
}

func BenchmarkGetProperty(b *testing.B) {
	entity, _ := NewEntity("agent", "test")
	entity.SetProperty("model", StringValue{Value: "gpt-4"})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entity.GetProperty("model")
	}
}
