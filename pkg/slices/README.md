# Slices Package

The `slices` package provides generic utility functions for slice operations in Go 1.18+. These functions reduce boilerplate and improve type safety throughout the codebase.

## Overview

This package implements common functional programming patterns for slice manipulation:
- Filtering and mapping
- Searching and finding
- Counting and grouping
- Partitioning and deduplication

## Usage

```go
import "github.com/shellkjell/langspace/pkg/slices"

// Filter elements
evens := slices.Filter(numbers, func(n int) bool { return n%2 == 0 })

// Map/transform elements
doubled := slices.Map(numbers, func(n int) int { return n * 2 })

// Find an element
val, found := slices.Find(users, func(u User) bool { return u.ID == 42 })

// Count matching elements
count := slices.Count(items, func(i Item) bool { return i.Active })

// Group by key
grouped := slices.GroupBy(users, func(u User) string { return u.Role })
```

## Functions

### Filtering

| Function | Description |
|----------|-------------|
| `Filter[T](slice, predicate)` | Returns elements matching predicate |
| `Partition[T](slice, predicate)` | Splits into matching/non-matching slices |
| `Unique[T, K](slice, key)` | Removes duplicates by key |
| `Remove[T](slice, equals)` | Removes first matching element |

### Transformation

| Function | Description |
|----------|-------------|
| `Map[T, U](slice, transform)` | Transforms each element |
| `GroupBy[T, K](slice, key)` | Groups elements by key |

### Searching

| Function | Description |
|----------|-------------|
| `Find[T](slice, predicate)` | Finds first matching element |
| `FindIndex[T](slice, predicate)` | Finds index of first match |
| `Any[T](slice, predicate)` | Checks if any element matches |
| `All[T](slice, predicate)` | Checks if all elements match |
| `Contains[T](slice, predicate)` | Alias for Any |

### Counting

| Function | Description |
|----------|-------------|
| `Count[T](slice, predicate)` | Counts matching elements |

## Design Decisions

### Why a Custom Package?

While Go's `slices` standard library package (added in 1.21) provides some utilities, it focuses on basic operations like sorting and comparing. This package provides higher-level functional programming primitives that are common in LangSpace:

- Entity filtering by type
- Relationship searching
- Entity grouping and counting

### Type Safety

All functions use Go generics to provide compile-time type safety. This eliminates runtime type assertions and makes code more maintainable.

### Performance

Functions are optimized for common use cases:
- Single-pass iteration where possible
- Minimal allocations
- No reflection

## Examples

### Filter Entities by Type

```go
agents := slices.Filter(entities, func(e ast.Entity) bool {
    return e.Type() == "agent"
})
```

### Find Entity by Name

```go
entity, found := slices.Find(entities, func(e ast.Entity) bool {
    return e.Name() == "my-agent"
})
```

### Group Entities by Type

```go
byType := slices.GroupBy(entities, func(e ast.Entity) string {
    return e.Type()
})
// byType["file"], byType["agent"], etc.
```

### Count Relationships

```go
assignedCount := slices.Count(relationships, func(r Relationship) bool {
    return r.Type == RelationTypeAssigned
})
```
