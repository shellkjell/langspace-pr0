# Abstract Syntax Tree (AST) Package

The `ast` package provides the core data structures and interfaces for representing LangSpace entities in memory. It serves as the backbone of the language's type system and entity management.

## Overview

The AST package defines the following key components:

- `Entity` interface: The fundamental building block of LangSpace
- `BaseEntity`: Common implementation shared across entity types
- `FileEntity`: Represents file system resources
- `AgentEntity`: Represents automation tasks
- `TaskEntity`: Represents task management

## Usage

```go
import "github.com/shellkjell/langspace/pkg/ast"

// Create a new file entity
entity, err := ast.NewEntity("file")
if err != nil {
    log.Fatal(err)
}

// Add properties
err = entity.AddProperty("config.json")
if err != nil {
    log.Fatal(err)
}
err = entity.AddProperty("contents")
if err != nil {
    log.Fatal(err)
}

// Add metadata
entity.SetMetadata("author", "john.doe")
entity.SetMetadata("version", "1.0")

// Retrieve metadata
author, ok := entity.GetMetadata("author")
if ok {
    fmt.Printf("Author: %s\n", author)
}

// Get all metadata
allMeta := entity.AllMetadata()
for key, value := range allMeta {
    fmt.Printf("%s: %s\n", key, value)
}
```

## Entity Types

### File Entity
- **Purpose**: Represents file system resources
- **Properties**:
  - `path`: File system path or name (required)
  - `property`: Property type such as `path` or `contents` (required)
- **Validation**: Must have exactly two properties

### Agent Entity
- **Purpose**: Represents automation tasks
- **Properties**:
  - `name`: Agent identifier (required)
  - `property`: Property type such as `instruction`, `model`, or `check(filename)` (required)
- **Validation**: Must have exactly two properties

### Task Entity
- **Purpose**: Represents task management and automation
- **Properties**:
  - `name`: Task identifier (required)
  - `property`: Property type such as `instruction`, `schedule`, or `priority` (required)
- **Validation**: Must have exactly two properties

## Entity Metadata

All entities support key-value metadata for storing additional information:

```go
// Set metadata
entity.SetMetadata("key", "value")

// Get metadata (returns value and existence flag)
value, exists := entity.GetMetadata("key")

// Get all metadata as a map (returns a copy)
allMetadata := entity.AllMetadata()
```

**Use cases for metadata:**
- Tracking entity creation time or author
- Storing version information
- Adding custom tags or labels
- Associating external identifiers

## Extension

To add new entity types:

1. Create a new struct implementing the `Entity` interface
2. Add the new type to `NewEntity` factory function
3. Implement type-specific validation rules
4. Include metadata support (metadata map field + interface methods)
5. Update relevant documentation

## Best Practices

- Always validate entities after creation
- Use the `NewEntity` factory function instead of direct struct initialization
- Handle all error cases when adding properties
- Consider using composition with `BaseEntity` for new entity types
- Use metadata for extensible, non-core properties
- Remember that `AllMetadata()` returns a copy to prevent unintended modifications
