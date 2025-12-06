# Workspace Package

The `workspace` package manages LangSpace workspaces, providing entity storage, querying, relationship management, event hooks, and lifecycle management. It serves as the runtime environment for LangSpace entities.

## Overview

A workspace is responsible for:
- Entity storage and retrieval
- Entity relationship management
- Event hooks for validation and notification
- Operation execution
- State management

## Usage

```go
import (
    "github.com/shellkjell/langspace/pkg/workspace"
    "github.com/shellkjell/langspace/pkg/validator"
)

// Create a new workspace with validator
ws := workspace.New().WithValidator(validator.New())

// Add entities
err := ws.AddEntity(entity)
if err != nil {
    log.Fatal(err)
}

// Update an existing entity
updatedEntity, _ := ast.NewEntity("file", "test.txt")
updatedEntity.SetProperty("path", ast.StringValue{Value: "/new/path"})
err = ws.UpdateEntity(updatedEntity)

// Upsert - add if doesn't exist, update if it does
err = ws.UpsertEntity(entity)

// Remove entities
err = ws.RemoveEntity("file", "test.txt")

// Query entities by type
files := ws.GetEntitiesByType("file")
agents := ws.GetEntitiesByType("agent")
tools := ws.GetEntitiesByType("tool")
scripts := ws.GetEntitiesByType("script")

// Get workspace statistics
stats := ws.Stat()
fmt.Printf("Total: %d, Files: %d, Agents: %d, Tools: %d, Relationships: %d, Hooks: %d\n",
    stats.TotalEntities, stats.FileEntities, stats.AgentEntities,
    stats.ToolEntities, stats.TotalRelationships, stats.TotalHooks)
```

## Entity Hooks

The workspace supports event hooks that allow you to intercept entity lifecycle events:

```go
// Hook types:
// - HookBeforeAdd: Called before an entity is added (can block addition)
// - HookAfterAdd: Called after an entity is added (notification only)
// - HookBeforeRemove: Called before an entity is removed (can block removal)
// - HookAfterRemove: Called after an entity is removed (notification only)
// - HookBeforeUpdate: Called before an entity is updated (can block update)
// - HookAfterUpdate: Called after an entity is updated (notification only)

// Register a validation hook (blocks invalid file names)
ws.OnEntityEvent(workspace.HookBeforeAdd, func(entity ast.Entity) error {
    if entity.Type() == "file" {
        name := entity.Properties()[0]
        if strings.HasPrefix(name, ".") {
            return fmt.Errorf("hidden files not allowed: %s", name)
        }
    }
    return nil
})

// Register a notification hook (logs additions)
ws.OnEntityEvent(workspace.HookAfterAdd, func(entity ast.Entity) error {
    log.Printf("Entity added: %s %s", entity.Type(), entity.Properties()[0])
    return nil // after-hooks errors are logged but don't affect the operation
})

// Register removal hooks
ws.OnEntityEvent(workspace.HookBeforeRemove, func(entity ast.Entity) error {
    if entity.Type() == "file" && entity.Properties()[0] == "critical.txt" {
        return fmt.Errorf("cannot remove critical file")
    }
    return nil
})
```

## Entity Relationships

The workspace supports defining relationships between entities:

```go
// Define relationship types
// - RelationTypeAssigned: An agent is assigned to work on a file or task
// - RelationTypeDepends: A task depends on another entity
// - RelationTypeProduces: An entity produces another entity
// - RelationTypeConsumes: An entity consumes/reads another entity

// Add a relationship (agent "validator" is assigned to file "config.json")
err := ws.AddRelationship("agent", "validator", "file", "config.json", workspace.RelationTypeAssigned)

// Get all relationships
relationships := ws.GetRelationships()

// Get relationships for a specific entity
fileRels := ws.GetRelationshipsForEntity("file", "config.json")

// Get related entities
relatedFiles := ws.GetRelatedEntities("agent", "validator", workspace.RelationTypeAssigned)

// Remove a relationship
err := ws.RemoveRelationship("agent", "validator", "file", "config.json", workspace.RelationTypeAssigned)
```

## Event System

The workspace emits events for all state changes. Unlike hooks, events cannot block operations and are called after operations complete successfully:

```go
// Event types:
// - EventEntityAdded: Emitted when an entity is added
// - EventEntityRemoved: Emitted when an entity is removed
// - EventEntityUpdated: Emitted when an entity is updated
// - EventRelationshipAdded: Emitted when a relationship is added
// - EventRelationshipRemoved: Emitted when a relationship is removed
// - EventWorkspaceCleared: Emitted when the workspace is cleared

// Register an event handler
ws.OnEvent(func(event workspace.Event) {
    switch event.Type {
    case workspace.EventEntityAdded:
        log.Printf("Entity added: %s %s", event.Entity.Type(), event.Entity.Name())
    case workspace.EventEntityRemoved:
        log.Printf("Entity removed: %s %s", event.Entity.Type(), event.Entity.Name())
    case workspace.EventEntityUpdated:
        log.Printf("Entity updated: %s %s", event.Entity.Type(), event.Entity.Name())
    case workspace.EventRelationshipAdded:
        log.Printf("Relationship added: %s -> %s",
            event.Relationship.SourceName, event.Relationship.TargetName)
    case workspace.EventRelationshipRemoved:
        log.Printf("Relationship removed: %s -> %s",
            event.Relationship.SourceName, event.Relationship.TargetName)
    case workspace.EventWorkspaceCleared:
        log.Println("Workspace cleared")
    }
})

// Multiple handlers can be registered
ws.OnEvent(func(event workspace.Event) {
    // Send metrics, update UI, etc.
    metrics.RecordWorkspaceEvent(event.Type)
})
```

## Entity Versioning

The workspace supports tracking entity history when versioning is enabled:

```go
// Enable versioning
ws := workspace.New().WithVersioning()

// Add and update an entity
entity, _ := ast.NewEntity("file", "config.json")
entity.SetProperty("path", ast.StringValue{Value: "/config.json"})
ws.AddEntity(entity)  // Creates version 1

updated, _ := ast.NewEntity("file", "config.json")
updated.SetProperty("path", ast.StringValue{Value: "/new/config.json"})
ws.UpdateEntity(updated)  // Creates version 2

// Get the number of versions
count := ws.GetEntityVersionCount("file", "config.json")  // Returns 2

// Get a specific version
v1, found := ws.GetEntityVersion("file", "config.json", 1)
if found {
    path, _ := v1.GetProperty("path")  // "/config.json"
}

// Get full version history
history := ws.GetEntityHistory("file", "config.json")
for _, version := range history {
    fmt.Printf("Version %d at %d: %s\n",
        version.Version, version.Timestamp, version.Entity.Name())
}
```

## Workspace Persistence

The workspace can be saved to and loaded from files or streams:

```go
// Save workspace to a file
err := ws.SaveToFile("workspace.json")
if err != nil {
    log.Fatal(err)
}

// Load workspace from a file
ws2 := workspace.New()
err = ws2.LoadFromFile("workspace.json")

// Save to any io.Writer
var buf bytes.Buffer
ws.SaveTo(&buf)

// Load from any io.Reader
ws.LoadFrom(strings.NewReader(jsonData))
```

Note: When loading, the workspace's entities and relationships are replaced with the loaded data. Hooks and event handlers are preserved and must be re-registered if needed.

## Workspace Snapshots

Snapshots capture a point-in-time state of the workspace that can be restored later:

```go
// Create a snapshot before making changes
snapshot, err := ws.CreateSnapshot("before-refactor")
if err != nil {
    log.Fatal(err)
}

// Make changes to the workspace
ws.AddEntity(newEntity)
ws.RemoveEntity("file", "old.txt")

// If something goes wrong, restore the snapshot
err = ws.RestoreSnapshot(snapshot)

// Use SnapshotStore to manage multiple snapshots
store := workspace.NewSnapshotStore()
store.Save(snapshot)

// Later, retrieve and restore
snap, found := store.Get("before-refactor")
if found {
    ws.RestoreSnapshot(snap)
}

// List all snapshots
ids := store.List()
fmt.Printf("Available snapshots: %v\n", ids)

// Delete old snapshots
store.Delete("before-refactor")
```

## Features

### Entity Management
- Add/Remove/Update entities with lifecycle hooks
- Query by type or property
- Relationship tracking
- Validation on addition/update
- Upsert support (add or update)
- Entity versioning with full history tracking

### Event Hooks
- `before_add`: Validate before entity addition (can reject)
- `after_add`: Notify after entity addition
- `before_remove`: Validate before entity removal (can reject)
- `after_remove`: Notify after entity removal
- `before_update`: Validate before entity update (can reject)
- `after_update`: Notify after entity update

### Workspace Events
Unlike hooks, events are notifications that cannot block operations:
- `EventEntityAdded`: Emitted after entity is added
- `EventEntityRemoved`: Emitted after entity is removed
- `EventEntityUpdated`: Emitted after entity is updated
- `EventRelationshipAdded`: Emitted after relationship is added
- `EventRelationshipRemoved`: Emitted after relationship is removed
- `EventWorkspaceCleared`: Emitted after workspace is cleared

### Relationship Types
- `assigned`: Agent-to-file/task assignment
- `depends`: Task dependency relationships
- `produces`: Entity production relationships
- `consumes`: Entity consumption relationships

### Operations
- File system operations
- Agent task execution
- Entity modifications
- State persistence

## Best Practices

- Initialize workspace at startup
- Register hooks before adding entities
- Use before-hooks for validation
- Use after-hooks for logging/notification
- Handle workspace errors appropriately
- Clean up resources when done
- Use appropriate query methods
- Define relationships after entities are added

## State Management

The workspace maintains:
- Active entities
- Entity relationships
- Registered hooks
- Operation status
- Resource usage

## Future Enhancements

Planned improvements include:
- Multi-workspace support
- Enhanced query capabilities
- Resource quotas
- Operation scheduling
- Relationship validation rules
