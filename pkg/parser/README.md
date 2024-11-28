# Parser Package

The `parser` package implements LangSpace's parsing logic, converting raw text input into structured AST entities. It provides detailed error reporting with line and column information.

## Overview

The parser follows a recursive descent approach and works closely with the tokenizer to process input text. Key components include:

- `Parser`: Main parsing engine
- Error handling with location information
- Entity construction and validation

## Usage

```go
import "github.com/shellkjell/langspace/pkg/parser"

// Create a new parser
input := `
file config.json contents "{\"key\": \"value\"}";
agent validator instruction "validate(config.json)";
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

## Parsing Process

1. **Tokenization**: Input is broken down into tokens
2. **Entity Recognition**: Tokens are analyzed to identify entity types
3. **Property Collection**: Entity properties are gathered and validated
4. **AST Construction**: Valid entities are constructed into an AST

## Error Handling

The parser provides detailed error information:
- Line and column numbers for error locations
- Descriptive error messages
- Context about the expected vs. actual token

## Syntax Rules

### Entity Declaration
```
<entity_type> <property1> <property2>;
```

### Examples
```
file config.json contents "{\"key\": \"value\"}";
agent validator instruction "validate(config.json)";
```

## Best Practices

- Always check for parsing errors
- Use the error location information for debugging
- Consider input size when parsing large files
- Handle multi-line content appropriately
