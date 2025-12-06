# Tokenizer Package

The `tokenizer` package implements lexical analysis for LangSpace, converting raw text input into a stream of tokens. It provides efficient token management and detailed position tracking.

## Overview

The tokenizer is responsible for:
- Breaking input text into tokens
- Classifying token types
- Tracking line and column positions
- Managing token memory efficiently

## Usage

```go
import "github.com/shellkjell/langspace/pkg/tokenizer"

// Create a new tokenizer
t := tokenizer.New()

// Tokenize input
input := `file config.json contents "Hello";`
tokens := t.Tokenize(input)

// Process tokens
for _, token := range tokens {
    fmt.Printf("Token: %s, Type: %s, Line: %d, Col: %d\n",
        token.Value, token.Type, token.Line, token.Column)
}
```

## Token Types

- `TokenTypeIdentifier`: Entity types, property names, and keywords
- `TokenTypeString`: String literals (double-quoted)
- `TokenTypeMultilineString`: Multi-line content (triple backticks)
- `TokenTypeSemicolon`: Statement terminators
- `TokenTypeComment`: Single-line comments (starting with `#`)

Note: Whitespace is automatically skipped during tokenization and is not represented as a token type.

### Comments

Comments start with `#` and continue to the end of the line:

```langspace
# This is a full-line comment
file "config.json" contents;  # This is an inline comment
```

## Features

### Position Tracking
- Line numbers (1-based)
- Column positions (0-based)
- Support for multi-line tokens

### Memory Management
- Efficient token representation
- Minimal allocations
- Automatic cleanup

### Error Handling
- Invalid character detection
- Unterminated string handling
- Position information in errors

## Best Practices

- Reuse tokenizer instances when possible
- Handle multi-line content appropriately
- Check for tokenization errors
- Use position information for error reporting

## Performance

The tokenizer is optimized for:
- Linear time complexity
- Minimal memory usage
- Fast token classification
- Efficient string handling

## Future Enhancements

Planned improvements include:
- Streaming tokenization
- Custom token type support
- Enhanced error recovery
- Token position caching
