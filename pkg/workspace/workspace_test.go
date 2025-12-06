package workspace

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
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

func TestWorkspace_UpdateEntity(t *testing.T) {
	t.Run("update_existing_entity", func(t *testing.T) {
		w := New()
		original := createFileEntity("test.txt")
		w.AddEntity(original)

		// Create updated version
		updated, _ := ast.NewEntity("file", "test.txt")
		updated.SetProperty("path", ast.StringValue{Value: "/new/path/to/test.txt"})

		err := w.UpdateEntity(updated)
		if err != nil {
			t.Errorf("UpdateEntity() error = %v", err)
		}

		// Verify entity was updated
		entity, found := w.GetEntityByName("file", "test.txt")
		if !found {
			t.Fatal("Entity should exist after update")
		}
		path, _ := entity.GetProperty("path")
		if path.(ast.StringValue).Value != "/new/path/to/test.txt" {
			t.Errorf("Property not updated: got %v", path)
		}
	})

	t.Run("update_nonexistent_entity", func(t *testing.T) {
		w := New()
		entity := createFileEntity("test.txt")

		err := w.UpdateEntity(entity)
		if err == nil {
			t.Error("UpdateEntity() should fail for nonexistent entity")
		}
	})

	t.Run("update_nil_entity", func(t *testing.T) {
		w := New()

		err := w.UpdateEntity(nil)
		if err == nil {
			t.Error("UpdateEntity() should fail for nil entity")
		}
	})

	t.Run("update_with_hooks", func(t *testing.T) {
		w := New()
		original := createFileEntity("test.txt")
		w.AddEntity(original)

		beforeCalled := false
		afterCalled := false

		w.OnEntityEvent(HookBeforeUpdate, func(entity ast.Entity) error {
			beforeCalled = true
			return nil
		})
		w.OnEntityEvent(HookAfterUpdate, func(entity ast.Entity) error {
			afterCalled = true
			return nil
		})

		updated, _ := ast.NewEntity("file", "test.txt")
		updated.SetProperty("path", ast.StringValue{Value: "/new/path"})

		err := w.UpdateEntity(updated)
		if err != nil {
			t.Errorf("UpdateEntity() error = %v", err)
		}

		if !beforeCalled {
			t.Error("Before-update hook was not called")
		}
		if !afterCalled {
			t.Error("After-update hook was not called")
		}
	})

	t.Run("before_update_hook_blocks", func(t *testing.T) {
		w := New()
		original := createFileEntity("test.txt")
		w.AddEntity(original)

		w.OnEntityEvent(HookBeforeUpdate, func(entity ast.Entity) error {
			return fmt.Errorf("update blocked")
		})

		updated, _ := ast.NewEntity("file", "test.txt")
		updated.SetProperty("path", ast.StringValue{Value: "/new/path"})

		err := w.UpdateEntity(updated)
		if err == nil {
			t.Error("UpdateEntity() should fail due to hook")
		}

		// Entity should have original value
		entity, _ := w.GetEntityByName("file", "test.txt")
		path, _ := entity.GetProperty("path")
		if path.(ast.StringValue).Value != "/path/to/test.txt" {
			t.Error("Entity should not be updated when hook fails")
		}
	})

	t.Run("update_with_validator", func(t *testing.T) {
		w := New().WithValidator(validator.New())
		original := createAgentEntity("assistant")
		w.AddEntity(original)

		// Create invalid entity (missing required model)
		invalid, _ := ast.NewEntity("agent", "assistant")
		// Don't set model property

		err := w.UpdateEntity(invalid)
		if err == nil {
			t.Error("UpdateEntity() should fail validation")
		}
	})
}

func TestWorkspace_UpsertEntity(t *testing.T) {
	t.Run("upsert_adds_new_entity", func(t *testing.T) {
		w := New()
		entity := createFileEntity("test.txt")

		err := w.UpsertEntity(entity)
		if err != nil {
			t.Errorf("UpsertEntity() error = %v", err)
		}

		if len(w.GetEntities()) != 1 {
			t.Errorf("Expected 1 entity, got %d", len(w.GetEntities()))
		}
	})

	t.Run("upsert_updates_existing_entity", func(t *testing.T) {
		w := New()
		original := createFileEntity("test.txt")
		w.AddEntity(original)

		// Create updated version
		updated, _ := ast.NewEntity("file", "test.txt")
		updated.SetProperty("path", ast.StringValue{Value: "/updated/path"})

		err := w.UpsertEntity(updated)
		if err != nil {
			t.Errorf("UpsertEntity() error = %v", err)
		}

		// Should still have only 1 entity
		if len(w.GetEntities()) != 1 {
			t.Errorf("Expected 1 entity after upsert, got %d", len(w.GetEntities()))
		}

		// Verify entity was updated
		entity, found := w.GetEntityByName("file", "test.txt")
		if !found {
			t.Fatal("Entity should exist after upsert")
		}
		path, _ := entity.GetProperty("path")
		if path.(ast.StringValue).Value != "/updated/path" {
			t.Errorf("Property not updated: got %v", path)
		}
	})

	t.Run("upsert_nil_entity", func(t *testing.T) {
		w := New()

		err := w.UpsertEntity(nil)
		if err == nil {
			t.Error("UpsertEntity() should fail for nil entity")
		}
	})

	t.Run("upsert_add_calls_add_hooks", func(t *testing.T) {
		w := New()
		beforeAddCalled := false
		afterAddCalled := false

		w.OnEntityEvent(HookBeforeAdd, func(entity ast.Entity) error {
			beforeAddCalled = true
			return nil
		})
		w.OnEntityEvent(HookAfterAdd, func(entity ast.Entity) error {
			afterAddCalled = true
			return nil
		})

		entity := createFileEntity("test.txt")
		w.UpsertEntity(entity)

		if !beforeAddCalled {
			t.Error("Before-add hook should be called for new entity")
		}
		if !afterAddCalled {
			t.Error("After-add hook should be called for new entity")
		}
	})

	t.Run("upsert_update_calls_update_hooks", func(t *testing.T) {
		w := New()
		original := createFileEntity("test.txt")
		w.AddEntity(original)

		beforeUpdateCalled := false
		afterUpdateCalled := false

		w.OnEntityEvent(HookBeforeUpdate, func(entity ast.Entity) error {
			beforeUpdateCalled = true
			return nil
		})
		w.OnEntityEvent(HookAfterUpdate, func(entity ast.Entity) error {
			afterUpdateCalled = true
			return nil
		})

		updated, _ := ast.NewEntity("file", "test.txt")
		updated.SetProperty("path", ast.StringValue{Value: "/new/path"})
		w.UpsertEntity(updated)

		if !beforeUpdateCalled {
			t.Error("Before-update hook should be called for existing entity")
		}
		if !afterUpdateCalled {
			t.Error("After-update hook should be called for existing entity")
		}
	})
}

func TestWorkspace_Events(t *testing.T) {
	t.Run("entity_added_event", func(t *testing.T) {
		w := New()
		var receivedEvent Event

		w.OnEvent(func(event Event) {
			receivedEvent = event
		})

		entity := createFileEntity("test.txt")
		w.AddEntity(entity)

		if receivedEvent.Type != EventEntityAdded {
			t.Errorf("Expected EventEntityAdded, got %v", receivedEvent.Type)
		}
		if receivedEvent.Entity == nil || receivedEvent.Entity.Name() != "test.txt" {
			t.Error("Event should contain the added entity")
		}
	})

	t.Run("entity_removed_event", func(t *testing.T) {
		w := New()
		entity := createFileEntity("test.txt")
		w.AddEntity(entity)

		var receivedEvent Event
		w.OnEvent(func(event Event) {
			receivedEvent = event
		})

		w.RemoveEntity("file", "test.txt")

		if receivedEvent.Type != EventEntityRemoved {
			t.Errorf("Expected EventEntityRemoved, got %v", receivedEvent.Type)
		}
		if receivedEvent.Entity == nil || receivedEvent.Entity.Name() != "test.txt" {
			t.Error("Event should contain the removed entity")
		}
	})

	t.Run("entity_updated_event", func(t *testing.T) {
		w := New()
		original := createFileEntity("test.txt")
		w.AddEntity(original)

		var receivedEvent Event
		w.OnEvent(func(event Event) {
			receivedEvent = event
		})

		updated, _ := ast.NewEntity("file", "test.txt")
		updated.SetProperty("path", ast.StringValue{Value: "/new/path"})
		w.UpdateEntity(updated)

		if receivedEvent.Type != EventEntityUpdated {
			t.Errorf("Expected EventEntityUpdated, got %v", receivedEvent.Type)
		}
		if receivedEvent.Entity == nil || receivedEvent.Entity.Name() != "test.txt" {
			t.Error("Event should contain the updated entity")
		}
	})

	t.Run("workspace_cleared_event", func(t *testing.T) {
		w := New()
		w.AddEntity(createFileEntity("test.txt"))

		var receivedEvent Event
		w.OnEvent(func(event Event) {
			receivedEvent = event
		})

		w.Clear()

		if receivedEvent.Type != EventWorkspaceCleared {
			t.Errorf("Expected EventWorkspaceCleared, got %v", receivedEvent.Type)
		}
	})

	t.Run("relationship_added_event", func(t *testing.T) {
		w := New()
		w.AddEntity(createFileEntity("test.txt"))
		w.AddEntity(createAgentEntity("assistant"))

		var receivedEvent Event
		w.OnEvent(func(event Event) {
			receivedEvent = event
		})

		w.AddRelationship("agent", "assistant", "file", "test.txt", RelationTypeAssigned)

		if receivedEvent.Type != EventRelationshipAdded {
			t.Errorf("Expected EventRelationshipAdded, got %v", receivedEvent.Type)
		}
		if receivedEvent.Relationship == nil {
			t.Error("Event should contain the added relationship")
		}
		if receivedEvent.Relationship.SourceName != "assistant" {
			t.Errorf("Relationship source should be 'assistant', got %s", receivedEvent.Relationship.SourceName)
		}
	})

	t.Run("relationship_removed_event", func(t *testing.T) {
		w := New()
		w.AddEntity(createFileEntity("test.txt"))
		w.AddEntity(createAgentEntity("assistant"))
		w.AddRelationship("agent", "assistant", "file", "test.txt", RelationTypeAssigned)

		var receivedEvent Event
		w.OnEvent(func(event Event) {
			receivedEvent = event
		})

		w.RemoveRelationship("agent", "assistant", "file", "test.txt", RelationTypeAssigned)

		if receivedEvent.Type != EventRelationshipRemoved {
			t.Errorf("Expected EventRelationshipRemoved, got %v", receivedEvent.Type)
		}
		if receivedEvent.Relationship == nil {
			t.Error("Event should contain the removed relationship")
		}
	})

	t.Run("multiple_event_handlers", func(t *testing.T) {
		w := New()
		callCount := 0

		w.OnEvent(func(event Event) {
			callCount++
		})
		w.OnEvent(func(event Event) {
			callCount++
		})

		w.AddEntity(createFileEntity("test.txt"))

		if callCount != 2 {
			t.Errorf("Expected 2 event handler calls, got %d", callCount)
		}
	})

	t.Run("events_track_all_operations", func(t *testing.T) {
		w := New()
		events := make([]EventType, 0)

		w.OnEvent(func(event Event) {
			events = append(events, event.Type)
		})

		// Perform various operations
		w.AddEntity(createFileEntity("file1.txt"))
		w.AddEntity(createAgentEntity("agent1"))
		w.AddRelationship("agent", "agent1", "file", "file1.txt", RelationTypeAssigned)

		updated, _ := ast.NewEntity("file", "file1.txt")
		updated.SetProperty("path", ast.StringValue{Value: "/new/path"})
		w.UpdateEntity(updated)

		w.RemoveRelationship("agent", "agent1", "file", "file1.txt", RelationTypeAssigned)
		w.RemoveEntity("file", "file1.txt")
		w.Clear()

		expectedEvents := []EventType{
			EventEntityAdded,         // file1.txt
			EventEntityAdded,         // agent1
			EventRelationshipAdded,   // relationship
			EventEntityUpdated,       // file1.txt update
			EventRelationshipRemoved, // relationship
			EventEntityRemoved,       // file1.txt
			EventWorkspaceCleared,    // clear
		}

		if len(events) != len(expectedEvents) {
			t.Errorf("Expected %d events, got %d", len(expectedEvents), len(events))
		}

		for i, expected := range expectedEvents {
			if i < len(events) && events[i] != expected {
				t.Errorf("Event %d: expected %v, got %v", i, expected, events[i])
			}
		}
	})

	t.Run("upsert_emits_correct_event", func(t *testing.T) {
		w := New()
		events := make([]EventType, 0)

		w.OnEvent(func(event Event) {
			events = append(events, event.Type)
		})

		// First upsert should add
		entity := createFileEntity("test.txt")
		w.UpsertEntity(entity)

		if len(events) != 1 || events[0] != EventEntityAdded {
			t.Error("First upsert should emit EntityAdded event")
		}

		// Second upsert should update
		updated, _ := ast.NewEntity("file", "test.txt")
		updated.SetProperty("path", ast.StringValue{Value: "/new/path"})
		w.UpsertEntity(updated)

		if len(events) != 2 || events[1] != EventEntityUpdated {
			t.Error("Second upsert should emit EntityUpdated event")
		}
	})
}

func TestWorkspace_Versioning(t *testing.T) {
	t.Run("versioning_disabled_by_default", func(t *testing.T) {
		w := New()
		w.AddEntity(createFileEntity("test.txt"))

		count := w.GetEntityVersionCount("file", "test.txt")
		if count != 0 {
			t.Errorf("Expected 0 versions when versioning disabled, got %d", count)
		}
	})

	t.Run("versioning_tracks_add", func(t *testing.T) {
		w := New().WithVersioning()
		w.AddEntity(createFileEntity("test.txt"))

		count := w.GetEntityVersionCount("file", "test.txt")
		if count != 1 {
			t.Errorf("Expected 1 version after add, got %d", count)
		}

		entity, found := w.GetEntityVersion("file", "test.txt", 1)
		if !found {
			t.Fatal("Version 1 should exist")
		}
		if entity.Name() != "test.txt" {
			t.Errorf("Expected entity name 'test.txt', got %s", entity.Name())
		}
	})

	t.Run("versioning_tracks_update", func(t *testing.T) {
		w := New().WithVersioning()
		original := createFileEntity("test.txt")
		w.AddEntity(original)

		// Update the entity
		updated, _ := ast.NewEntity("file", "test.txt")
		updated.SetProperty("path", ast.StringValue{Value: "/updated/path"})
		w.UpdateEntity(updated)

		count := w.GetEntityVersionCount("file", "test.txt")
		if count != 2 {
			t.Errorf("Expected 2 versions after update, got %d", count)
		}

		// Check version 1 has original path
		v1, _ := w.GetEntityVersion("file", "test.txt", 1)
		path1, _ := v1.GetProperty("path")
		if path1.(ast.StringValue).Value != "/path/to/test.txt" {
			t.Errorf("Version 1 should have original path")
		}

		// Check version 2 has updated path
		v2, _ := w.GetEntityVersion("file", "test.txt", 2)
		path2, _ := v2.GetProperty("path")
		if path2.(ast.StringValue).Value != "/updated/path" {
			t.Errorf("Version 2 should have updated path")
		}
	})

	t.Run("versioning_tracks_upsert", func(t *testing.T) {
		w := New().WithVersioning()

		// First upsert (add)
		entity := createFileEntity("test.txt")
		w.UpsertEntity(entity)

		count := w.GetEntityVersionCount("file", "test.txt")
		if count != 1 {
			t.Errorf("Expected 1 version after upsert add, got %d", count)
		}

		// Second upsert (update)
		updated, _ := ast.NewEntity("file", "test.txt")
		updated.SetProperty("path", ast.StringValue{Value: "/upserted/path"})
		w.UpsertEntity(updated)

		count = w.GetEntityVersionCount("file", "test.txt")
		if count != 2 {
			t.Errorf("Expected 2 versions after upsert update, got %d", count)
		}
	})

	t.Run("get_entity_version_invalid", func(t *testing.T) {
		w := New().WithVersioning()
		w.AddEntity(createFileEntity("test.txt"))

		// Version 0 doesn't exist
		_, found := w.GetEntityVersion("file", "test.txt", 0)
		if found {
			t.Error("Version 0 should not exist")
		}

		// Version 2 doesn't exist yet
		_, found = w.GetEntityVersion("file", "test.txt", 2)
		if found {
			t.Error("Version 2 should not exist")
		}

		// Non-existent entity
		_, found = w.GetEntityVersion("file", "nonexistent.txt", 1)
		if found {
			t.Error("Non-existent entity should not have versions")
		}
	})

	t.Run("get_entity_history", func(t *testing.T) {
		w := New().WithVersioning()
		w.AddEntity(createFileEntity("test.txt"))

		updated, _ := ast.NewEntity("file", "test.txt")
		updated.SetProperty("path", ast.StringValue{Value: "/v2/path"})
		w.UpdateEntity(updated)

		updated2, _ := ast.NewEntity("file", "test.txt")
		updated2.SetProperty("path", ast.StringValue{Value: "/v3/path"})
		w.UpdateEntity(updated2)

		history := w.GetEntityHistory("file", "test.txt")
		if len(history) != 3 {
			t.Fatalf("Expected 3 versions in history, got %d", len(history))
		}

		if history[0].Version != 1 || history[1].Version != 2 || history[2].Version != 3 {
			t.Error("Versions should be numbered 1, 2, 3")
		}

		// Verify timestamps are set
		for _, v := range history {
			if v.Timestamp == 0 {
				t.Error("Timestamp should be set")
			}
		}
	})

	t.Run("get_entity_history_nonexistent", func(t *testing.T) {
		w := New().WithVersioning()

		history := w.GetEntityHistory("file", "nonexistent.txt")
		if history != nil {
			t.Error("Non-existent entity should have nil history")
		}
	})

	t.Run("versioning_multiple_entities", func(t *testing.T) {
		w := New().WithVersioning()

		w.AddEntity(createFileEntity("file1.txt"))
		w.AddEntity(createFileEntity("file2.txt"))
		w.AddEntity(createAgentEntity("agent1"))

		if w.GetEntityVersionCount("file", "file1.txt") != 1 {
			t.Error("file1.txt should have 1 version")
		}
		if w.GetEntityVersionCount("file", "file2.txt") != 1 {
			t.Error("file2.txt should have 1 version")
		}
		if w.GetEntityVersionCount("agent", "agent1") != 1 {
			t.Error("agent1 should have 1 version")
		}

		// Update file1 twice
		for i := 0; i < 2; i++ {
			updated, _ := ast.NewEntity("file", "file1.txt")
			updated.SetProperty("path", ast.StringValue{Value: fmt.Sprintf("/v%d", i+2)})
			w.UpdateEntity(updated)
		}

		if w.GetEntityVersionCount("file", "file1.txt") != 3 {
			t.Error("file1.txt should have 3 versions after 2 updates")
		}
		if w.GetEntityVersionCount("file", "file2.txt") != 1 {
			t.Error("file2.txt should still have 1 version")
		}
	})
}

func TestWorkspace_Persistence(t *testing.T) {
	t.Run("serialize_empty_workspace", func(t *testing.T) {
		w := New()
		sw, err := w.Serialize()
		if err != nil {
			t.Fatalf("Serialize() error = %v", err)
		}
		if sw.Version != 1 {
			t.Errorf("Expected version 1, got %d", sw.Version)
		}
		if len(sw.Entities) != 0 {
			t.Errorf("Expected 0 entities, got %d", len(sw.Entities))
		}
		if len(sw.Relationships) != 0 {
			t.Errorf("Expected 0 relationships, got %d", len(sw.Relationships))
		}
	})

	t.Run("serialize_with_entities", func(t *testing.T) {
		w := New()
		w.AddEntity(createFileEntity("test.txt"))
		w.AddEntity(createAgentEntity("assistant"))

		sw, err := w.Serialize()
		if err != nil {
			t.Fatalf("Serialize() error = %v", err)
		}
		if len(sw.Entities) != 2 {
			t.Errorf("Expected 2 entities, got %d", len(sw.Entities))
		}
	})

	t.Run("serialize_with_relationships", func(t *testing.T) {
		w := New()
		w.AddEntity(createFileEntity("test.txt"))
		w.AddEntity(createAgentEntity("assistant"))
		w.AddRelationship("agent", "assistant", "file", "test.txt", RelationTypeAssigned)

		sw, err := w.Serialize()
		if err != nil {
			t.Fatalf("Serialize() error = %v", err)
		}
		if len(sw.Relationships) != 1 {
			t.Errorf("Expected 1 relationship, got %d", len(sw.Relationships))
		}
	})

	t.Run("save_and_load_roundtrip", func(t *testing.T) {
		w := New()
		w.AddEntity(createFileEntity("test.txt"))
		w.AddEntity(createAgentEntity("assistant"))
		w.AddRelationship("agent", "assistant", "file", "test.txt", RelationTypeAssigned)

		// Save to buffer
		var buf strings.Builder
		err := w.SaveTo(&buf)
		if err != nil {
			t.Fatalf("SaveTo() error = %v", err)
		}

		// Load into new workspace
		w2 := New()
		err = w2.LoadFrom(strings.NewReader(buf.String()))
		if err != nil {
			t.Fatalf("LoadFrom() error = %v", err)
		}

		// Verify entities
		if len(w2.GetEntities()) != 2 {
			t.Errorf("Expected 2 entities after load, got %d", len(w2.GetEntities()))
		}

		// Verify specific entity
		file, found := w2.GetEntityByName("file", "test.txt")
		if !found {
			t.Fatal("File entity not found after load")
		}
		path, _ := file.GetProperty("path")
		if path.(ast.StringValue).Value != "/path/to/test.txt" {
			t.Errorf("Property not preserved: got %v", path)
		}

		// Verify relationships
		if len(w2.GetRelationships()) != 1 {
			t.Errorf("Expected 1 relationship after load, got %d", len(w2.GetRelationships()))
		}
	})

	t.Run("save_and_load_file", func(t *testing.T) {
		w := New()
		w.AddEntity(createFileEntity("test.txt"))

		// Create temp file
		tmpFile, err := os.CreateTemp("", "workspace_test_*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath)

		// Save to file
		err = w.SaveToFile(tmpPath)
		if err != nil {
			t.Fatalf("SaveToFile() error = %v", err)
		}

		// Load from file
		w2 := New()
		err = w2.LoadFromFile(tmpPath)
		if err != nil {
			t.Fatalf("LoadFromFile() error = %v", err)
		}

		if len(w2.GetEntities()) != 1 {
			t.Errorf("Expected 1 entity after load, got %d", len(w2.GetEntities()))
		}
	})

	t.Run("load_preserves_metadata", func(t *testing.T) {
		w := New()
		entity := createFileEntity("test.txt")
		entity.SetMetadata("author", "test-user")
		entity.SetMetadata("version", "1.0")
		w.AddEntity(entity)

		var buf strings.Builder
		w.SaveTo(&buf)

		w2 := New()
		w2.LoadFrom(strings.NewReader(buf.String()))

		loaded, _ := w2.GetEntityByName("file", "test.txt")
		author, found := loaded.GetMetadata("author")
		if !found || author != "test-user" {
			t.Errorf("Metadata not preserved: author = %q", author)
		}
	})

	t.Run("load_clears_existing_data", func(t *testing.T) {
		w := New()
		w.AddEntity(createFileEntity("existing.txt"))

		// Create workspace with different data
		w2 := New()
		w2.AddEntity(createAgentEntity("new-agent"))

		var buf strings.Builder
		w2.SaveTo(&buf)

		// Load into w - should replace existing data
		w.LoadFrom(strings.NewReader(buf.String()))

		entities := w.GetEntities()
		if len(entities) != 1 {
			t.Fatalf("Expected 1 entity, got %d", len(entities))
		}
		if entities[0].Type() != "agent" || entities[0].Name() != "new-agent" {
			t.Error("Existing data should be replaced by loaded data")
		}
	})

	t.Run("load_preserves_hooks", func(t *testing.T) {
		w := New()
		hookCalled := false
		w.OnEntityEvent(HookAfterAdd, func(entity ast.Entity) error {
			hookCalled = true
			return nil
		})

		// Create source workspace
		w2 := New()
		w2.AddEntity(createFileEntity("test.txt"))

		var buf strings.Builder
		w2.SaveTo(&buf)

		// Load into w - hooks should be preserved
		w.LoadFrom(strings.NewReader(buf.String()))

		// Add a new entity - hook should fire
		w.AddEntity(createFileEntity("new.txt"))
		if !hookCalled {
			t.Error("Hooks should be preserved after load")
		}
	})

	t.Run("serialize_various_value_types", func(t *testing.T) {
		w := New()
		entity, _ := ast.NewEntity("file", "complex.txt")
		entity.SetProperty("string_prop", ast.StringValue{Value: "hello"})
		entity.SetProperty("number_prop", ast.NumberValue{Value: 42.5})
		entity.SetProperty("bool_prop", ast.BoolValue{Value: true})
		entity.SetProperty("array_prop", ast.ArrayValue{
			Elements: []ast.Value{
				ast.StringValue{Value: "a"},
				ast.StringValue{Value: "b"},
			},
		})
		entity.SetProperty("ref_prop", ast.ReferenceValue{
			Type: "agent",
			Name: "test",
			Path: []string{"output"},
		})
		entity.SetProperty("var_prop", ast.VariableValue{Name: "input"})
		w.AddEntity(entity)

		var buf strings.Builder
		err := w.SaveTo(&buf)
		if err != nil {
			t.Fatalf("SaveTo() error = %v", err)
		}

		w2 := New()
		err = w2.LoadFrom(strings.NewReader(buf.String()))
		if err != nil {
			t.Fatalf("LoadFrom() error = %v", err)
		}

		loaded, _ := w2.GetEntityByName("file", "complex.txt")

		// Verify string
		strVal, _ := loaded.GetProperty("string_prop")
		if strVal.(ast.StringValue).Value != "hello" {
			t.Error("String property not preserved")
		}

		// Verify number
		numVal, _ := loaded.GetProperty("number_prop")
		if numVal.(ast.NumberValue).Value != 42.5 {
			t.Error("Number property not preserved")
		}

		// Verify bool
		boolVal, _ := loaded.GetProperty("bool_prop")
		if boolVal.(ast.BoolValue).Value != true {
			t.Error("Bool property not preserved")
		}

		// Verify array
		arrVal, _ := loaded.GetProperty("array_prop")
		arr := arrVal.(ast.ArrayValue)
		if len(arr.Elements) != 2 {
			t.Error("Array property not preserved")
		}

		// Verify reference
		refVal, _ := loaded.GetProperty("ref_prop")
		ref := refVal.(ast.ReferenceValue)
		if ref.Type != "agent" || ref.Name != "test" {
			t.Error("Reference property not preserved")
		}

		// Verify variable
		varVal, _ := loaded.GetProperty("var_prop")
		if varVal.(ast.VariableValue).Name != "input" {
			t.Error("Variable property not preserved")
		}
	})

	t.Run("load_invalid_json", func(t *testing.T) {
		w := New()
		err := w.LoadFrom(strings.NewReader("invalid json"))
		if err == nil {
			t.Error("LoadFrom should fail on invalid JSON")
		}
	})

	t.Run("load_unknown_entity_type", func(t *testing.T) {
		w := New()
		json := `{"version":1,"entities":[{"type":"unknown","name":"test","properties":{}}],"relationships":[]}`
		err := w.LoadFrom(strings.NewReader(json))
		if err == nil {
			t.Error("LoadFrom should fail on unknown entity type")
		}
	})
}

func TestWorkspace_Snapshots(t *testing.T) {
	t.Run("create_snapshot", func(t *testing.T) {
		w := New()
		w.AddEntity(createFileEntity("test.txt"))

		snapshot, err := w.CreateSnapshot("v1")
		if err != nil {
			t.Fatalf("CreateSnapshot() error = %v", err)
		}

		if snapshot.ID != "v1" {
			t.Errorf("Snapshot ID should be 'v1', got %s", snapshot.ID)
		}
		if snapshot.Timestamp == 0 {
			t.Error("Snapshot should have a timestamp")
		}
		if snapshot.Data == nil {
			t.Error("Snapshot should have data")
		}
		if len(snapshot.Data.Entities) != 1 {
			t.Errorf("Snapshot should have 1 entity, got %d", len(snapshot.Data.Entities))
		}
	})

	t.Run("restore_snapshot", func(t *testing.T) {
		w := New()
		w.AddEntity(createFileEntity("original.txt"))

		// Create snapshot
		snapshot, _ := w.CreateSnapshot("before-changes")

		// Modify workspace
		w.RemoveEntity("file", "original.txt")
		w.AddEntity(createAgentEntity("new-agent"))

		// Verify workspace changed
		if len(w.GetEntitiesByType("file")) != 0 {
			t.Error("File should be removed")
		}
		if len(w.GetEntitiesByType("agent")) != 1 {
			t.Error("Agent should be added")
		}

		// Restore snapshot
		err := w.RestoreSnapshot(snapshot)
		if err != nil {
			t.Fatalf("RestoreSnapshot() error = %v", err)
		}

		// Verify workspace restored
		files := w.GetEntitiesByType("file")
		if len(files) != 1 || files[0].Name() != "original.txt" {
			t.Error("Workspace should be restored to snapshot state")
		}
		if len(w.GetEntitiesByType("agent")) != 0 {
			t.Error("Agent should not exist after restore")
		}
	})

	t.Run("restore_nil_snapshot", func(t *testing.T) {
		w := New()
		err := w.RestoreSnapshot(nil)
		if err == nil {
			t.Error("RestoreSnapshot(nil) should fail")
		}
	})

	t.Run("restore_snapshot_nil_data", func(t *testing.T) {
		w := New()
		snapshot := &Snapshot{ID: "empty", Timestamp: 1234567890}
		err := w.RestoreSnapshot(snapshot)
		if err == nil {
			t.Error("RestoreSnapshot with nil data should fail")
		}
	})

	t.Run("snapshot_preserves_relationships", func(t *testing.T) {
		w := New()
		w.AddEntity(createFileEntity("test.txt"))
		w.AddEntity(createAgentEntity("assistant"))
		w.AddRelationship("agent", "assistant", "file", "test.txt", RelationTypeAssigned)

		// Create snapshot
		snapshot, _ := w.CreateSnapshot("with-rel")

		// Clear workspace
		w.Clear()

		// Restore
		w.RestoreSnapshot(snapshot)

		// Verify relationships
		rels := w.GetRelationships()
		if len(rels) != 1 {
			t.Errorf("Expected 1 relationship, got %d", len(rels))
		}
	})

	t.Run("multiple_snapshots", func(t *testing.T) {
		w := New()

		// State 1
		w.AddEntity(createFileEntity("v1.txt"))
		snap1, _ := w.CreateSnapshot("v1")

		// State 2
		w.AddEntity(createFileEntity("v2.txt"))
		snap2, _ := w.CreateSnapshot("v2")

		// State 3
		w.AddEntity(createFileEntity("v3.txt"))

		// Restore to v1
		w.RestoreSnapshot(snap1)
		if len(w.GetEntities()) != 1 {
			t.Errorf("After restore to v1, expected 1 entity, got %d", len(w.GetEntities()))
		}

		// Restore to v2
		w.RestoreSnapshot(snap2)
		if len(w.GetEntities()) != 2 {
			t.Errorf("After restore to v2, expected 2 entities, got %d", len(w.GetEntities()))
		}
	})
}

func TestSnapshotStore(t *testing.T) {
	t.Run("save_and_get", func(t *testing.T) {
		store := NewSnapshotStore()
		w := New()
		w.AddEntity(createFileEntity("test.txt"))
		snapshot, _ := w.CreateSnapshot("test")

		err := store.Save(snapshot)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		retrieved, found := store.Get("test")
		if !found {
			t.Fatal("Snapshot should be found")
		}
		if retrieved.ID != "test" {
			t.Errorf("Retrieved snapshot ID should be 'test', got %s", retrieved.ID)
		}
	})

	t.Run("save_nil_snapshot", func(t *testing.T) {
		store := NewSnapshotStore()
		err := store.Save(nil)
		if err == nil {
			t.Error("Save(nil) should fail")
		}
	})

	t.Run("save_empty_id", func(t *testing.T) {
		store := NewSnapshotStore()
		snapshot := &Snapshot{Timestamp: 123}
		err := store.Save(snapshot)
		if err == nil {
			t.Error("Save with empty ID should fail")
		}
	})

	t.Run("get_nonexistent", func(t *testing.T) {
		store := NewSnapshotStore()
		_, found := store.Get("nonexistent")
		if found {
			t.Error("Non-existent snapshot should not be found")
		}
	})

	t.Run("delete", func(t *testing.T) {
		store := NewSnapshotStore()
		w := New()
		snapshot, _ := w.CreateSnapshot("to-delete")
		store.Save(snapshot)

		deleted := store.Delete("to-delete")
		if !deleted {
			t.Error("Delete should return true for existing snapshot")
		}

		_, found := store.Get("to-delete")
		if found {
			t.Error("Snapshot should not exist after delete")
		}

		// Delete non-existent
		deleted = store.Delete("nonexistent")
		if deleted {
			t.Error("Delete should return false for non-existent snapshot")
		}
	})

	t.Run("list", func(t *testing.T) {
		store := NewSnapshotStore()
		w := New()

		snap1, _ := w.CreateSnapshot("first")
		snap2, _ := w.CreateSnapshot("second")
		store.Save(snap1)
		store.Save(snap2)

		ids := store.List()
		if len(ids) != 2 {
			t.Errorf("Expected 2 snapshots, got %d", len(ids))
		}
	})

	t.Run("count", func(t *testing.T) {
		store := NewSnapshotStore()
		if store.Count() != 0 {
			t.Error("Empty store should have 0 snapshots")
		}

		w := New()
		snap1, _ := w.CreateSnapshot("s1")
		snap2, _ := w.CreateSnapshot("s2")
		store.Save(snap1)
		store.Save(snap2)

		if store.Count() != 2 {
			t.Errorf("Expected 2 snapshots, got %d", store.Count())
		}
	})
}

func TestWorkspace_Config(t *testing.T) {
	t.Run("default_config", func(t *testing.T) {
		w := New()
		cfg := w.GetConfig()

		if cfg.MaxEntities != 0 {
			t.Errorf("Default MaxEntities should be 0, got %d", cfg.MaxEntities)
		}
		if cfg.MaxVersions != 100 {
			t.Errorf("Default MaxVersions should be 100, got %d", cfg.MaxVersions)
		}
		if cfg.AllowDuplicateNames {
			t.Error("Default AllowDuplicateNames should be false")
		}
		if !cfg.StrictValidation {
			t.Error("Default StrictValidation should be true")
		}
	})

	t.Run("with_config", func(t *testing.T) {
		cfg := &Config{
			MaxEntities:      10,
			MaxRelationships: 20,
			MaxVersions:      5,
			EnableVersioning: true,
		}
		w := New().WithConfig(cfg)

		retrieved := w.GetConfig()
		if retrieved.MaxEntities != 10 {
			t.Errorf("MaxEntities should be 10, got %d", retrieved.MaxEntities)
		}
		if retrieved.MaxRelationships != 20 {
			t.Errorf("MaxRelationships should be 20, got %d", retrieved.MaxRelationships)
		}
	})

	t.Run("max_entities_limit", func(t *testing.T) {
		cfg := &Config{MaxEntities: 2}
		w := New().WithConfig(cfg)

		// Add first two entities
		w.AddEntity(createFileEntity("file1.txt"))
		w.AddEntity(createFileEntity("file2.txt"))

		// Third should fail
		err := w.AddEntity(createFileEntity("file3.txt"))
		if err == nil {
			t.Error("AddEntity should fail when max entities reached")
		}
		if len(w.GetEntities()) != 2 {
			t.Errorf("Should have 2 entities, got %d", len(w.GetEntities()))
		}
	})

	t.Run("max_relationships_limit", func(t *testing.T) {
		cfg := &Config{MaxRelationships: 1}
		w := New().WithConfig(cfg)

		w.AddEntity(createFileEntity("file1.txt"))
		w.AddEntity(createFileEntity("file2.txt"))
		w.AddEntity(createAgentEntity("agent1"))

		// First relationship should succeed
		err := w.AddRelationship("agent", "agent1", "file", "file1.txt", RelationTypeAssigned)
		if err != nil {
			t.Fatalf("First relationship should succeed: %v", err)
		}

		// Second should fail
		err = w.AddRelationship("agent", "agent1", "file", "file2.txt", RelationTypeAssigned)
		if err == nil {
			t.Error("AddRelationship should fail when max relationships reached")
		}
	})

	t.Run("allowed_entity_types", func(t *testing.T) {
		cfg := &Config{
			AllowedEntityTypes: []string{"file", "agent"},
		}
		w := New().WithConfig(cfg)

		// Allowed types should succeed
		err := w.AddEntity(createFileEntity("test.txt"))
		if err != nil {
			t.Errorf("file should be allowed: %v", err)
		}

		err = w.AddEntity(createAgentEntity("assistant"))
		if err != nil {
			t.Errorf("agent should be allowed: %v", err)
		}

		// Disallowed type should fail
		err = w.AddEntity(createToolEntity("mytool"))
		if err == nil {
			t.Error("tool should not be allowed")
		}
	})

	t.Run("duplicate_names_not_allowed", func(t *testing.T) {
		cfg := &Config{AllowDuplicateNames: false}
		w := New().WithConfig(cfg)

		w.AddEntity(createFileEntity("test.txt"))

		// Adding another file with the same name should fail
		err := w.AddEntity(createFileEntity("test.txt"))
		if err == nil {
			t.Error("Duplicate name should not be allowed")
		}
	})

	t.Run("duplicate_names_allowed", func(t *testing.T) {
		cfg := &Config{AllowDuplicateNames: true}
		w := New().WithConfig(cfg)

		w.AddEntity(createFileEntity("test.txt"))

		// Adding another file with the same name should succeed
		err := w.AddEntity(createFileEntity("test.txt"))
		if err != nil {
			t.Errorf("Duplicate name should be allowed: %v", err)
		}
		if len(w.GetEntities()) != 2 {
			t.Errorf("Should have 2 entities, got %d", len(w.GetEntities()))
		}
	})

	t.Run("max_versions_limit", func(t *testing.T) {
		cfg := &Config{
			EnableVersioning: true,
			MaxVersions:      3,
		}
		w := New().WithConfig(cfg)

		// Add entity and update it multiple times
		w.AddEntity(createFileEntity("test.txt")) // v1

		for i := 2; i <= 5; i++ {
			updated, _ := ast.NewEntity("file", "test.txt")
			updated.SetProperty("path", ast.StringValue{Value: fmt.Sprintf("/v%d", i)})
			w.UpdateEntity(updated)
		}

		// Should only keep last 3 versions
		count := w.GetEntityVersionCount("file", "test.txt")
		if count != 3 {
			t.Errorf("Should have 3 versions (MaxVersions), got %d", count)
		}

		// Check that version numbers are renumbered correctly
		history := w.GetEntityHistory("file", "test.txt")
		if len(history) != 3 {
			t.Fatalf("Expected 3 versions in history, got %d", len(history))
		}
		for i, v := range history {
			if v.Version != i+1 {
				t.Errorf("Version %d should have Version=%d, got %d", i, i+1, v.Version)
			}
		}
	})

	t.Run("enable_versioning_via_config", func(t *testing.T) {
		cfg := &Config{EnableVersioning: true}
		w := New().WithConfig(cfg)

		w.AddEntity(createFileEntity("test.txt"))

		count := w.GetEntityVersionCount("file", "test.txt")
		if count != 1 {
			t.Errorf("Versioning should be enabled via config, got %d versions", count)
		}
	})

	t.Run("get_config_returns_copy", func(t *testing.T) {
		cfg := &Config{
			MaxEntities:        5,
			AllowedEntityTypes: []string{"file"},
		}
		w := New().WithConfig(cfg)

		retrieved := w.GetConfig()
		retrieved.MaxEntities = 100
		retrieved.AllowedEntityTypes[0] = "agent"

		// Original config should be unchanged
		current := w.GetConfig()
		if current.MaxEntities != 5 {
			t.Error("GetConfig should return a copy, not modify original")
		}
		if current.AllowedEntityTypes[0] != "file" {
			t.Error("GetConfig should return a copy of AllowedEntityTypes")
		}
	})

	t.Run("nil_config", func(t *testing.T) {
		w := New().WithConfig(nil)

		// Should use defaults
		cfg := w.GetConfig()
		if cfg.MaxVersions != 100 {
			t.Error("Nil config should result in defaults")
		}
	})
}

func TestDependencyGraph(t *testing.T) {
	t.Run("add_dependency", func(t *testing.T) {
		dg := NewDependencyGraph()

		err := dg.AddDependency("file", "main.go", "file", "utils.go")
		if err != nil {
			t.Fatalf("AddDependency failed: %v", err)
		}

		deps := dg.GetDependencies("file", "main.go")
		if len(deps) != 1 {
			t.Errorf("Expected 1 dependency, got %d", len(deps))
		}
	})

	t.Run("detect_circular_dependency", func(t *testing.T) {
		dg := NewDependencyGraph()

		// A -> B -> C
		dg.AddDependency("file", "a", "file", "b")
		dg.AddDependency("file", "b", "file", "c")

		// C -> A would create a cycle
		err := dg.AddDependency("file", "c", "file", "a")
		if err == nil {
			t.Error("Should detect circular dependency")
		}
	})

	t.Run("direct_self_reference", func(t *testing.T) {
		dg := NewDependencyGraph()

		err := dg.AddDependency("file", "a", "file", "a")
		if err == nil {
			t.Error("Should detect self-reference")
		}
	})

	t.Run("remove_dependency", func(t *testing.T) {
		dg := NewDependencyGraph()

		dg.AddDependency("file", "a", "file", "b")
		dg.AddDependency("file", "a", "file", "c")

		dg.RemoveDependency("file", "a", "file", "b")

		deps := dg.GetDependencies("file", "a")
		if len(deps) != 1 {
			t.Errorf("Expected 1 dependency after removal, got %d", len(deps))
		}
	})

	t.Run("remove_entity", func(t *testing.T) {
		dg := NewDependencyGraph()

		dg.AddDependency("file", "a", "file", "b")
		dg.AddDependency("file", "b", "file", "c")
		dg.AddDependency("file", "d", "file", "b")

		dg.RemoveEntity("file", "b")

		// a should have no dependencies
		deps := dg.GetDependencies("file", "a")
		if len(deps) != 0 {
			t.Error("a should have no dependencies after b is removed")
		}

		// d should have no dependencies
		deps = dg.GetDependencies("file", "d")
		if len(deps) != 0 {
			t.Error("d should have no dependencies after b is removed")
		}
	})

	t.Run("get_dependents", func(t *testing.T) {
		dg := NewDependencyGraph()

		// a -> c, b -> c
		dg.AddDependency("file", "a", "file", "c")
		dg.AddDependency("file", "b", "file", "c")

		dependents := dg.GetDependents("file", "c")
		if len(dependents) != 2 {
			t.Errorf("Expected 2 dependents, got %d", len(dependents))
		}
	})

	t.Run("get_transitive_dependencies", func(t *testing.T) {
		dg := NewDependencyGraph()

		// a -> b -> c -> d
		dg.AddDependency("file", "a", "file", "b")
		dg.AddDependency("file", "b", "file", "c")
		dg.AddDependency("file", "c", "file", "d")

		trans := dg.GetTransitiveDependencies("file", "a")
		if len(trans) != 3 {
			t.Errorf("Expected 3 transitive dependencies, got %d", len(trans))
		}
	})

	t.Run("topological_sort", func(t *testing.T) {
		dg := NewDependencyGraph()

		// a -> b, a -> c, b -> d, c -> d
		dg.AddDependency("file", "a", "file", "b")
		dg.AddDependency("file", "a", "file", "c")
		dg.AddDependency("file", "b", "file", "d")
		dg.AddDependency("file", "c", "file", "d")

		sorted, err := dg.TopologicalSort()
		if err != nil {
			t.Fatalf("TopologicalSort failed: %v", err)
		}

		if len(sorted) != 4 {
			t.Errorf("Expected 4 entities in sort, got %d", len(sorted))
		}

		// d should come before b and c
		// b and c should come before a
		dIdx, bIdx, cIdx, aIdx := -1, -1, -1, -1
		for i, s := range sorted {
			switch s {
			case "file:d":
				dIdx = i
			case "file:b":
				bIdx = i
			case "file:c":
				cIdx = i
			case "file:a":
				aIdx = i
			}
		}

		if dIdx > bIdx || dIdx > cIdx {
			t.Error("d should come before b and c")
		}
		if bIdx > aIdx || cIdx > aIdx {
			t.Error("b and c should come before a")
		}
	})

	t.Run("topological_sort_with_cycle", func(t *testing.T) {
		dg := NewDependencyGraph()

		// Create a cycle manually by manipulating the map
		dg.AddDependency("file", "a", "file", "b")

		// This shouldn't create a cycle since we check
		// But if somehow there's a cycle, TopologicalSort should detect it
		_, err := dg.TopologicalSort()
		if err != nil {
			t.Errorf("No cycle should exist: %v", err)
		}
	})

	t.Run("clear", func(t *testing.T) {
		dg := NewDependencyGraph()

		dg.AddDependency("file", "a", "file", "b")
		dg.AddDependency("file", "b", "file", "c")

		dg.Clear()

		if dg.Count() != 0 {
			t.Errorf("Expected 0 dependencies after clear, got %d", dg.Count())
		}
	})

	t.Run("count", func(t *testing.T) {
		dg := NewDependencyGraph()

		dg.AddDependency("file", "a", "file", "b")
		dg.AddDependency("file", "a", "file", "c")
		dg.AddDependency("file", "b", "file", "d")

		if dg.Count() != 3 {
			t.Errorf("Expected 3 dependencies, got %d", dg.Count())
		}
	})

	t.Run("duplicate_dependency", func(t *testing.T) {
		dg := NewDependencyGraph()

		dg.AddDependency("file", "a", "file", "b")
		dg.AddDependency("file", "a", "file", "b") // Duplicate

		deps := dg.GetDependencies("file", "a")
		if len(deps) != 1 {
			t.Errorf("Duplicate should not be added, got %d dependencies", len(deps))
		}
	})
}

func TestConcurrentProcessing(t *testing.T) {
	t.Run("add_entities_batch", func(t *testing.T) {
		w := New()

		// Create multiple entities
		var entities []ast.Entity
		for i := 0; i < 10; i++ {
			e, _ := ast.NewEntity("file", fmt.Sprintf("file%d.txt", i))
			entities = append(entities, e)
		}

		results := w.AddEntitiesBatch(entities, 4)

		if len(results) != 10 {
			t.Fatalf("Expected 10 results, got %d", len(results))
		}

		// Check all succeeded
		for i, r := range results {
			if r.Error != nil {
				t.Errorf("Entity %d failed: %v", i, r.Error)
			}
		}

		// Verify entities were added
		if len(w.GetEntities()) != 10 {
			t.Errorf("Expected 10 entities in workspace, got %d", len(w.GetEntities()))
		}
	})

	t.Run("update_entities_batch", func(t *testing.T) {
		w := New()

		// Add entities first
		for i := 0; i < 5; i++ {
			e, _ := ast.NewEntity("file", fmt.Sprintf("file%d.txt", i))
			w.AddEntity(e)
		}

		// Create updated entities
		var updates []ast.Entity
		for i := 0; i < 5; i++ {
			e, _ := ast.NewEntity("file", fmt.Sprintf("file%d.txt", i))
			e.SetProperty("updated", ast.BoolValue{Value: true})
			updates = append(updates, e)
		}

		results := w.UpdateEntitiesBatch(updates, 2)

		// Check all succeeded
		for i, r := range results {
			if r.Error != nil {
				t.Errorf("Update %d failed: %v", i, r.Error)
			}
		}
	})

	t.Run("upsert_entities_batch", func(t *testing.T) {
		w := New()

		// Add some entities first
		for i := 0; i < 3; i++ {
			e, _ := ast.NewEntity("file", fmt.Sprintf("file%d.txt", i))
			w.AddEntity(e)
		}

		// Upsert mix of existing and new
		var entities []ast.Entity
		for i := 0; i < 6; i++ {
			e, _ := ast.NewEntity("file", fmt.Sprintf("file%d.txt", i))
			entities = append(entities, e)
		}

		results := w.UpsertEntitiesBatch(entities, 3)

		for i, r := range results {
			if r.Error != nil {
				t.Errorf("Upsert %d failed: %v", i, r.Error)
			}
		}

		// Should have 6 entities (3 updated + 3 new)
		if len(w.GetEntities()) != 6 {
			t.Errorf("Expected 6 entities, got %d", len(w.GetEntities()))
		}
	})

	t.Run("process_entities_concurrently", func(t *testing.T) {
		w := New()

		var entities []ast.Entity
		for i := 0; i < 5; i++ {
			e, _ := ast.NewEntity("file", fmt.Sprintf("file%d.txt", i))
			entities = append(entities, e)
		}

		var processed int32
		results := w.ProcessEntitiesConcurrently(entities, func(e ast.Entity) error {
			// Simulate some work
			atomic.AddInt32(&processed, 1)
			return nil
		}, 2)

		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}

		if atomic.LoadInt32(&processed) != 5 {
			t.Errorf("Expected 5 processed, got %d", processed)
		}
	})

	t.Run("process_with_errors", func(t *testing.T) {
		w := New()

		var entities []ast.Entity
		for i := 0; i < 4; i++ {
			e, _ := ast.NewEntity("file", fmt.Sprintf("file%d.txt", i))
			entities = append(entities, e)
		}

		results := w.ProcessEntitiesConcurrently(entities, func(e ast.Entity) error {
			if e.Name() == "file2.txt" {
				return fmt.Errorf("error processing %s", e.Name())
			}
			return nil
		}, 0) // Unlimited concurrency

		var errorCount int
		for _, r := range results {
			if r.Error != nil {
				errorCount++
			}
		}

		if errorCount != 1 {
			t.Errorf("Expected 1 error, got %d", errorCount)
		}
	})

	t.Run("transform_entities", func(t *testing.T) {
		w := New()

		// Add file entities
		for i := 0; i < 5; i++ {
			e, _ := ast.NewEntity("file", fmt.Sprintf("file%d.txt", i))
			w.AddEntity(e)
		}

		// Add agent entity (should not be transformed)
		agent, _ := ast.NewEntity("agent", "test-agent")
		w.AddEntity(agent)

		// Transform only file entities
		transformed, errors := w.TransformEntities(
			func(e ast.Entity) bool {
				return e.Type() == "file"
			},
			func(e ast.Entity) (ast.Entity, error) {
				newEntity, _ := ast.NewEntity(e.Type(), e.Name())
				newEntity.SetProperty("transformed", ast.BoolValue{Value: true})
				return newEntity, nil
			},
			2,
		)

		if len(errors) != 0 {
			t.Errorf("Expected no errors, got %d", len(errors))
		}

		if len(transformed) != 5 {
			t.Errorf("Expected 5 transformed entities, got %d", len(transformed))
		}
	})

	t.Run("filter_entities_concurrently", func(t *testing.T) {
		w := New()

		// Add mix of entities
		for i := 0; i < 10; i++ {
			e, _ := ast.NewEntity("file", fmt.Sprintf("file%d.txt", i))
			w.AddEntity(e)
		}
		for i := 0; i < 5; i++ {
			e, _ := ast.NewEntity("agent", fmt.Sprintf("agent%d", i))
			w.AddEntity(e)
		}

		// Filter only even-numbered files
		filtered := w.FilterEntitiesConcurrently(func(e ast.Entity) bool {
			if e.Type() != "file" {
				return false
			}
			name := e.Name()
			// Simple check for even numbers
			return len(name) > 0 && (name[4] == '0' || name[4] == '2' || name[4] == '4' || name[4] == '6' || name[4] == '8')
		}, 4)

		if len(filtered) != 5 {
			t.Errorf("Expected 5 filtered entities (even numbered), got %d", len(filtered))
		}
	})

	t.Run("foreach_entity", func(t *testing.T) {
		w := New()

		for i := 0; i < 5; i++ {
			e, _ := ast.NewEntity("file", fmt.Sprintf("file%d.txt", i))
			w.AddEntity(e)
		}

		var count int32
		results := w.ForEachEntity(func(e ast.Entity) error {
			atomic.AddInt32(&count, 1)
			return nil
		}, 2)

		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}

		if atomic.LoadInt32(&count) != 5 {
			t.Errorf("Expected count 5, got %d", count)
		}
	})

	t.Run("foreach_entity_of_type", func(t *testing.T) {
		w := New()

		// Add mix
		for i := 0; i < 5; i++ {
			e, _ := ast.NewEntity("file", fmt.Sprintf("file%d.txt", i))
			w.AddEntity(e)
		}
		for i := 0; i < 3; i++ {
			e, _ := ast.NewEntity("agent", fmt.Sprintf("agent%d", i))
			w.AddEntity(e)
		}

		var count int32
		results := w.ForEachEntityOfType("agent", func(e ast.Entity) error {
			atomic.AddInt32(&count, 1)
			return nil
		}, 2)

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		if atomic.LoadInt32(&count) != 3 {
			t.Errorf("Expected count 3, got %d", count)
		}
	})

	t.Run("empty_entities", func(t *testing.T) {
		w := New()

		results := w.AddEntitiesBatch(nil, 4)
		if results != nil {
			t.Errorf("Expected nil results for nil input")
		}

		results = w.AddEntitiesBatch([]ast.Entity{}, 4)
		if results != nil {
			t.Errorf("Expected nil results for empty input")
		}
	})
}

func TestCustomValidators(t *testing.T) {
	t.Run("type_specific_validator", func(t *testing.T) {
		w := New()

		// Register a validator that only accepts file entities with .go extension
		w.RegisterEntityValidator("file", func(e ast.Entity) error {
			name := e.Name()
			if len(name) < 3 || name[len(name)-3:] != ".go" {
				return fmt.Errorf("file must have .go extension")
			}
			return nil
		})

		// Valid file
		validFile, _ := ast.NewEntity("file", "main.go")
		err := w.AddEntity(validFile)
		if err != nil {
			t.Errorf("Valid file should be accepted: %v", err)
		}

		// Invalid file
		invalidFile, _ := ast.NewEntity("file", "config.txt")
		err = w.AddEntity(invalidFile)
		if err == nil {
			t.Error("Invalid file should be rejected")
		}

		// Agent (different type, should not be validated)
		agent, _ := ast.NewEntity("agent", "test-agent")
		err = w.AddEntity(agent)
		if err != nil {
			t.Errorf("Agent should not be affected by file validator: %v", err)
		}
	})

	t.Run("global_validator", func(t *testing.T) {
		w := New()

		// Register a global validator
		w.RegisterGlobalValidator(func(e ast.Entity) error {
			if e.Name() == "" {
				return fmt.Errorf("entity name cannot be empty")
			}
			return nil
		})

		// Valid entity
		validFile, _ := ast.NewEntity("file", "test.go")
		err := w.AddEntity(validFile)
		if err != nil {
			t.Errorf("Valid entity should be accepted: %v", err)
		}

		// All entity types should be validated
		agent, _ := ast.NewEntity("agent", "test-agent")
		err = w.AddEntity(agent)
		if err != nil {
			t.Errorf("Valid agent should be accepted: %v", err)
		}
	})

	t.Run("multiple_validators_same_type", func(t *testing.T) {
		w := New()

		// First validator: name must start with lowercase
		w.RegisterEntityValidator("file", func(e ast.Entity) error {
			name := e.Name()
			if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
				return fmt.Errorf("file name must start with lowercase")
			}
			return nil
		})

		// Second validator: name must not be empty
		w.RegisterEntityValidator("file", func(e ast.Entity) error {
			if e.Name() == "" {
				return fmt.Errorf("file name cannot be empty")
			}
			return nil
		})

		// Valid file
		validFile, _ := ast.NewEntity("file", "main.go")
		err := w.AddEntity(validFile)
		if err != nil {
			t.Errorf("Valid file should pass both validators: %v", err)
		}

		// File that fails first validator
		invalidFile, _ := ast.NewEntity("file", "Main.go")
		err = w.AddEntity(invalidFile)
		if err == nil {
			t.Error("File with uppercase should be rejected")
		}
	})

	t.Run("validator_on_update", func(t *testing.T) {
		w := New()

		// Add entity first (no validators)
		entity, _ := ast.NewEntity("file", "test.go")
		w.AddEntity(entity)

		// Now add a validator
		w.RegisterEntityValidator("file", func(e ast.Entity) error {
			if val, ok := e.GetProperty("locked"); ok {
				if boolVal, ok := val.(ast.BoolValue); ok && boolVal.Value {
					return fmt.Errorf("cannot update locked file")
				}
			}
			return nil
		})

		// Update with locked property
		updated, _ := ast.NewEntity("file", "test.go")
		updated.SetProperty("locked", ast.BoolValue{Value: true})
		err := w.UpdateEntity(updated)
		if err == nil {
			t.Error("Should reject update of locked file")
		}
	})

	t.Run("validator_on_upsert", func(t *testing.T) {
		w := New()

		w.RegisterEntityValidator("file", func(e ast.Entity) error {
			if strings.HasPrefix(e.Name(), ".") {
				return fmt.Errorf("hidden files not allowed")
			}
			return nil
		})

		// Upsert valid file
		validFile, _ := ast.NewEntity("file", "test.go")
		err := w.UpsertEntity(validFile)
		if err != nil {
			t.Errorf("Valid file upsert should succeed: %v", err)
		}

		// Upsert hidden file (invalid)
		hiddenFile, _ := ast.NewEntity("file", ".gitignore")
		err = w.UpsertEntity(hiddenFile)
		if err == nil {
			t.Error("Hidden file upsert should fail")
		}
	})

	t.Run("clear_validators", func(t *testing.T) {
		w := New()

		w.RegisterEntityValidator("file", func(e ast.Entity) error {
			return fmt.Errorf("always reject")
		})

		// Entity should be rejected
		entity, _ := ast.NewEntity("file", "test.go")
		err := w.AddEntity(entity)
		if err == nil {
			t.Error("Should reject before clearing")
		}

		// Clear validators
		w.ClearValidators()

		// Now entity should be accepted
		err = w.AddEntity(entity)
		if err != nil {
			t.Errorf("Should accept after clearing: %v", err)
		}
	})

	t.Run("clear_validators_for_type", func(t *testing.T) {
		w := New()

		w.RegisterEntityValidator("file", func(e ast.Entity) error {
			return fmt.Errorf("reject files")
		})
		w.RegisterEntityValidator("agent", func(e ast.Entity) error {
			return fmt.Errorf("reject agents")
		})

		// Clear only file validators
		w.ClearValidatorsForType("file")

		// File should be accepted
		file, _ := ast.NewEntity("file", "test.go")
		err := w.AddEntity(file)
		if err != nil {
			t.Errorf("File should be accepted after clearing file validators: %v", err)
		}

		// Agent should still be rejected
		agent, _ := ast.NewEntity("agent", "test-agent")
		err = w.AddEntity(agent)
		if err == nil {
			t.Error("Agent should still be rejected")
		}
	})

	t.Run("global_and_type_validators_combined", func(t *testing.T) {
		w := New()

		// Global validator
		w.RegisterGlobalValidator(func(e ast.Entity) error {
			if len(e.Name()) < 2 {
				return fmt.Errorf("name too short")
			}
			return nil
		})

		// Type-specific validator
		w.RegisterEntityValidator("file", func(e ast.Entity) error {
			if !strings.HasSuffix(e.Name(), ".go") {
				return fmt.Errorf("must be .go file")
			}
			return nil
		})

		// Should fail global validator
		shortName, _ := ast.NewEntity("file", "a")
		err := w.AddEntity(shortName)
		if err == nil || !strings.Contains(err.Error(), "name too short") {
			t.Error("Should fail global validator")
		}

		// Should fail type validator
		wrongExt, _ := ast.NewEntity("file", "test.txt")
		err = w.AddEntity(wrongExt)
		if err == nil || !strings.Contains(err.Error(), "must be .go file") {
			t.Error("Should fail type validator")
		}

		// Should pass both
		valid, _ := ast.NewEntity("file", "main.go")
		err = w.AddEntity(valid)
		if err != nil {
			t.Errorf("Should pass both validators: %v", err)
		}
	})
}

func TestPipeline(t *testing.T) {
	t.Run("simple_pipeline", func(t *testing.T) {
		pipeline := NewPipeline("test-pipeline")
		pipeline.AddStage("add-property", func(e ast.Entity) (ast.Entity, error) {
			e.SetProperty("processed", ast.BoolValue{Value: true})
			return e, nil
		})

		entity, _ := ast.NewEntity("file", "test.go")
		result := pipeline.Execute(entity)

		if result.Error != nil {
			t.Fatalf("Pipeline failed: %v", result.Error)
		}

		if len(result.StagesExecuted) != 1 {
			t.Errorf("Expected 1 stage executed, got %d", len(result.StagesExecuted))
		}

		val, ok := result.ResultEntity.GetProperty("processed")
		if !ok {
			t.Error("Property 'processed' not set")
		}
		if bv, ok := val.(ast.BoolValue); !ok || !bv.Value {
			t.Error("Property 'processed' should be true")
		}
	})

	t.Run("multi_stage_pipeline", func(t *testing.T) {
		pipeline := NewPipeline("multi-stage")
		pipeline.AddStage("stage1", func(e ast.Entity) (ast.Entity, error) {
			e.SetProperty("stage1", ast.BoolValue{Value: true})
			return e, nil
		})
		pipeline.AddStage("stage2", func(e ast.Entity) (ast.Entity, error) {
			e.SetProperty("stage2", ast.BoolValue{Value: true})
			return e, nil
		})
		pipeline.AddStage("stage3", func(e ast.Entity) (ast.Entity, error) {
			e.SetProperty("stage3", ast.BoolValue{Value: true})
			return e, nil
		})

		entity, _ := ast.NewEntity("file", "test.go")
		result := pipeline.Execute(entity)

		if len(result.StagesExecuted) != 3 {
			t.Errorf("Expected 3 stages executed, got %d", len(result.StagesExecuted))
		}

		for i := 1; i <= 3; i++ {
			prop := fmt.Sprintf("stage%d", i)
			if _, ok := result.ResultEntity.GetProperty(prop); !ok {
				t.Errorf("Property '%s' not set", prop)
			}
		}
	})

	t.Run("conditional_stage", func(t *testing.T) {
		pipeline := NewPipeline("conditional")
		pipeline.AddConditionalStage("files-only",
			func(e ast.Entity) bool { return e.Type() == "file" },
			func(e ast.Entity) (ast.Entity, error) {
				e.SetProperty("is-file", ast.BoolValue{Value: true})
				return e, nil
			})

		// File should be transformed
		file, _ := ast.NewEntity("file", "test.go")
		fileResult := pipeline.Execute(file)
		if len(fileResult.StagesExecuted) != 1 {
			t.Error("File should have 1 stage executed")
		}

		// Agent should be skipped
		agent, _ := ast.NewEntity("agent", "test-agent")
		agentResult := pipeline.Execute(agent)
		if len(agentResult.StagesSkipped) != 1 {
			t.Error("Agent should have 1 stage skipped")
		}
		if len(agentResult.StagesExecuted) != 0 {
			t.Error("Agent should have 0 stages executed")
		}
	})

	t.Run("stage_error", func(t *testing.T) {
		pipeline := NewPipeline("error-pipeline")
		pipeline.AddStage("success-stage", func(e ast.Entity) (ast.Entity, error) {
			e.SetProperty("stage1", ast.BoolValue{Value: true})
			return e, nil
		})
		pipeline.AddStage("error-stage", func(e ast.Entity) (ast.Entity, error) {
			return nil, fmt.Errorf("intentional error")
		})
		pipeline.AddStage("never-reached", func(e ast.Entity) (ast.Entity, error) {
			e.SetProperty("stage3", ast.BoolValue{Value: true})
			return e, nil
		})

		entity, _ := ast.NewEntity("file", "test.go")
		result := pipeline.Execute(entity)

		if result.Error == nil {
			t.Error("Should have an error")
		}
		if result.FailedStageName != "error-stage" {
			t.Errorf("Expected failed stage 'error-stage', got %s", result.FailedStageName)
		}
		if len(result.StagesExecuted) != 1 {
			t.Errorf("Should have 1 stage executed before error, got %d", len(result.StagesExecuted))
		}
	})

	t.Run("execute_all", func(t *testing.T) {
		pipeline := NewPipeline("batch")
		pipeline.AddStage("add-counter", func(e ast.Entity) (ast.Entity, error) {
			e.SetProperty("processed", ast.BoolValue{Value: true})
			return e, nil
		})

		var entities []ast.Entity
		for i := 0; i < 5; i++ {
			e, _ := ast.NewEntity("file", fmt.Sprintf("file%d.go", i))
			entities = append(entities, e)
		}

		results := pipeline.ExecuteAll(entities)

		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}

		for _, r := range results {
			if r.Error != nil {
				t.Errorf("Unexpected error: %v", r.Error)
			}
		}
	})

	t.Run("workspace_execute_pipeline", func(t *testing.T) {
		w := New()

		// Add entities
		for i := 0; i < 5; i++ {
			e, _ := ast.NewEntity("file", fmt.Sprintf("file%d.go", i))
			w.AddEntity(e)
		}
		for i := 0; i < 3; i++ {
			e, _ := ast.NewEntity("agent", fmt.Sprintf("agent%d", i))
			w.AddEntity(e)
		}

		pipeline := NewPipeline("files-only")
		pipeline.AddStage("mark", func(e ast.Entity) (ast.Entity, error) {
			e.SetProperty("marked", ast.BoolValue{Value: true})
			return e, nil
		})

		// Execute only on files
		results := w.ExecutePipeline(pipeline, func(e ast.Entity) bool {
			return e.Type() == "file"
		})

		if len(results) != 5 {
			t.Errorf("Expected 5 results (files only), got %d", len(results))
		}
	})

	t.Run("workspace_execute_and_update", func(t *testing.T) {
		w := New()

		// Add entities
		for i := 0; i < 3; i++ {
			e, _ := ast.NewEntity("file", fmt.Sprintf("file%d.go", i))
			w.AddEntity(e)
		}

		pipeline := NewPipeline("updater")
		pipeline.AddStage("add-updated-flag", func(e ast.Entity) (ast.Entity, error) {
			// Create new entity to trigger update
			newE, _ := ast.NewEntity(e.Type(), e.Name())
			newE.SetProperty("updated", ast.BoolValue{Value: true})
			return newE, nil
		})

		results, err := w.ExecutePipelineAndUpdate(pipeline, nil)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		// Verify entities were updated
		for _, e := range w.GetEntitiesByType("file") {
			val, ok := e.GetProperty("updated")
			if !ok {
				t.Error("Entity should have 'updated' property")
			}
			if bv, ok := val.(ast.BoolValue); !ok || !bv.Value {
				t.Error("'updated' should be true")
			}
		}
	})

	t.Run("empty_pipeline", func(t *testing.T) {
		pipeline := NewPipeline("empty")

		entity, _ := ast.NewEntity("file", "test.go")
		result := pipeline.Execute(entity)

		if result.Error != nil {
			t.Errorf("Empty pipeline should not error: %v", result.Error)
		}
		if result.ResultEntity != result.OriginalEntity {
			t.Error("Empty pipeline should return original entity")
		}
	})
}
