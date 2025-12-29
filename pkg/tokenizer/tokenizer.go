// Package tokenizer implements lexical analysis for the LangSpace language.
// It converts raw input text into a stream of tokens that can be processed by the parser.
package tokenizer

import (
	"unicode"
)

// TokenType represents the type of a token
type TokenType int

const (
	// TokenTypeIdentifier represents an identifier (e.g., name, type)
	TokenTypeIdentifier TokenType = iota
	// TokenTypeString represents a single-line string literal
	TokenTypeString
	// TokenTypeSemicolon represents a semicolon (;)
	TokenTypeSemicolon
	// TokenTypeMultilineString represents a multiline string literal
	TokenTypeMultilineString
	// TokenTypeComment represents a comment (# ...)
	TokenTypeComment
	// TokenTypeLeftBrace represents an opening brace ({)
	TokenTypeLeftBrace
	// TokenTypeRightBrace represents a closing brace (})
	TokenTypeRightBrace
	// TokenTypeLeftBracket represents an opening bracket ([)
	TokenTypeLeftBracket
	// TokenTypeRightBracket represents a closing bracket (])
	TokenTypeRightBracket
	// TokenTypeLeftParen represents an opening parenthesis (()
	TokenTypeLeftParen
	// TokenTypeRightParen represents a closing parenthesis ())
	TokenTypeRightParen
	// TokenTypeColon represents a colon (:)
	TokenTypeColon
	// TokenTypeComma represents a comma (,)
	TokenTypeComma
	// TokenTypeDot represents a dot (.)
	TokenTypeDot
	// TokenTypeEquals represents an equals sign (=)
	TokenTypeEquals
	// TokenTypeArrow represents an arrow (=>)
	TokenTypeArrow
	// TokenTypeDoubleEquals represents a double equals sign (==)
	TokenTypeDoubleEquals
	// TokenTypeNotEquals represents a not equals sign (!=)
	TokenTypeNotEquals
	// TokenTypeLess represents a less than sign (<)
	TokenTypeLess
	// TokenTypeGreater represents a greater than sign (>)
	TokenTypeGreater
	// TokenTypeLessEquals represents a less than or equal sign (<=)
	TokenTypeLessEquals
	// TokenTypeGreaterEquals represents a greater than or equal sign (>=)
	TokenTypeGreaterEquals
	// TokenTypeDollar represents a dollar sign ($)
	TokenTypeDollar
	// TokenTypeNumber represents a numeric literal
	TokenTypeNumber
	// TokenTypeBoolean represents a boolean literal (true/false)
	TokenTypeBoolean
)

// Token represents a lexical token
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

// Tokenizer represents a LangSpace tokenizer
type Tokenizer struct {
	// Note: pool was removed as it was not being utilized.
	// If memory optimization becomes necessary for large inputs,
	// consider implementing string interning or token reuse.
}

// New creates a new Tokenizer instance
func New() *Tokenizer {
	return &Tokenizer{}
}

// Tokenize breaks the input string into tokens
func (t *Tokenizer) Tokenize(input string) []Token {
	var tokens []Token
	line := 1
	column := 1
	i := 0

	for i < len(input) {
		switch {
		case input[i] == '#':
			// Handle single-line comments
			startCol := column
			start := i
			for i < len(input) && input[i] != '\n' {
				i++
				column++
			}
			tokens = append(tokens, Token{
				Type:   TokenTypeComment,
				Value:  input[start:i],
				Line:   line,
				Column: startCol,
			})
			// Note: newline will be processed in next iteration

		case unicode.IsSpace(rune(input[i])):
			if input[i] == '\n' {
				line++
				column = 1
			} else {
				column++
			}
			i++

		case input[i] == '{':
			tokens = append(tokens, Token{
				Type:   TokenTypeLeftBrace,
				Value:  "{",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == '}':
			tokens = append(tokens, Token{
				Type:   TokenTypeRightBrace,
				Value:  "}",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == '[':
			tokens = append(tokens, Token{
				Type:   TokenTypeLeftBracket,
				Value:  "[",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == ']':
			tokens = append(tokens, Token{
				Type:   TokenTypeRightBracket,
				Value:  "]",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == '(':
			tokens = append(tokens, Token{
				Type:   TokenTypeLeftParen,
				Value:  "(",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == ')':
			tokens = append(tokens, Token{
				Type:   TokenTypeRightParen,
				Value:  ")",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == ':':
			tokens = append(tokens, Token{
				Type:   TokenTypeColon,
				Value:  ":",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == ',':
			tokens = append(tokens, Token{
				Type:   TokenTypeComma,
				Value:  ",",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == '.':
			tokens = append(tokens, Token{
				Type:   TokenTypeDot,
				Value:  ".",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == '$':
			tokens = append(tokens, Token{
				Type:   TokenTypeDollar,
				Value:  "$",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == '=' && i+1 < len(input) && input[i+1] == '>':
			tokens = append(tokens, Token{
				Type:   TokenTypeArrow,
				Value:  "=>",
				Line:   line,
				Column: column,
			})
			i += 2
			column += 2

		case input[i] == '=' && i+1 < len(input) && input[i+1] == '=':
			tokens = append(tokens, Token{
				Type:   TokenTypeDoubleEquals,
				Value:  "==",
				Line:   line,
				Column: column,
			})
			i += 2
			column += 2

		case input[i] == '!' && i+1 < len(input) && input[i+1] == '=':
			tokens = append(tokens, Token{
				Type:   TokenTypeNotEquals,
				Value:  "!=",
				Line:   line,
				Column: column,
			})
			i += 2
			column += 2

		case input[i] == '<' && i+1 < len(input) && input[i+1] == '=':
			tokens = append(tokens, Token{
				Type:   TokenTypeLessEquals,
				Value:  "<=",
				Line:   line,
				Column: column,
			})
			i += 2
			column += 2

		case input[i] == '>' && i+1 < len(input) && input[i+1] == '=':
			tokens = append(tokens, Token{
				Type:   TokenTypeGreaterEquals,
				Value:  ">=",
				Line:   line,
				Column: column,
			})
			i += 2
			column += 2

		case input[i] == '<':
			tokens = append(tokens, Token{
				Type:   TokenTypeLess,
				Value:  "<",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == '>':
			tokens = append(tokens, Token{
				Type:   TokenTypeGreater,
				Value:  ">",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == '=':
			tokens = append(tokens, Token{
				Type:   TokenTypeEquals,
				Value:  "=",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case unicode.IsDigit(rune(input[i])) || (input[i] == '-' && i+1 < len(input) && unicode.IsDigit(rune(input[i+1]))):
			start := i
			startCol := column
			if input[i] == '-' {
				i++
				column++
			}
			for i < len(input) && (unicode.IsDigit(rune(input[i])) || input[i] == '.') {
				i++
				column++
			}
			tokens = append(tokens, Token{
				Type:   TokenTypeNumber,
				Value:  input[start:i],
				Line:   line,
				Column: startCol,
			})

		case unicode.IsLetter(rune(input[i])) || input[i] == '_':
			start := i
			startCol := column
			for i < len(input) && (unicode.IsLetter(rune(input[i])) || unicode.IsDigit(rune(input[i])) || input[i] == '_' || input[i] == '-') {
				i++
				column++
			}
			value := input[start:i]
			// Check for boolean literals
			if value == "true" || value == "false" {
				tokens = append(tokens, Token{
					Type:   TokenTypeBoolean,
					Value:  value,
					Line:   line,
					Column: startCol,
				})
			} else {
				tokens = append(tokens, Token{
					Type:   TokenTypeIdentifier,
					Value:  value,
					Line:   line,
					Column: startCol,
				})
			}

		case i+2 < len(input) && input[i] == '`' && input[i+1] == '`' && input[i+2] == '`':
			startCol := column
			startLine := line
			i += 3 // Skip opening triple backticks
			column += 3
			start := i

			// Find closing triple backticks
			for i < len(input) {
				if i+2 < len(input) && input[i] == '`' && input[i+1] == '`' && input[i+2] == '`' {
					break
				}
				if input[i] == '\n' {
					line++
					column = 1
				} else {
					column++
				}
				i++
			}

			if i+2 < len(input) {
				tokens = append(tokens, Token{
					Type:   TokenTypeMultilineString,
					Value:  input[start:i],
					Line:   startLine,
					Column: startCol,
				})
				i += 3 // Skip closing triple backticks
				column += 3
			}

		case input[i] == '"':
			startCol := column
			i++ // Skip opening quote
			column++
			start := i
			for i < len(input) && input[i] != '"' {
				if input[i] == '\\' && i+1 < len(input) {
					// Skip escaped character
					i += 2
					column += 2
					continue
				}
				if input[i] == '\n' {
					line++
					column = 1
				} else {
					column++
				}
				i++
			}
			if i < len(input) {
				tokens = append(tokens, Token{
					Type:   TokenTypeString,
					Value:  input[start:i],
					Line:   line,
					Column: startCol,
				})
				i++ // Skip closing quote
				column++
			}

		case input[i] == ';':
			tokens = append(tokens, Token{
				Type:   TokenTypeSemicolon,
				Value:  ";",
				Line:   line,
				Column: column,
			})
			i++
			column++

		default:
			// Skip invalid characters
			i++
			column++
		}
	}

	return tokens
}

func (t TokenType) String() string {
	switch t {
	case TokenTypeIdentifier:
		return "IDENTIFIER"
	case TokenTypeString:
		return "STRING"
	case TokenTypeMultilineString:
		return "MULTILINE_STRING"
	case TokenTypeSemicolon:
		return "SEMICOLON"
	case TokenTypeComment:
		return "COMMENT"
	case TokenTypeLeftBrace:
		return "LEFT_BRACE"
	case TokenTypeRightBrace:
		return "RIGHT_BRACE"
	case TokenTypeLeftBracket:
		return "LEFT_BRACKET"
	case TokenTypeRightBracket:
		return "RIGHT_BRACKET"
	case TokenTypeLeftParen:
		return "LEFT_PAREN"
	case TokenTypeRightParen:
		return "RIGHT_PAREN"
	case TokenTypeColon:
		return "COLON"
	case TokenTypeComma:
		return "COMMA"
	case TokenTypeDot:
		return "DOT"
	case TokenTypeEquals:
		return "EQUALS"
	case TokenTypeArrow:
		return "ARROW"
	case TokenTypeDoubleEquals:
		return "DOUBLE_EQUALS"
	case TokenTypeNotEquals:
		return "NOT_EQUALS"
	case TokenTypeLess:
		return "LESS"
	case TokenTypeGreater:
		return "GREATER"
	case TokenTypeLessEquals:
		return "LESS_EQUALS"
	case TokenTypeGreaterEquals:
		return "GREATER_EQUALS"
	case TokenTypeDollar:
		return "DOLLAR"
	case TokenTypeNumber:
		return "NUMBER"
	case TokenTypeBoolean:
		return "BOOLEAN"
	default:
		return "UNKNOWN"
	}
}
