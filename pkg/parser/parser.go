package parser

import (
	"fmt"
	"strings"

	"github.com/shellkjell/langspace/pkg/ast"
	"github.com/shellkjell/langspace/pkg/tokenizer"
)

// Package parser implements the LangSpace language parser, responsible for converting
// raw input text into structured AST entities. The parser follows a recursive descent
// approach and provides detailed error reporting with line and column information.

// ParseError represents a parsing error with location information
type ParseError struct {
	Line    int    // Line number where the error occurred
	Column  int    // Column number where the error occurred
	Message string // Error message
}

// Error implements the error interface
func (e ParseError) Error() string {
	return fmt.Sprintf("at line %d, col %d: %s", e.Line, e.Column, e.Message)
}

// ParseResult contains the result of parsing, including any recovered errors
type ParseResult struct {
	Entities []ast.Entity // Successfully parsed entities
	Errors   []ParseError // Errors encountered during parsing
}

// HasErrors returns true if there were any parsing errors
func (r ParseResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// ErrorString returns all errors as a single string
func (r ParseResult) ErrorString() string {
	if len(r.Errors) == 0 {
		return ""
	}
	var msgs []string
	for _, e := range r.Errors {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "; ")
}

// Parser represents a LangSpace parser instance that processes input text and
// generates AST entities. It maintains internal state during parsing and uses
// a tokenizer for lexical analysis.
type Parser struct {
	input         string               // Raw input text to be parsed
	tokenizer     *tokenizer.Tokenizer // Tokenizer instance for lexical analysis
	errorRecovery bool                 // Whether to attempt error recovery
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
		input:         input,
		tokenizer:     tokenizer.New(),
		errorRecovery: false,
	}
}

// WithErrorRecovery enables error recovery mode, allowing the parser
// to continue parsing after encountering errors and collect multiple
// errors in a single pass.
func (p *Parser) WithErrorRecovery() *Parser {
	p.errorRecovery = true
	return p
}

// ParseWithRecovery processes the input and returns a ParseResult containing
// both successfully parsed entities and any errors encountered.
// When error recovery is enabled, the parser will attempt to skip past
// errors and continue parsing subsequent entities.
func (p *Parser) ParseWithRecovery() ParseResult {
	result := ParseResult{
		Entities: make([]ast.Entity, 0),
		Errors:   make([]ParseError, 0),
	}

	allTokens := p.tokenizer.Tokenize(p.input)
	if len(allTokens) == 0 {
		return result
	}

	// Filter out comment tokens
	tokens := make([]tokenizer.Token, 0, len(allTokens))
	for _, t := range allTokens {
		if t.Type != tokenizer.TokenTypeComment {
			tokens = append(tokens, t)
		}
	}

	if len(tokens) == 0 {
		return result
	}

	i := 0
	for i < len(tokens) {
		entity, newIndex, err := p.parseEntity(tokens, i)
		if err != nil {
			result.Errors = append(result.Errors, *err)
			// Recovery: skip to next semicolon or end of input
			i = p.skipToRecoveryPoint(tokens, i)
			continue
		}
		if entity != nil {
			result.Entities = append(result.Entities, entity)
		}
		i = newIndex
	}

	return result
}

// skipToRecoveryPoint advances the token index past the next semicolon
// to attempt recovery after an error
func (p *Parser) skipToRecoveryPoint(tokens []tokenizer.Token, start int) int {
	for i := start; i < len(tokens); i++ {
		if tokens[i].Type == tokenizer.TokenTypeSemicolon {
			return i + 1
		}
	}
	return len(tokens)
}

// parseEntity parses a single entity from the token stream
// Returns the entity, the new token index, and any error
func (p *Parser) parseEntity(tokens []tokenizer.Token, i int) (ast.Entity, int, *ParseError) {
	if i+2 >= len(tokens) {
		return nil, i, &ParseError{
			Line:    tokens[i].Line,
			Column:  tokens[i].Column,
			Message: "unexpected end of input",
		}
	}

	if tokens[i].Type != tokenizer.TokenTypeIdentifier {
		return nil, i, &ParseError{
			Line:    tokens[i].Line,
			Column:  tokens[i].Column,
			Message: "expected entity type",
		}
	}

	entity, err := ast.NewEntity(tokens[i].Value)
	if err != nil {
		return nil, i, &ParseError{
			Line:    tokens[i].Line,
			Column:  tokens[i].Column,
			Message: err.Error(),
		}
	}
	i++

	// Handle path
	if tokens[i].Type != tokenizer.TokenTypeString {
		return nil, i, &ParseError{
			Line:    tokens[i].Line,
			Column:  tokens[i].Column,
			Message: "expected string",
		}
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
				return nil, i, &ParseError{
					Line:    tokens[i].Line,
					Column:  tokens[i].Column,
					Message: err.Error(),
				}
			}
			if err := entity.AddProperty("contents"); err != nil {
				return nil, i, &ParseError{
					Line:    tokens[i].Line,
					Column:  tokens[i].Column,
					Message: err.Error(),
				}
			}
			i++
		} else {
			// Handle "contents" without multiline content
			if err := entity.AddProperty(path); err != nil {
				return nil, i, &ParseError{
					Line:    tokens[i-1].Line,
					Column:  tokens[i-1].Column,
					Message: err.Error(),
				}
			}
			if err := entity.AddProperty("contents"); err != nil {
				return nil, i, &ParseError{
					Line:    tokens[i-1].Line,
					Column:  tokens[i-1].Column,
					Message: err.Error(),
				}
			}
		}
	} else if i < len(tokens) && tokens[i].Type == tokenizer.TokenTypeIdentifier {
		// For regular content, use the original path and identifier as property
		if err := entity.AddProperty(path); err != nil {
			return nil, i, &ParseError{
				Line:    tokens[i].Line,
				Column:  tokens[i].Column,
				Message: err.Error(),
			}
		}
		if err := entity.AddProperty(tokens[i].Value); err != nil {
			return nil, i, &ParseError{
				Line:    tokens[i].Line,
				Column:  tokens[i].Column,
				Message: err.Error(),
			}
		}
		i++
	} else {
		col := 0
		line := 0
		if i < len(tokens) {
			line = tokens[i].Line
			col = tokens[i].Column
		} else if i > 0 {
			line = tokens[i-1].Line
			col = tokens[i-1].Column
		}
		return nil, i, &ParseError{
			Line:    line,
			Column:  col,
			Message: "expected property",
		}
	}

	if i >= len(tokens) || tokens[i].Type != tokenizer.TokenTypeSemicolon {
		col := 0
		line := 0
		if i < len(tokens) {
			line = tokens[i].Line
			col = tokens[i].Column
		} else if i > 0 {
			line = tokens[i-1].Line
			col = tokens[i-1].Column
		}
		return nil, i, &ParseError{
			Line:    line,
			Column:  col,
			Message: "expected semicolon",
		}
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

	return entity, i, nil
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
	result := p.ParseWithRecovery()
	if result.HasErrors() {
		// Return the first error for backward compatibility
		return nil, result.Errors[0]
	}
	return result.Entities, nil
}
