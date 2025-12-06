// Package slices provides generic utility functions for slice operations.
// These functions reduce boilerplate and improve type safety throughout the codebase.
package slices

// Filter returns a new slice containing only elements that satisfy the predicate.
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Map transforms each element in a slice using the provided function.
func Map[T, U any](slice []T, transform func(T) U) []U {
	result := make([]U, len(slice))
	for i, item := range slice {
		result[i] = transform(item)
	}
	return result
}

// Find returns the first element that satisfies the predicate, along with whether it was found.
func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, item := range slice {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// FindIndex returns the index of the first element that satisfies the predicate, or -1 if not found.
func FindIndex[T any](slice []T, predicate func(T) bool) int {
	for i, item := range slice {
		if predicate(item) {
			return i
		}
	}
	return -1
}

// Any returns true if any element satisfies the predicate.
func Any[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if predicate(item) {
			return true
		}
	}
	return false
}

// All returns true if all elements satisfy the predicate.
// Returns true for empty slices.
func All[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// Count returns the number of elements that satisfy the predicate.
func Count[T any](slice []T, predicate func(T) bool) int {
	count := 0
	for _, item := range slice {
		if predicate(item) {
			count++
		}
	}
	return count
}

// Remove returns a new slice with the first occurrence of an element removed.
// Comparison uses the provided equals function.
func Remove[T any](slice []T, equals func(T) bool) []T {
	for i, item := range slice {
		if equals(item) {
			return append(slice[:i:i], slice[i+1:]...)
		}
	}
	return slice
}

// Contains returns true if the slice contains an element satisfying the predicate.
func Contains[T any](slice []T, predicate func(T) bool) bool {
	return Any(slice, predicate)
}

// Unique returns a new slice with duplicate elements removed.
// Uses the provided key function to determine uniqueness.
func Unique[T any, K comparable](slice []T, key func(T) K) []T {
	seen := make(map[K]struct{})
	var result []T
	for _, item := range slice {
		k := key(item)
		if _, exists := seen[k]; !exists {
			seen[k] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// GroupBy groups elements by a key function.
func GroupBy[T any, K comparable](slice []T, key func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range slice {
		k := key(item)
		result[k] = append(result[k], item)
	}
	return result
}

// Partition splits a slice into two slices based on a predicate.
// The first slice contains elements that satisfy the predicate,
// the second contains elements that don't.
func Partition[T any](slice []T, predicate func(T) bool) (matching []T, notMatching []T) {
	for _, item := range slice {
		if predicate(item) {
			matching = append(matching, item)
		} else {
			notMatching = append(notMatching, item)
		}
	}
	return
}
