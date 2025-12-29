# LangSpace

[![Go Report Card](https://goreportcard.com/badge/github.com/shellkjell/langspace)](https://goreportcard.com/report/github.com/shellkjell/langspace)
[![GoDoc](https://pkg.go.dev/badge/github.com/shellkjell/langspace.svg)](https://pkg.go.dev/github.com/shellkjell/langspace)
[![License](https://img.shields.io/badge/License-GPL%20v2-blue.svg)](LICENSE.md)

LangSpace is a declarative language for composing AI agent workflows. It provides a readable, versionable format for defining agents, tools, and multi-step pipelines that can be executed directly or compiled to other targets.

## Overview

LangSpace sits between writing raw Python/TypeScript and using no-code builders. It captures the full specification of an AI workflow in a single file that can be version-controlled, shared, and executed.

````langspace
agent "code-reviewer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: ```
    You are a senior code reviewer. Analyze code for security
    vulnerabilities, performance issues, and best practices.
  ```

  tools: [read_file, search_codebase]
}

intent "review-changes" {
  use: agent("code-reviewer")
  input: git.diff(base: "main")
  output: file("review.md")
}
````

## Installation

```bash
go get github.com/shellkjell/langspace
```

## Language Syntax

LangSpace uses block-based declarations with key-value properties.

### Files

Files represent static data: prompts, configuration, or output destinations.

````langspace
# Inline contents
file "prompt.md" {
  contents: ```
    You are a helpful assistant.
  ```
}

# Reference to external file
file "config.json" {
  path: "./config/app.json"
}
````

### Agents

Agents are LLM-powered actors with specific roles and capabilities.

````langspace
agent "analyst" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.5

  instruction: ```
    Analyze the provided data and generate insights.
  ```

  tools: [read_file, query_database]
}
````

### Tools

Tools extend agent capabilities by connecting to external systems.

```langspace
tool "search_codebase" {
  description: "Search for patterns in source code"

  parameters: {
    query: string required
    file_pattern: string optional
  }

  handler: mcp("filesystem-server")
}
```

### Intentions

Intentions express what you want to accomplish.

```langspace
intent "summarize-docs" {
  use: agent("summarizer")
  input: file("docs/**/*.md")
  output: file("summary.md")
}
```

### Pipelines

Pipelines chain multiple agents together with data flowing between steps.

```langspace
pipeline "analyze-and-report" {
  step "analyze" {
    use: agent("analyzer")
    input: $input
  }

  step "report" {
    use: agent("reporter")
    input: step("analyze").output
  }

  output: step("report").output
}
```

### MCP Integration

Connect to Model Context Protocol servers for tool access.

```langspace
mcp "filesystem" {
  command: "npx"
  args: ["-y", "@anthropic/mcp-filesystem", "/workspace"]
}

agent "file-manager" {
  tools: [
    mcp("filesystem").read_file,
    mcp("filesystem").write_file,
  ]
}
```

### Scripts

Scripts enable code-first agent actions — a more efficient alternative to multiple tool calls. Instead of loading full data into the context window through repeated tool invocations, agents write executable code that performs complex operations in a single execution.

````langspace
# Define a reusable script template
script "db-update" {
  language: "python"
  runtime: "python3"

  capabilities: [database]

  parameters: {
    table: string required
    id: string required
    updates: object required
  }

  code: ```python
    import db

    # Find, modify, and save in one execution
    record = db.find(table, {"id": id})
    record.update(updates)
    db.save(table, record)

    print(f"Updated {table}/{id}")
  ```

  timeout: "30s"
}

# Agent that generates and executes scripts
agent "efficient-data-manager" {
  model: "claude-sonnet-4-20250514"

  instruction: ```
    Perform database operations by writing Python scripts
    rather than making individual tool calls. This is more
    efficient and keeps the context window small.
  ```

  scripts: [script("db-update")]
}
````

**Why Scripts over Tools?**

| Approach | Context Usage | Round Trips |
|----------|---------------|-------------|
| Multiple tool calls | High (full data loaded each time) | Many |
| Single script execution | Low (only results returned) | One |

See [examples/09-scripts.ls](examples/09-scripts.ls) for more patterns.

### Configuration

Set global defaults for providers and models.

```langspace
config {
  default_model: "claude-sonnet-4-20250514"

  providers: {
    anthropic: {
      api_key: env("ANTHROPIC_API_KEY")
    }
  }
}
```

### Comments

Single-line comments start with `#`:

```langspace
# This is a comment
agent "example" { }  # Inline comment
```

## Usage

### As a Library

```go
import (
    "github.com/shellkjell/langspace/pkg/parser"
    "github.com/shellkjell/langspace/pkg/workspace"
)

// Parse LangSpace definitions
input := `
agent "reviewer" {
  model: "claude-sonnet-4-20250514"
  instruction: "Review code for issues"
}
`

p := parser.New(input)
entities, err := p.Parse()
if err != nil {
    log.Fatal(err)
}

// Add to workspace
ws := workspace.New()
for _, entity := range entities {
    ws.AddEntity(entity)
}
```

### Command Line (Planned)

```bash
# Execute a workflow
langspace run workflow.ls --input ./src

# Compile to Python
langspace compile --target python workflow.ls -o workflow.py
```

## Examples

See the [examples/](examples/) directory:

| Example | Description | Status |
|---------|-------------|--------|
| `01-hello-world.ls` | Basic agent definition | - Parses |
| `02-files.ls` | File declarations and references | - Parses |
| `03-agents.ls` | Agent configuration options | - Parses |
| `04-intentions.ls` | Expressing desired outcomes | - Parses |
| `05-pipelines.ls` | Multi-step workflows | - Parses |
| `06-tools-mcp.ls` | Tool definitions and MCP integration | - Parses |
| `07-config.ls` | Global configuration | - Parses |
| `08-complete-code-review.ls` | Full code review workflow | - Parses |
| `09-scripts.ls` | Code-first agent actions | - Parses |

> **Parser Capabilities:**
> - Nested blocks (e.g., `step { }` inside `pipeline { }`) ✅
> - Typed parameters (e.g., `query: string required "description"`) ✅
> - Inline typed blocks (e.g., `handler: http { ... }`) ✅
> - Property access chains (e.g., `params.location`) ✅
> - Inline enums (e.g., `type: enum ["a", "b", "c"]`) ✅
> - Method calls on objects (e.g., `git.staged_files()`) ✅
> - Comparison expressions (e.g., `env("X") == "true"`) ✅
> - Control flow (`branch`, `loop`, `break_if`) ✅
>
> See [examples/advanced/](examples/advanced/) for production-ready workflow patterns.

## Architecture

```
langspace/
├── cmd/langspace/     # CLI application
├── pkg/
│   ├── ast/           # Entity types and interfaces
│   ├── parser/        # Language parser
│   ├── slices/        # Generic slice utilities
│   ├── tokenizer/     # Lexical analysis
│   ├── validator/     # Entity validation
│   └── workspace/     # Workspace management
```

## Project Status

**Current Phase: Execution & Compilation**

### - Implemented & Working
- Block-based tokenizer with DSL syntax support
- Parser for entity declarations with nested blocks and typed parameters
- AST representation with entity metadata and relationships
- Nested entity parsing (pipelines with steps, tools with handlers)
- Typed parameter declarations (`name: string required "desc"`)
- Property access chains (`params.location`, `config.defaults.timeout`)
- Method calls (`git.staged_files()`) and comparison expressions
- Control flow (`branch`, `loop`, `break_if`)
- Validation framework with extensible custom validators
- Workspace management with entity hooks, events, snapshots, and persistence
- Error recovery parsing with detailed line/column error reporting
- LLM integration (Anthropic, OpenAI providers)
- Tool execution loop with shell and HTTP handlers
- MCP (Model Context Protocol) client integration
- Sandboxed Python and Shell script execution
- Trigger engine for automated workflows
- CLI commands: `parse`, `run`, `validate`, `serve`
- Comprehensive test coverage (100+ tests)

### - In Progress / Planned
- Compilation targets (Python/LangGraph, TypeScript)
- `langspace compile` command
- Syntax highlighting (VSCode extension)
- Authentication and authorization

See [ROADMAP.md](ROADMAP.md) for the full development plan and [PRD.md](PRD.md) for the product specification.

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Requirements

- Go 1.23+
- Make (optional, for build automation)

### Running Tests

```bash
go test ./...
go test -bench=. -benchmem ./...
go test -race ./...
```

## License

[GNU GPL v2](LICENSE.md)