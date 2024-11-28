package validator

import (
	"fmt"
	"strings"

	"github.com/shellkjell/langspace/pkg/ast"
)

// Package validator provides entity validation functionality for LangSpace.
// It ensures that entities conform to their type-specific rules and constraints,
// providing detailed error messages for validation failures.

// Validator performs entity validation according to LangSpace's type system rules.
// It can be extended with custom validation rules and error formatting.
type Validator struct {
	// Add fields here if needed for future extensions, such as:
	// - Custom validation rules
	// - Error message templates
	// - Validation context
}

// New creates a new Validator instance configured with default validation rules.
// In the future, this constructor could accept options for customizing validation
// behavior.
//
// Returns:
//   - *Validator: A new validator instance ready to validate entities
func New() *Validator {
	return &Validator{}
}

// ValidateEntity validates an entity according to its type-specific rules.
// This is the main entry point for entity validation.
//
// Parameters:
//   - entity: The entity to validate
//
// Returns:
//   - error: Detailed validation error if the entity is invalid
//
// Validation includes:
//   - Entity type verification
//   - Required property checks
//   - Property value validation
//   - Cross-property consistency checks
func (v *Validator) ValidateEntity(entity ast.Entity) error {
	if entity == nil {
		return fmt.Errorf("entity cannot be nil")
	}

	switch entity.Type() {
	case "file":
		return v.validateFileEntity(entity)
	case "agent":
		return v.validateAgentEntity(entity)
	default:
		return fmt.Errorf("unknown entity type: %s", entity.Type())
	}
}

// validateFileEntity performs file-specific validation rules.
// File entities must have exactly two properties:
//   1. A non-empty path
//   2. A property type of either "path" or "contents"
//
// Parameters:
//   - entity: The file entity to validate
//
// Returns:
//   - error: Detailed validation error if the file entity is invalid
func (v *Validator) validateFileEntity(entity ast.Entity) error {
	props := entity.Properties()
	if len(props) != 2 {
		return fmt.Errorf("file entity must have exactly 2 properties (path and property), got %d", len(props))
	}

	path := props[0]
	property := props[1]

	if path == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	if property != "path" && property != "contents" {
		return fmt.Errorf("invalid file property: %s (must be 'path' or 'contents')", property)
	}

	return nil
}

// validateAgentEntity validates an agent entity
func (v *Validator) validateAgentEntity(entity ast.Entity) error {
	props := entity.Properties()
	if len(props) != 2 {
		return fmt.Errorf("agent entity must have exactly 2 properties (name and property), got %d", len(props))
	}

	name := props[0]
	property := props[1]

	if name == "" {
		return fmt.Errorf("agent name cannot be empty")
	}

	if !strings.HasPrefix(property, "check(") || !strings.HasSuffix(property, ")") {
		return fmt.Errorf("invalid agent property format: %s (must be 'check(filename)')", property)
	}

	return nil
}
