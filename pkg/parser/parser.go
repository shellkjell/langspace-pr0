package parser

import (
	"fmt"

	"github.com/shellkjell/langspace/pkg/ast"
	"github.com/shellkjell/langspace/pkg/tokenizer"
)

// Package parser implements the LangSpace language parser, responsible for converting
// raw input text into structured AST entities. The parser follows a recursive descent
// approach and provides detailed error reporting with line and column information.

// Parser represents a LangSpace parser instance that processes input text and
// generates AST entities. It maintains internal state during parsing and uses
// a tokenizer for lexical analysis.
type Parser struct {
	input     string       // Raw input text to be parsed
	tokenizer *tokenizer.Tokenizer // Tokenizer instance for lexical analysis
}

// New creates a new Parser instance with the given input text.
// The parser is initialized with a fresh tokenizer instance.
// Parameters:
//   - input: The raw input text to be parsed
//
// Returns:
//   - *Parser: A new parser instance ready to process the input
func New(input string) *Parser {
	return &Parser{
		input:     input,
		tokenizer: tokenizer.New(),
	}
}

// Parse processes the input text and returns a slice of parsed entities.
// The parsing follows these steps:
// 1. Tokenize the input using the tokenizer
// 2. Process tokens sequentially to build entities
// 3. Validate entity structure during parsing
//
// Returns:
//   - []ast.Entity: Slice of parsed entities
//   - error: Detailed error with line/column information if parsing fails
//
// Errors are returned for:
//   - Unexpected end of input
//   - Invalid entity types
//   - Missing or malformed properties
//   - Invalid token sequences
func (p *Parser) Parse() ([]ast.Entity, error) {
	tokens := p.tokenizer.Tokenize(p.input)
	if len(tokens) == 0 {
		return nil, nil
	}

	var entities []ast.Entity
	i := 0

	for i < len(tokens) {
		if i+2 >= len(tokens) {
			return nil, fmt.Errorf("unexpected end of input at line %d, col %d", tokens[i].Line, tokens[i].Column)
		}

		if tokens[i].Type != tokenizer.TokenTypeIdentifier {
			return nil, fmt.Errorf("expected entity type at line %d, col %d", tokens[i].Line, tokens[i].Column)
		}

		entity, err := ast.NewEntity(tokens[i].Value)
		if err != nil {
			return nil, fmt.Errorf("at line %d, col %d: %v", tokens[i].Line, tokens[i].Column, err)
		}
		i++

		// Handle path
		if tokens[i].Type != tokenizer.TokenTypeString {
			return nil, fmt.Errorf("expected string at line %d, col %d", tokens[i].Line, tokens[i].Column)
		}
		path := tokens[i].Value
		i++

		// Check for contents identifier before multiline string
		if i < len(tokens) && tokens[i].Type == tokenizer.TokenTypeIdentifier && tokens[i].Value == "contents" {
			i++ // Skip "contents" identifier
			if i < len(tokens) && tokens[i].Type == tokenizer.TokenTypeMultilineString {
				// For multiline content, the content becomes the path and "contents" is the property
				content := tokens[i].Value
				if len(content) > 0 && content[0] == '\n' {
					content = content[1:] // Trim leading newline
				}
				if err := entity.AddProperty(content); err != nil {
					return nil, fmt.Errorf("at line %d, col %d: %v", tokens[i].Line, tokens[i].Column, err)
				}
				if err := entity.AddProperty("contents"); err != nil {
					return nil, fmt.Errorf("at line %d, col %d: %v", tokens[i].Line, tokens[i].Column, err)
				}
				i++
			} else {
				// Handle "contents" without multiline content
				if err := entity.AddProperty(path); err != nil {
					return nil, fmt.Errorf("at line %d, col %d: %v", tokens[i-1].Line, tokens[i-1].Column, err)
				}
				if err := entity.AddProperty("contents"); err != nil {
					return nil, fmt.Errorf("at line %d, col %d: %v", tokens[i-1].Line, tokens[i-1].Column, err)
				}
			}
		} else if tokens[i].Type == tokenizer.TokenTypeIdentifier {
			// For regular content, use the original path and identifier as property
			if err := entity.AddProperty(path); err != nil {
				return nil, fmt.Errorf("at line %d, col %d: %v", tokens[i].Line, tokens[i].Column, err)
			}
			if err := entity.AddProperty(tokens[i].Value); err != nil {
				return nil, fmt.Errorf("at line %d, col %d: %v", tokens[i].Line, tokens[i].Column, err)
			}
			i++
		} else {
			return nil, fmt.Errorf("expected property at line %d, col %d", tokens[i].Line, tokens[i].Column)
		}

		if i >= len(tokens) || tokens[i].Type != tokenizer.TokenTypeSemicolon {
			return nil, fmt.Errorf("expected semicolon at line %d, col %d", tokens[i-1].Line, tokens[i-1].Column)
		}
		i++ // Skip semicolon

		// Skip any trailing identifiers after semicolon if there are no more tokens
		if i < len(tokens) && tokens[i].Type == tokenizer.TokenTypeIdentifier {
			// Only skip if this is the last token or the next token isn't a string
			// This ensures we don't skip the start of the next entity
			if i+1 >= len(tokens) || tokens[i+1].Type != tokenizer.TokenTypeString {
				i++
			}
		}

		entities = append(entities, entity)
	}

	return entities, nil
}
