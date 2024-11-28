# Abstract Syntax Tree (AST) Package

The `ast` package provides the core data structures and interfaces for representing LangSpace entities in memory. It serves as the backbone of the language's type system and entity management.

## Overview

The AST package defines the following key components:

- `Entity` interface: The fundamental building block of LangSpace
- `BaseEntity`: Common implementation shared across entity types
- `FileEntity`: Represents file system resources
- `AgentEntity`: Represents automation tasks

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
```

## Entity Types

### File Entity
- **Purpose**: Represents file system resources
- **Properties**:
  - `path`: File system path (required)
  - `contents`: File contents as string (required)
- **Validation**: Must have exactly two properties

### Agent Entity
- **Purpose**: Represents automation tasks
- **Properties**:
  - `name`: Agent identifier (required)
  - `instruction`: Task instruction (required)
- **Validation**: Must have exactly two properties

## Extension

To add new entity types:

1. Create a new struct implementing the `Entity` interface
2. Add the new type to `NewEntity` factory function
3. Implement type-specific validation rules
4. Update relevant documentation

## Best Practices

- Always validate entities after creation
- Use the `NewEntity` factory function instead of direct struct initialization
- Handle all error cases when adding properties
- Consider using composition with `BaseEntity` for new entity types
