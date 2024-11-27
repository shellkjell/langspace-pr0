# LangSpace

[![Go Report Card](https://goreportcard.com/badge/github.com/shellkjell/langspace)](https://goreportcard.com/report/github.com/shellkjell/langspace)
[![GoDoc](https://godoc.org/github.com/shellkjell/langspace?status.svg)](https://godoc.org/github.com/shellkjell/langspace)
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
import "github.com/username/langspace"

// Create a new workspace
workspace := langspace.NewWorkspace()

// Parse entities
input := `
file config.json contents "{\"key\": \"value\"}";
agent validator instruction "validate(config.json)";
`

parser := langspace.NewParser(input)
entities, err := parser.Parse()
if err != nil {
    log.Fatal(err)
}

// Add entities to workspace
for _, entity := range entities {
    if err := workspace.AddEntity(entity); err != nil {
        log.Fatal(err)
    }
}
```

## Installation

```bash
go get github.com/username/langspace
```

## Language Syntax

LangSpace uses a clean, declarative syntax:

```langspace
# File declaration
file example.txt contents "Hello, World!";

# Agent declaration
agent validator instruction "validate(example.txt)";

# Multi-line content
file script.sh contents "#!/bin/bash
echo 'Starting script'
./run-tests.sh";
```

## Performance

LangSpace is designed for high performance:

| Operation | Time | Memory | Allocations |
|-----------|------|---------|------------|
| Small Input (3 entities) | 1.9μs | 280B | 12 |
| Large Input (200 entities) | 137μs | 20.5KB | 609 |

See [PERFORMANCE.md](PERFORMANCE.md) for detailed benchmarks and optimization strategies.

## Architecture

LangSpace follows clean architecture principles:

```
langspace/
├── entity/     # Core entity types and interfaces
├── parser/     # Language parser and token management
├── workspace/  # Workspace management and operations
└── validator/  # Entity validation and error reporting
```

## Project Status

Current version: v0.1.0

See [ROADMAP.md](ROADMAP.md) for planned features and development timeline.

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) before submitting PRs.

### Development Requirements

- Go 1.21 or higher
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