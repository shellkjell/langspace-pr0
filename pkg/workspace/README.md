# Workspace Package

The `workspace` package manages LangSpace workspaces, providing entity storage, querying, relationship management, event hooks, and lifecycle management. It serves as the runtime environment for LangSpace entities.

## Overview

A workspace is responsible for:
- Entity storage and retrieval
- Entity relationship management
- Event hooks for validation and notification
- Configuration-based limits and constraints
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

## Workspace Configuration

Configure workspace behavior with limits and constraints:

```go
// Create a custom configuration
cfg := &workspace.Config{
    MaxEntities:         100,              // Limit total entities
    MaxRelationships:    500,              // Limit total relationships
    MaxVersions:         10,               // Keep last 10 versions per entity
    AllowDuplicateNames: false,            // Prevent duplicate entity names
    StrictValidation:    true,             // Require validation
    EnableVersioning:    true,             // Enable version tracking
    AllowedEntityTypes:  []string{"file", "agent", "tool"}, // Restrict entity types
}

ws := workspace.New().WithConfig(cfg)

// Get current configuration
currentCfg := ws.GetConfig()

// Use defaults
ws := workspace.New() // Uses DefaultConfig()
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `MaxEntities` | 0 (unlimited) | Maximum number of entities |
| `MaxRelationships` | 0 (unlimited) | Maximum number of relationships |
| `MaxVersions` | 100 | Maximum versions kept per entity |
| `AllowDuplicateNames` | false | Allow entities with same type and name |
| `StrictValidation` | true | Require all entities to pass validation |
| `EnableVersioning` | false | Enable entity version tracking |
| `AllowedEntityTypes` | nil (all) | Restrict which entity types can be added |

## Custom Entity Validators

Register custom validation functions for specific entity types or globally:

```go
// Register a type-specific validator for file entities
ws.RegisterEntityValidator("file", func(e ast.Entity) error {
    name := e.Name()
    if !strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, ".ls") {
        return fmt.Errorf("only .go and .ls files are allowed")
    }
    return nil
})

// Register another validator for the same type (both will run)
ws.RegisterEntityValidator("file", func(e ast.Entity) error {
    if strings.HasPrefix(e.Name(), ".") {
        return fmt.Errorf("hidden files not allowed")
    }
    return nil
})

// Register a global validator that applies to all entity types
ws.RegisterGlobalValidator(func(e ast.Entity) error {
    if len(e.Name()) < 2 {
        return fmt.Errorf("entity name must be at least 2 characters")
    }
    return nil
})

// Validators run during AddEntity, UpdateEntity, and UpsertEntity
file, _ := ast.NewEntity("file", "test.txt")
err := ws.AddEntity(file)
// err: "custom validation failed: only .go and .ls files are allowed"

// Clear all validators
ws.ClearValidators()

// Clear validators for a specific type
ws.ClearValidatorsForType("file")
```

Validators are executed in the following order:
1. Global validators (registered with `RegisterGlobalValidator`)
2. Type-specific validators (registered with `RegisterEntityValidator`)

Both global and type-specific validators must pass for an entity to be accepted.

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

## Dependency Graph

The `DependencyGraph` tracks dependencies between entities and supports topological sorting and cycle detection:

```go
// Create a new dependency graph
dg := workspace.NewDependencyGraph()

// Add dependencies (entity A depends on entity B)
err := dg.AddDependency("file", "main.go", "file", "utils.go")
if err != nil {
    // Error if adding this would create a circular dependency
    log.Printf("Circular dependency detected: %v", err)
}

// Check direct dependencies
deps := dg.GetDependencies("file", "main.go")  // Returns ["file:utils.go"]

// Check what depends on an entity (reverse lookup)
dependents := dg.GetDependents("file", "utils.go")  // Returns ["file:main.go"]

// Get all transitive dependencies
transitive := dg.GetTransitiveDependencies("file", "main.go")

// Topological sort - dependencies first
sorted, err := dg.TopologicalSort()
if err != nil {
    log.Printf("Cycle in graph: %v", err)
}
// sorted: ["file:utils.go", "file:main.go"] - dependencies come before dependents

// Remove a specific dependency
dg.RemoveDependency("file", "main.go", "file", "utils.go")

// Remove an entity and all its dependencies
dg.RemoveEntity("file", "utils.go")

// Get dependency count
count := dg.Count()

// Clear all dependencies
dg.Clear()
```

### Circular Dependency Detection

The graph automatically prevents circular dependencies:

```go
dg := workspace.NewDependencyGraph()

// A depends on B
dg.AddDependency("file", "a", "file", "b")  // OK

// B depends on C
dg.AddDependency("file", "b", "file", "c")  // OK

// C depends on A would create a cycle: A -> B -> C -> A
err := dg.AddDependency("file", "c", "file", "a")
// err: "adding this dependency would create a cycle"

// Self-references are also detected
err = dg.AddDependency("file", "x", "file", "x")
// err: "adding this dependency would create a cycle"
```

## Concurrent Entity Processing

The workspace supports concurrent processing of entities for improved performance with large entity sets:

```go
// Add multiple entities concurrently (limit to 4 concurrent operations)
entities := []ast.Entity{entity1, entity2, entity3, ...}
results := ws.AddEntitiesBatch(entities, 4)

// Check results
for _, r := range results {
    if r.Error != nil {
        log.Printf("Failed to add %s: %v", r.Entity.Name(), r.Error)
    }
}

// Update multiple entities concurrently
results = ws.UpdateEntitiesBatch(updates, 4)

// Upsert multiple entities (add or update)
results = ws.UpsertEntitiesBatch(entities, 4)

// Process entities with a custom function
results = ws.ProcessEntitiesConcurrently(entities, func(e ast.Entity) error {
    // Your processing logic
    return nil
}, 4)

// Execute function for each entity in the workspace
results = ws.ForEachEntity(func(e ast.Entity) error {
    log.Printf("Processing: %s", e.Name())
    return nil
}, 4)

// Execute only for specific entity type
results = ws.ForEachEntityOfType("file", func(e ast.Entity) error {
    // Process only file entities
    return nil
}, 4)
```

### Transformation and Filtering

```go
// Transform entities matching a predicate
transformed, errors := ws.TransformEntities(
    // Predicate: which entities to transform
    func(e ast.Entity) bool {
        return e.Type() == "file"
    },
    // Transformer: how to transform
    func(e ast.Entity) (ast.Entity, error) {
        newEntity, _ := ast.NewEntity(e.Type(), e.Name())
        newEntity.SetProperty("processed", ast.BoolValue{Value: true})
        return newEntity, nil
    },
    4, // Max concurrency
)

// Filter entities concurrently (useful for expensive predicates)
filtered := ws.FilterEntitiesConcurrently(func(e ast.Entity) bool {
    // Expensive check that benefits from parallelism
    return expensiveCheck(e)
}, 4)
```

### Concurrency Control

All concurrent operations accept a `maxConcurrency` parameter:
- `0` or negative: Unlimited concurrency
- Positive value: Maximum number of concurrent operations

Results are returned in the same order as input entities, regardless of processing order.

## Entity Transformation Pipeline

Define multi-stage transformation pipelines for processing entities:

```go
// Create a new pipeline
pipeline := workspace.NewPipeline("file-processor")

// Add transformation stages
pipeline.AddStage("normalize-name", func(e ast.Entity) (ast.Entity, error) {
    // Transform the entity
    e.SetProperty("normalized", ast.BoolValue{Value: true})
    return e, nil
})

// Add conditional stages that only apply to matching entities
pipeline.AddConditionalStage("go-files-only",
    func(e ast.Entity) bool {
        return strings.HasSuffix(e.Name(), ".go")
    },
    func(e ast.Entity) (ast.Entity, error) {
        e.SetProperty("is-go-file", ast.BoolValue{Value: true})
        return e, nil
    })

// Execute on a single entity
entity, _ := ast.NewEntity("file", "main.go")
result := pipeline.Execute(entity)

if result.Error != nil {
    log.Printf("Failed at stage %s: %v", result.FailedStageName, result.Error)
} else {
    fmt.Printf("Executed: %v, Skipped: %v\n", result.StagesExecuted, result.StagesSkipped)
}

// Execute on multiple entities
entities := []ast.Entity{entity1, entity2, entity3}
results := pipeline.ExecuteAll(entities)
```

### Workspace Integration

```go
// Execute pipeline on workspace entities
results := ws.ExecutePipeline(pipeline, func(e ast.Entity) bool {
    return e.Type() == "file"  // Only process file entities
})

// Execute and automatically update entities in the workspace
results, err := ws.ExecutePipelineAndUpdate(pipeline, nil)  // nil = all entities
if err != nil {
    log.Printf("Some updates failed: %v", err)
}
```

### Pipeline Result

Each `PipelineResult` contains:
- `OriginalEntity`: The entity before transformation
- `ResultEntity`: The entity after all stages
- `StagesExecuted`: Names of stages that ran
- `StagesSkipped`: Names of stages skipped (predicate failed)
- `Error`: First error encountered (if any)
- `FailedStageName`: Name of the stage that failed

## Features

### Entity Management
- Add/Remove/Update entities with lifecycle hooks
- Query by type or property
- Relationship tracking
- Validation on addition/update
- Upsert support (add or update)
- Entity versioning with full history tracking
- Custom entity validators (type-specific and global)

### Concurrent Processing
- Batch add/update/upsert operations
- Concurrent entity processing with configurable limits
- Parallel transformation and filtering
- Thread-safe operations

### Transformation Pipeline
- Multi-stage transformation pipelines
- Conditional stages with predicates
- Automatic workspace updates
- Detailed execution results

### Dependency Tracking
- Entity dependency graph
- Circular dependency detection
- Topological sorting
- Transitive dependency resolution

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
- Use custom validators for type-specific rules
- Use pipelines for complex multi-step transformations
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
