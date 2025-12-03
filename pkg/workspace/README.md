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

// Remove entities
err = ws.RemoveEntity("file", "test.txt")

// Query entities by type
files := ws.GetEntitiesByType("file")
agents := ws.GetEntitiesByType("agent")
tasks := ws.GetEntitiesByType("task")

// Get workspace statistics
stats := ws.Stat()
fmt.Printf("Total: %d, Files: %d, Agents: %d, Tasks: %d, Relationships: %d, Hooks: %d\n",
    stats.TotalEntities, stats.FileEntities, stats.AgentEntities, 
    stats.TaskEntities, stats.TotalRelationships, stats.TotalHooks)
```

## Entity Hooks

The workspace supports event hooks that allow you to intercept entity lifecycle events:

```go
// Hook types:
// - HookBeforeAdd: Called before an entity is added (can block addition)
// - HookAfterAdd: Called after an entity is added (notification only)
// - HookBeforeRemove: Called before an entity is removed (can block removal)
// - HookAfterRemove: Called after an entity is removed (notification only)

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

## Features

### Entity Management
- Add/Remove entities with lifecycle hooks
- Query by type or property
- Relationship tracking
- Validation on addition

### Event Hooks
- `before_add`: Validate before entity addition (can reject)
- `after_add`: Notify after entity addition
- `before_remove`: Validate before entity removal (can reject)
- `after_remove`: Notify after entity removal

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

### Event System
- Entity lifecycle events
- Operation completion events
- Error events
- State change notifications

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
- Workspace persistence
- Multi-workspace support
- Enhanced query capabilities
- Resource quotas
- Operation scheduling
- Relationship validation rules
