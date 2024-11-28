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

- `TokenTypeIdentifier`: Entity types and names
- `TokenTypeString`: String literals
- `TokenTypeSemicolon`: Statement terminators
- `TokenTypeWhitespace`: Spaces and newlines (filtered)

## Features

### Position Tracking
- Line numbers (1-based)
- Column positions (0-based)
- Support for multi-line tokens

### Memory Management
- Token pooling for efficiency
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
