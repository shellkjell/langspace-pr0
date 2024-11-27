package parser

import (
	"fmt"

	"github.com/shellkjell/langspace/pkg/ast"
	"github.com/shellkjell/langspace/pkg/tokenizer"
)

// Parser represents a LangSpace parser instance
type Parser struct {
	input     string
	tokenizer *tokenizer.Tokenizer
}

// New creates a new Parser instance
func New(input string) *Parser {
	return &Parser{
		input:     input,
		tokenizer: tokenizer.New(),
	}
}

// Parse parses the input and returns a slice of entities
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

		if tokens[i].Type != tokenizer.TokenTypeString {
			return nil, fmt.Errorf("expected string at line %d, col %d", tokens[i].Line, tokens[i].Column)
		}

		if err := entity.AddProperty(tokens[i].Value); err != nil {
			return nil, fmt.Errorf("at line %d, col %d: %v", tokens[i].Line, tokens[i].Column, err)
		}
		i++

		if tokens[i].Type != tokenizer.TokenTypeIdentifier {
			return nil, fmt.Errorf("expected property at line %d, col %d", tokens[i].Line, tokens[i].Column)
		}

		if err := entity.AddProperty(tokens[i].Value); err != nil {
			return nil, fmt.Errorf("at line %d, col %d: %v", tokens[i].Line, tokens[i].Column, err)
		}
		i++

		if i >= len(tokens) || tokens[i].Type != tokenizer.TokenTypeSemicolon {
			return nil, fmt.Errorf("expected semicolon at line %d, col %d", tokens[i-1].Line, tokens[i-1].Column)
		}
		i++

		entities = append(entities, entity)
	}

	return entities, nil
}
