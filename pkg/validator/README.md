# Validator Package

The `validator` package ensures that LangSpace entities conform to their type-specific rules and constraints. It provides detailed error messages and extensible validation logic.

## Overview

The validator performs several types of validation:
- Entity type verification
- Required property checks
- Property value validation
- Cross-property consistency checks

## Usage

```go
import "github.com/shellkjell/langspace/pkg/validator"

// Create a validator
v := validator.New()

// Validate an entity
err := v.ValidateEntity(entity)
if err != nil {
    log.Fatalf("Validation error: %v", err)
}
```

## Validation Rules

### File Entities
- Must have exactly two properties
- Path property must be non-empty
- Property type must be either "path" or "contents"

### Agent Entities
- Must have exactly two properties
- Name must be non-empty
- Instruction must be a valid command format

## Error Messages

The validator provides detailed error messages that include:
- The specific validation rule that failed
- The entity type and properties involved
- Suggestions for fixing the validation error

## Extension Points

The validator can be extended in several ways:
1. Add custom validation rules
2. Customize error message templates
3. Add validation context for complex rules
4. Implement custom property validators

## Best Practices

- Validate entities immediately after creation
- Handle validation errors appropriately
- Use validation results for error reporting
- Consider adding custom validators for domain-specific rules

## Future Enhancements

Planned improvements include:
- Async validation support
- Custom validation rule registration
- Validation result caching
- Enhanced error reporting
