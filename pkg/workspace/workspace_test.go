package workspace

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/shellkjell/langspace/pkg/ast"
	"github.com/shellkjell/langspace/pkg/validator"
)

func createFileEntity(name string) ast.Entity {
	entity, _ := ast.NewEntity("file", name)
	entity.SetProperty("path", ast.StringValue{Value: "/path/to/" + name})
	return entity
}

func createAgentEntity(name string) ast.Entity {
	entity, _ := ast.NewEntity("agent", name)
	entity.SetProperty("model", ast.StringValue{Value: "gpt-4o"})
	entity.SetProperty("instruction", ast.StringValue{Value: "You are a helpful assistant."})
	return entity
}

func createToolEntity(name string) ast.Entity {
	entity, _ := ast.NewEntity("tool", name)
	entity.SetProperty("command", ast.StringValue{Value: "echo hello"})
	return entity
}

func TestWorkspace_AddEntity(t *testing.T) {
	tests := []struct {
		name      string
		entity    ast.Entity
		wantError bool
	}{
		{
			name:      "nil entity",
			entity:    nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := New()
			err := w.AddEntity(tt.entity)
			if (err != nil) != tt.wantError {
				t.Errorf("Workspace.AddEntity() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}

	// Test adding valid entities
	w := New()

	if err := w.AddEntity(createFileEntity("test.txt")); err != nil {
		t.Errorf("Workspace.AddEntity() error = %v for valid file entity", err)
	}

	if err := w.AddEntity(createAgentEntity("validator")); err != nil {
		t.Errorf("Workspace.AddEntity() error = %v for valid agent entity", err)
	}

	if err := w.AddEntity(createToolEntity("linter")); err != nil {
		t.Errorf("Workspace.AddEntity() error = %v for valid tool entity", err)
	}
}

func TestWorkspace_GetEntities(t *testing.T) {
	w := New()
	w.AddEntity(createFileEntity("test.txt"))

	entities := w.GetEntities()
	if len(entities) != 1 {
		t.Errorf("Workspace.GetEntities() returned %d entities, want 1", len(entities))
	}

	if entities[0].Type() != "file" {
		t.Errorf("Workspace.GetEntities()[0].Type() = %v, want file", entities[0].Type())
	}
}

func TestWorkspace_GetEntitiesByType(t *testing.T) {
	w := New()
	w.AddEntity(createFileEntity("test.txt"))
	w.AddEntity(createAgentEntity("validator"))
	w.AddEntity(createToolEntity("linter"))

	// Test getting file entities
	fileEntities := w.GetEntitiesByType("file")
	if len(fileEntities) != 1 {
		t.Errorf("Workspace.GetEntitiesByType(file) returned %d entities, want 1", len(fileEntities))
	}

	// Test getting agent entities
	agentEntities := w.GetEntitiesByType("agent")
	if len(agentEntities) != 1 {
		t.Errorf("Workspace.GetEntitiesByType(agent) returned %d entities, want 1", len(agentEntities))
	}

	// Test getting tool entities
	toolEntities := w.GetEntitiesByType("tool")
	if len(toolEntities) != 1 {
		t.Errorf("Workspace.GetEntitiesByType(tool) returned %d entities, want 1", len(toolEntities))
	}

	// Test getting non-existent type
	unknownEntities := w.GetEntitiesByType("unknown")
	if len(unknownEntities) != 0 {
		t.Errorf("Workspace.GetEntitiesByType(unknown) returned %d entities, want 0", len(unknownEntities))
	}
}

func TestWorkspace_GetEntityByName(t *testing.T) {
	w := New()
	w.AddEntity(createFileEntity("test.txt"))
	w.AddEntity(createAgentEntity("validator"))

	// Test finding existing entity
	entity, found := w.GetEntityByName("file", "test.txt")
	if !found {
		t.Error("GetEntityByName() should find existing entity")
	}
	if entity.Name() != "test.txt" {
		t.Errorf("GetEntityByName() returned entity with name %q, want test.txt", entity.Name())
	}

	// Test finding non-existent entity
	_, found = w.GetEntityByName("file", "nonexistent.txt")
	if found {
		t.Error("GetEntityByName() should not find non-existent entity")
	}
}

func TestWorkspace_Clear(t *testing.T) {
	w := New()
	w.AddEntity(createFileEntity("test.txt"))
	w.AddEntity(createAgentEntity("validator"))
	w.AddEntity(createToolEntity("linter"))

	// Clear the workspace
	w.Clear()

	// Verify it's empty
	entities := w.GetEntities()
	if len(entities) != 0 {
		t.Errorf("Workspace.Clear() did not clear entities, got %d entities", len(entities))
	}
}

func TestWorkspace_Concurrency(t *testing.T) {
	w := New()
	var wg sync.WaitGroup
	numGoroutines := 10

	// Test concurrent adding of entities
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			w.AddEntity(createFileEntity(fmt.Sprintf("test%d.txt", i)))
		}(i)
	}
	wg.Wait()

	// Verify all entities were added
	entities := w.GetEntities()
	if len(entities) != numGoroutines {
		t.Errorf("Concurrent AddEntity() resulted in %d entities, want %d", len(entities), numGoroutines)
	}
}

func BenchmarkWorkspace_AddEntity(b *testing.B) {
	w := New()
	entity := createFileEntity("test.txt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.AddEntity(entity)
	}
}

func BenchmarkWorkspace_GetEntities(b *testing.B) {
	w := New()
	w.AddEntity(createFileEntity("test.txt"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.GetEntities()
	}
}

// mockValidator is a test implementation of the Validator interface
type mockValidator struct {
	shouldError bool
	errorMsg    string
}

func (m *mockValidator) ValidateEntity(entity ast.Entity) error {
	if m.shouldError {
		return fmt.Errorf(m.errorMsg)
	}
	return nil
}

func TestWorkspace_WithValidator(t *testing.T) {
	tests := []struct {
		name        string
		validator   *mockValidator
		entity      ast.Entity
		wantError   bool
		errorPrefix string
	}{
		{
			name: "valid entity with validator",
			validator: &mockValidator{
				shouldError: false,
			},
			entity:    createFileEntity("test.txt"),
			wantError: false,
		},
		{
			name: "invalid entity with validator",
			validator: &mockValidator{
				shouldError: true,
				errorMsg:    "test validation error",
			},
			entity:      createFileEntity("test.txt"),
			wantError:   true,
			errorPrefix: "validation failed: test validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := New().WithValidator(tt.validator)
			err := w.AddEntity(tt.entity)

			if (err != nil) != tt.wantError {
				t.Errorf("AddEntity() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError && err != nil && tt.errorPrefix != "" {
				if !strings.HasPrefix(err.Error(), tt.errorPrefix) {
					t.Errorf("AddEntity() error = %v, want prefix %v", err, tt.errorPrefix)
				}
			}
		})
	}
}

func TestWorkspace_WithRealValidator(t *testing.T) {
	w := New().WithValidator(validator.New())

	// Test valid file entity
	if err := w.AddEntity(createFileEntity("test.txt")); err != nil {
		t.Errorf("AddEntity() with valid file should not error, got: %v", err)
	}

	// Test valid agent entity
	if err := w.AddEntity(createAgentEntity("validator")); err != nil {
		t.Errorf("AddEntity() with valid agent should not error, got: %v", err)
	}

	// Test valid tool entity
	if err := w.AddEntity(createToolEntity("linter")); err != nil {
		t.Errorf("AddEntity() with valid tool should not error, got: %v", err)
	}

	// Verify stats
	stats := w.Stat()
	if stats.TotalEntities != 3 {
		t.Errorf("Expected 3 total entities, got %d", stats.TotalEntities)
	}
	if stats.FileEntities != 1 {
		t.Errorf("Expected 1 file entity, got %d", stats.FileEntities)
	}
	if stats.AgentEntities != 1 {
		t.Errorf("Expected 1 agent entity, got %d", stats.AgentEntities)
	}
	if stats.ToolEntities != 1 {
		t.Errorf("Expected 1 tool entity, got %d", stats.ToolEntities)
	}
}

func TestWorkspace_Relationships(t *testing.T) {
	w := New()

	// Add entities
	w.AddEntity(createFileEntity("config.json"))
	w.AddEntity(createAgentEntity("validator"))
	w.AddEntity(createToolEntity("linter"))

	// Test adding a valid relationship
	err := w.AddRelationship("agent", "validator", "file", "config.json", RelationTypeAssigned)
	if err != nil {
		t.Errorf("AddRelationship() error = %v", err)
	}

	// Test adding a relationship to a non-existent entity
	err = w.AddRelationship("agent", "nonexistent", "file", "config.json", RelationTypeAssigned)
	if err == nil {
		t.Error("AddRelationship() should error for non-existent source entity")
	}

	err = w.AddRelationship("agent", "validator", "file", "nonexistent.txt", RelationTypeAssigned)
	if err == nil {
		t.Error("AddRelationship() should error for non-existent target entity")
	}

	// Test adding a duplicate relationship
	err = w.AddRelationship("agent", "validator", "file", "config.json", RelationTypeAssigned)
	if err == nil {
		t.Error("AddRelationship() should error for duplicate relationship")
	}

	// Add more relationships
	w.AddRelationship("tool", "linter", "file", "config.json", RelationTypeConsumes)
	w.AddRelationship("agent", "validator", "tool", "linter", RelationTypeAssigned)

	// Verify stats
	stats := w.Stat()
	if stats.TotalRelationships != 3 {
		t.Errorf("Expected 3 relationships, got %d", stats.TotalRelationships)
	}
}

func TestWorkspace_GetRelationships(t *testing.T) {
	w := New()

	// Add entities
	w.AddEntity(createFileEntity("config.json"))
	w.AddEntity(createAgentEntity("validator"))

	// Add relationship
	w.AddRelationship("agent", "validator", "file", "config.json", RelationTypeAssigned)

	// Get all relationships
	relationships := w.GetRelationships()
	if len(relationships) != 1 {
		t.Errorf("GetRelationships() returned %d relationships, want 1", len(relationships))
	}

	rel := relationships[0]
	if rel.SourceType != "agent" || rel.SourceName != "validator" ||
		rel.TargetType != "file" || rel.TargetName != "config.json" ||
		rel.Type != RelationTypeAssigned {
		t.Errorf("GetRelationships() returned unexpected relationship: %+v", rel)
	}
}

func TestWorkspace_GetRelationshipsForEntity(t *testing.T) {
	w := New()

	// Add entities
	w.AddEntity(createFileEntity("config.json"))
	w.AddEntity(createAgentEntity("validator"))
	w.AddEntity(createToolEntity("linter"))

	// Add relationships
	w.AddRelationship("agent", "validator", "file", "config.json", RelationTypeAssigned)
	w.AddRelationship("tool", "linter", "file", "config.json", RelationTypeConsumes)

	// Get relationships for file entity (should have 2)
	fileRels := w.GetRelationshipsForEntity("file", "config.json")
	if len(fileRels) != 2 {
		t.Errorf("GetRelationshipsForEntity() returned %d relationships, want 2", len(fileRels))
	}

	// Get relationships for agent entity (should have 1)
	agentRels := w.GetRelationshipsForEntity("agent", "validator")
	if len(agentRels) != 1 {
		t.Errorf("GetRelationshipsForEntity() returned %d relationships, want 1", len(agentRels))
	}
}

func TestWorkspace_GetRelatedEntities(t *testing.T) {
	w := New()

	// Add entities
	w.AddEntity(createFileEntity("config.json"))
	w.AddEntity(createAgentEntity("validator"))

	// Add relationship
	w.AddRelationship("agent", "validator", "file", "config.json", RelationTypeAssigned)

	// Get related entities from agent perspective
	related := w.GetRelatedEntities("agent", "validator", RelationTypeAssigned)
	if len(related) != 1 {
		t.Errorf("GetRelatedEntities() returned %d entities, want 1", len(related))
	}
	if related[0].Type() != "file" {
		t.Errorf("GetRelatedEntities() returned entity of type %s, want file", related[0].Type())
	}

	// Get related entities from file perspective
	related = w.GetRelatedEntities("file", "config.json", RelationTypeAssigned)
	if len(related) != 1 {
		t.Errorf("GetRelatedEntities() returned %d entities, want 1", len(related))
	}
	if related[0].Type() != "agent" {
		t.Errorf("GetRelatedEntities() returned entity of type %s, want agent", related[0].Type())
	}
}

func TestWorkspace_RemoveRelationship(t *testing.T) {
	w := New()

	// Add entities
	w.AddEntity(createFileEntity("config.json"))
	w.AddEntity(createAgentEntity("validator"))

	// Add relationship
	w.AddRelationship("agent", "validator", "file", "config.json", RelationTypeAssigned)

	// Verify relationship exists
	if len(w.GetRelationships()) != 1 {
		t.Error("Relationship should exist before removal")
	}

	// Remove relationship
	err := w.RemoveRelationship("agent", "validator", "file", "config.json", RelationTypeAssigned)
	if err != nil {
		t.Errorf("RemoveRelationship() error = %v", err)
	}

	// Verify relationship is removed
	if len(w.GetRelationships()) != 0 {
		t.Error("Relationship should be removed")
	}

	// Try to remove non-existent relationship
	err = w.RemoveRelationship("agent", "validator", "file", "config.json", RelationTypeAssigned)
	if err == nil {
		t.Error("RemoveRelationship() should error for non-existent relationship")
	}
}

func TestWorkspace_Clear_WithRelationships(t *testing.T) {
	w := New()

	// Add entities
	w.AddEntity(createFileEntity("config.json"))
	w.AddEntity(createAgentEntity("validator"))

	// Add relationship
	w.AddRelationship("agent", "validator", "file", "config.json", RelationTypeAssigned)

	// Clear workspace
	w.Clear()

	// Verify both entities and relationships are cleared
	if len(w.GetEntities()) != 0 {
		t.Error("Entities should be cleared")
	}
	if len(w.GetRelationships()) != 0 {
		t.Error("Relationships should be cleared")
	}
}

func TestWorkspace_EntityHooks(t *testing.T) {
	t.Run("before_add_hook_success", func(t *testing.T) {
		w := New()
		hookCalled := false

		w.OnEntityEvent(HookBeforeAdd, func(entity ast.Entity) error {
			hookCalled = true
			return nil
		})

		err := w.AddEntity(createFileEntity("test.txt"))
		if err != nil {
			t.Errorf("AddEntity() error = %v", err)
		}
		if !hookCalled {
			t.Error("Before-add hook was not called")
		}
	})

	t.Run("before_add_hook_blocks", func(t *testing.T) {
		w := New()

		w.OnEntityEvent(HookBeforeAdd, func(entity ast.Entity) error {
			return fmt.Errorf("blocked by hook")
		})

		err := w.AddEntity(createFileEntity("test.txt"))
		if err == nil {
			t.Error("AddEntity() should have failed due to hook")
		}
		if !strings.Contains(err.Error(), "blocked by hook") {
			t.Errorf("Error should mention hook failure, got: %v", err)
		}

		// Entity should not be added
		if len(w.GetEntities()) != 0 {
			t.Error("Entity should not be added when hook fails")
		}
	})

	t.Run("after_add_hook_called", func(t *testing.T) {
		w := New()
		hookCalled := false

		w.OnEntityEvent(HookAfterAdd, func(entity ast.Entity) error {
			hookCalled = true
			return nil
		})

		err := w.AddEntity(createFileEntity("test.txt"))
		if err != nil {
			t.Errorf("AddEntity() error = %v", err)
		}
		if !hookCalled {
			t.Error("After-add hook was not called")
		}
	})

	t.Run("multiple_hooks", func(t *testing.T) {
		w := New()
		callOrder := []string{}

		w.OnEntityEvent(HookBeforeAdd, func(entity ast.Entity) error {
			callOrder = append(callOrder, "before1")
			return nil
		})
		w.OnEntityEvent(HookBeforeAdd, func(entity ast.Entity) error {
			callOrder = append(callOrder, "before2")
			return nil
		})
		w.OnEntityEvent(HookAfterAdd, func(entity ast.Entity) error {
			callOrder = append(callOrder, "after1")
			return nil
		})

		w.AddEntity(createFileEntity("test.txt"))

		expected := []string{"before1", "before2", "after1"}
		if len(callOrder) != len(expected) {
			t.Errorf("Expected %d hook calls, got %d", len(expected), len(callOrder))
		}
		for i, v := range expected {
			if callOrder[i] != v {
				t.Errorf("Hook call order[%d] = %s, want %s", i, callOrder[i], v)
			}
		}
	})
}

func TestWorkspace_RemoveEntity(t *testing.T) {
	t.Run("remove_existing_entity", func(t *testing.T) {
		w := New()
		w.AddEntity(createFileEntity("test.txt"))

		err := w.RemoveEntity("file", "test.txt")
		if err != nil {
			t.Errorf("RemoveEntity() error = %v", err)
		}
		if len(w.GetEntities()) != 0 {
			t.Error("Entity should be removed")
		}
	})

	t.Run("remove_nonexistent_entity", func(t *testing.T) {
		w := New()

		err := w.RemoveEntity("file", "nonexistent.txt")
		if err == nil {
			t.Error("RemoveEntity() should fail for nonexistent entity")
		}
	})

	t.Run("remove_entity_clears_relationships", func(t *testing.T) {
		w := New()

		// Add entities
		w.AddEntity(createFileEntity("config.json"))
		w.AddEntity(createAgentEntity("validator"))

		// Add relationship
		w.AddRelationship("agent", "validator", "file", "config.json", RelationTypeAssigned)

		// Verify relationship exists
		if len(w.GetRelationships()) != 1 {
			t.Error("Relationship should exist before removal")
		}

		// Remove entity
		w.RemoveEntity("file", "config.json")

		// Verify relationship is also removed
		if len(w.GetRelationships()) != 0 {
			t.Error("Relationship should be removed when entity is removed")
		}
	})

	t.Run("remove_entity_with_hooks", func(t *testing.T) {
		w := New()
		beforeCalled := false
		afterCalled := false

		w.OnEntityEvent(HookBeforeRemove, func(entity ast.Entity) error {
			beforeCalled = true
			return nil
		})
		w.OnEntityEvent(HookAfterRemove, func(entity ast.Entity) error {
			afterCalled = true
			return nil
		})

		w.AddEntity(createFileEntity("test.txt"))
		w.RemoveEntity("file", "test.txt")

		if !beforeCalled {
			t.Error("Before-remove hook was not called")
		}
		if !afterCalled {
			t.Error("After-remove hook was not called")
		}
	})

	t.Run("before_remove_hook_blocks", func(t *testing.T) {
		w := New()

		w.OnEntityEvent(HookBeforeRemove, func(entity ast.Entity) error {
			return fmt.Errorf("removal blocked")
		})

		w.AddEntity(createFileEntity("test.txt"))

		err := w.RemoveEntity("file", "test.txt")
		if err == nil {
			t.Error("RemoveEntity() should fail due to hook")
		}

		// Entity should still exist
		if len(w.GetEntities()) != 1 {
			t.Error("Entity should not be removed when hook fails")
		}
	})
}

func TestWorkspace_Stats_WithHooks(t *testing.T) {
	w := New()

	w.OnEntityEvent(HookBeforeAdd, func(entity ast.Entity) error { return nil })
	w.OnEntityEvent(HookAfterAdd, func(entity ast.Entity) error { return nil })
	w.OnEntityEvent(HookBeforeRemove, func(entity ast.Entity) error { return nil })

	stats := w.Stat()
	if stats.TotalHooks != 3 {
		t.Errorf("Expected 3 hooks, got %d", stats.TotalHooks)
	}
}
