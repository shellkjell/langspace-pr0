package workspace

import (
	"sync"
	"testing"

	"github.com/shellkjell/langspace/pkg/ast"
)

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
	fileEntity, _ := ast.NewEntity("file")
	fileEntity.AddProperty("test.txt")
	fileEntity.AddProperty("content")

	if err := w.AddEntity(fileEntity); err != nil {
		t.Errorf("Workspace.AddEntity() error = %v for valid file entity", err)
	}

	agentEntity, _ := ast.NewEntity("agent")
	agentEntity.AddProperty("validator")
	agentEntity.AddProperty("check(test.txt)")

	if err := w.AddEntity(agentEntity); err != nil {
		t.Errorf("Workspace.AddEntity() error = %v for valid agent entity", err)
	}
}

func TestWorkspace_GetEntities(t *testing.T) {
	w := New()
	fileEntity, _ := ast.NewEntity("file")
	fileEntity.AddProperty("test.txt")
	fileEntity.AddProperty("content")
	w.AddEntity(fileEntity)

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

	// Add a file entity
	fileEntity, _ := ast.NewEntity("file")
	fileEntity.AddProperty("test.txt")
	fileEntity.AddProperty("content")
	w.AddEntity(fileEntity)

	// Add an agent entity
	agentEntity, _ := ast.NewEntity("agent")
	agentEntity.AddProperty("validator")
	agentEntity.AddProperty("check(test.txt)")
	w.AddEntity(agentEntity)

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

	// Test getting non-existent type
	unknownEntities := w.GetEntitiesByType("unknown")
	if len(unknownEntities) != 0 {
		t.Errorf("Workspace.GetEntitiesByType(unknown) returned %d entities, want 0", len(unknownEntities))
	}
}

func TestWorkspace_Clear(t *testing.T) {
	w := New()

	// Add some entities
	fileEntity, _ := ast.NewEntity("file")
	fileEntity.AddProperty("test.txt")
	fileEntity.AddProperty("content")
	w.AddEntity(fileEntity)

	agentEntity, _ := ast.NewEntity("agent")
	agentEntity.AddProperty("validator")
	agentEntity.AddProperty("check(test.txt)")
	w.AddEntity(agentEntity)

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
		go func() {
			defer wg.Done()
			fileEntity, _ := ast.NewEntity("file")
			fileEntity.AddProperty("test.txt")
			fileEntity.AddProperty("content")
			w.AddEntity(fileEntity)
		}()
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
	fileEntity, _ := ast.NewEntity("file")
	fileEntity.AddProperty("test.txt")
	fileEntity.AddProperty("content")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.AddEntity(fileEntity)
	}
}

func BenchmarkWorkspace_GetEntities(b *testing.B) {
	w := New()
	fileEntity, _ := ast.NewEntity("file")
	fileEntity.AddProperty("test.txt")
	fileEntity.AddProperty("content")
	w.AddEntity(fileEntity)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.GetEntities()
	}
}
