# LangSpace

[![Go Report Card](https://goreportcard.com/badge/github.com/shellkjell/langspace-pr0)](https://goreportcard.com/report/github.com/shellkjell/langspace-pr0)
[![GoDoc](https://godoc.org/github.com/shellkjell/langspace-pr0?status.svg)](https://godoc.org/github.com/shellkjell/langspace-pr0)
[![License](https://img.shields.io/badge/License-GPL%20v2-blue.svg)](LICENSE.md)

LangSpace is a high-performance, extensible domain-specific language (DSL) designed for managing virtual workspaces. It provides a type-safe, memory-efficient way to declare and manipulate workspace entities with comprehensive error reporting.

## Key Features

- **🚀 High Performance**
  - Memory-efficient token pooling
  - Linear scaling with input size (~0.7μs/entity)
  - Optimized memory allocation (~100 bytes/entity)

- **🛡️ Type Safety**
  - Strongly typed entity system
  - Comprehensive error reporting with line/column information
  - Validation at parse time

- **🔌 Extensibility**
  - Pluggable entity type system
  - Custom entity validators
  - Event hooks for entity lifecycle

- **📝 Modern Syntax**
  - Clean, declarative syntax
  - Multi-line string support
  - Rich error messages

## Quick Start

```go
import (
    "github.com/shellkjell/langspace/pkg/parser"
    "github.com/shellkjell/langspace/pkg/workspace"
)

// Create a new workspace
ws := workspace.New()

// Parse entities
input := `
file "config.json" contents;
agent "validator" instruction;
task "build" instruction;
`

p := parser.New(input)
entities, err := p.Parse()
if err != nil {
    log.Fatal(err)
}

// Add entities to workspace
for _, entity := range entities {
    if err := ws.AddEntity(entity); err != nil {
        log.Fatal(err)
    }
}
```

## Installation

```bash
go get github.com/shellkjell/langspace
```

## Language Syntax

LangSpace uses a clean, declarative syntax:

```langspace
# File declaration with path property
file "example.txt" path;

# File declaration with contents property
file "config.json" contents;

# Agent declaration
agent "validator" instruction;
agent "gpt-4" model;

# Task declaration
task "build" instruction;
task "backup" schedule;
task "urgent" priority;

# Multi-line content using triple backticks
file "script.sh" contents ```
#!/bin/bash
echo 'Starting script'
./run-tests.sh
```;
```

### Comments

LangSpace supports single-line comments starting with `#`:

```langspace
# This is a comment
file "config.json" contents;  # Inline comment
```

## Performance

LangSpace is designed for high performance:

| Operation | Time | Memory | Allocations |
|-----------|------|---------|------------|
| Small Input (3 entities) | ~594 ns | 1.6 KB | 13 |
| Large Input (200 entities) | ~27 μs | 103 KB | 222 |

See [PERFORMANCE.md](PERFORMANCE.md) for detailed benchmarks and optimization strategies.

## Architecture

LangSpace follows clean architecture principles:

```
langspace/
├── cmd/
│   └── langspace/  # CLI application
├── internal/
│   └── pool/       # Memory-efficient token pooling
├── pkg/
│   ├── ast/        # Core entity types and interfaces
│   ├── parser/     # Language parser
│   ├── tokenizer/  # Lexical analysis and token management
│   ├── validator/  # Entity validation and error reporting
│   └── workspace/  # Workspace management and operations
```

## Project Status

Current version: v0.1.0

See [ROADMAP.md](ROADMAP.md) for planned features and development timeline.

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) before submitting PRs.

### Development Requirements

- Go 1.23 or higher
- Make (for build automation)
- golangci-lint (for code quality)

### Running Tests

```bash
# Run all tests
go test ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Run with race detection
go test -race ./...
```

## License

LangSpace is licensed under the [GNU GPL v2](LICENSE.md).

## Acknowledgments

Special thanks to:
- The Go team for the excellent standard library
- Our contributors and users for their valuable feedback