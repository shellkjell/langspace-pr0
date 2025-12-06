# Parser Package

The `parser` package implements LangSpace's parsing logic, converting raw text input into structured AST entities. It provides detailed error reporting with line and column information, and supports error recovery for collecting multiple errors in a single pass.

## Overview

The parser follows a recursive descent approach and works closely with the tokenizer to process input text. Key components include:

- `Parser`: Main parsing engine
- `ParseError`: Error type with line/column information
- `ParseResult`: Container for entities and errors
- Error recovery for robust parsing
- Entity construction and validation

## Usage

### Basic Parsing

```go
import "github.com/shellkjell/langspace/pkg/parser"

// Create a new parser
input := `
file "config.json" contents;
agent "validator" instruction;
`
p := parser.New(input)

// Parse input into entities
entities, err := p.Parse()
if err != nil {
    log.Fatalf("Parse error: %v", err)
}

// Process entities
for _, entity := range entities {
    fmt.Printf("Entity type: %s\n", entity.Type())
}
```

### Error Recovery Mode

For better error reporting, use error recovery to collect all errors in one pass:

```go
p := parser.New(input).WithErrorRecovery()
result := p.ParseWithRecovery()

// Check for errors
if result.HasErrors() {
    for _, err := range result.Errors {
        fmt.Printf("Error at line %d, col %d: %s\n",
            err.Line, err.Column, err.Message)
    }
}

// Still process successfully parsed entities
for _, entity := range result.Entities {
    fmt.Printf("Parsed: %s\n", entity.Type())
}

// Or get all errors as a single string
if result.HasErrors() {
    log.Printf("Parsing had errors: %s", result.ErrorString())
}
```

## Parsing Process

1. **Tokenization**: Input is broken down into tokens
2. **Comment Filtering**: Comment tokens are filtered out
3. **Entity Recognition**: Tokens are analyzed to identify entity types
4. **Property Collection**: Entity properties are gathered and validated
5. **AST Construction**: Valid entities are constructed into an AST
6. **Error Recovery**: On error, skip to next semicolon and continue

## Error Handling

The parser provides detailed error information:
- Line and column numbers for error locations
- Descriptive error messages
- Context about the expected vs. actual token
- Multiple error collection with recovery mode

### ParseError Type

```go
type ParseError struct {
    Line    int    // Line number (1-indexed)
    Column  int    // Column number (1-indexed)
    Message string // Error description
}
```

### ParseResult Type

```go
type ParseResult struct {
    Entities []ast.Entity // Successfully parsed entities
    Errors   []ParseError // All errors encountered
}

// Helper methods
result.HasErrors() bool       // True if any errors occurred
result.ErrorString() string   // All errors as semicolon-separated string
```

## Syntax Rules

### Block Entity Declaration (Preferred)
```langspace
<entity_type> "<name>" {
  property1: value1
  property2: value2
}
```

### Legacy Entity Declaration
```
<entity_type> "<name>" <property>;
```

### Comments
```langspace
# This is a comment
agent "validator" {  # Inline comment
  model: "claude-sonnet-4-20250514"
}
```

### Block Syntax Examples
```langspace
file "config.json" {
  path: "./config/app.json"
}

agent "reviewer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3
  instruction: ```
    You are a code reviewer.
  ```
  tools: [read_file, search_codebase]
}

script "update-db" {
  language: "python"
  runtime: "python3"
  capabilities: [database]
  code: ```python
    import db
    db.update("table", data)
  ```
}
```

### Legacy Syntax Examples (Backward Compatible)
```
file "config.json" contents;
agent "validator" instruction;
```

## Best Practices

- Use `ParseWithRecovery()` for better error reporting in IDEs/editors
- Always check for parsing errors
- Use the error location information for debugging
- Consider input size when parsing large files
- Handle multi-line content appropriately
- Use comments to document your LangSpace files
