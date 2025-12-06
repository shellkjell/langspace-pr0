package slices

import (
	"reflect"
	"testing"
)

func TestFilter(t *testing.T) {
	tests := []struct {
		name      string
		slice     []int
		predicate func(int) bool
		want      []int
	}{
		{
			name:      "filter even numbers",
			slice:     []int{1, 2, 3, 4, 5, 6},
			predicate: func(n int) bool { return n%2 == 0 },
			want:      []int{2, 4, 6},
		},
		{
			name:      "filter empty slice",
			slice:     []int{},
			predicate: func(n int) bool { return true },
			want:      nil,
		},
		{
			name:      "filter none match",
			slice:     []int{1, 3, 5},
			predicate: func(n int) bool { return n%2 == 0 },
			want:      nil,
		},
		{
			name:      "filter all match",
			slice:     []int{2, 4, 6},
			predicate: func(n int) bool { return n%2 == 0 },
			want:      []int{2, 4, 6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Filter(tt.slice, tt.predicate)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMap(t *testing.T) {
	tests := []struct {
		name      string
		slice     []int
		transform func(int) string
		want      []string
	}{
		{
			name:      "int to string",
			slice:     []int{1, 2, 3},
			transform: func(n int) string { return string(rune('0' + n)) },
			want:      []string{"1", "2", "3"},
		},
		{
			name:      "empty slice",
			slice:     []int{},
			transform: func(n int) string { return "" },
			want:      []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Map(tt.slice, tt.transform)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFind(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		got, found := Find(slice, func(n int) bool { return n == 3 })
		if !found || got != 3 {
			t.Errorf("Find() = %v, %v, want 3, true", got, found)
		}
	})

	t.Run("not found", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		got, found := Find(slice, func(n int) bool { return n == 10 })
		if found {
			t.Errorf("Find() = %v, %v, want 0, false", got, found)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		slice := []int{}
		_, found := Find(slice, func(n int) bool { return true })
		if found {
			t.Error("Find() on empty slice should return false")
		}
	})
}

func TestFindIndex(t *testing.T) {
	tests := []struct {
		name      string
		slice     []int
		predicate func(int) bool
		want      int
	}{
		{
			name:      "found at beginning",
			slice:     []int{1, 2, 3},
			predicate: func(n int) bool { return n == 1 },
			want:      0,
		},
		{
			name:      "found in middle",
			slice:     []int{1, 2, 3},
			predicate: func(n int) bool { return n == 2 },
			want:      1,
		},
		{
			name:      "not found",
			slice:     []int{1, 2, 3},
			predicate: func(n int) bool { return n == 10 },
			want:      -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindIndex(tt.slice, tt.predicate)
			if got != tt.want {
				t.Errorf("FindIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAny(t *testing.T) {
	tests := []struct {
		name      string
		slice     []int
		predicate func(int) bool
		want      bool
	}{
		{
			name:      "some match",
			slice:     []int{1, 2, 3},
			predicate: func(n int) bool { return n == 2 },
			want:      true,
		},
		{
			name:      "none match",
			slice:     []int{1, 3, 5},
			predicate: func(n int) bool { return n%2 == 0 },
			want:      false,
		},
		{
			name:      "empty slice",
			slice:     []int{},
			predicate: func(n int) bool { return true },
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Any(tt.slice, tt.predicate)
			if got != tt.want {
				t.Errorf("Any() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAll(t *testing.T) {
	tests := []struct {
		name      string
		slice     []int
		predicate func(int) bool
		want      bool
	}{
		{
			name:      "all match",
			slice:     []int{2, 4, 6},
			predicate: func(n int) bool { return n%2 == 0 },
			want:      true,
		},
		{
			name:      "some don't match",
			slice:     []int{2, 3, 6},
			predicate: func(n int) bool { return n%2 == 0 },
			want:      false,
		},
		{
			name:      "empty slice",
			slice:     []int{},
			predicate: func(n int) bool { return false },
			want:      true, // vacuously true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := All(tt.slice, tt.predicate)
			if got != tt.want {
				t.Errorf("All() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name      string
		slice     []int
		predicate func(int) bool
		want      int
	}{
		{
			name:      "count evens",
			slice:     []int{1, 2, 3, 4, 5, 6},
			predicate: func(n int) bool { return n%2 == 0 },
			want:      3,
		},
		{
			name:      "count none",
			slice:     []int{1, 3, 5},
			predicate: func(n int) bool { return n%2 == 0 },
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Count(tt.slice, tt.predicate)
			if got != tt.want {
				t.Errorf("Count() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name   string
		slice  []int
		equals func(int) bool
		want   []int
	}{
		{
			name:   "remove middle",
			slice:  []int{1, 2, 3},
			equals: func(n int) bool { return n == 2 },
			want:   []int{1, 3},
		},
		{
			name:   "remove first",
			slice:  []int{1, 2, 3},
			equals: func(n int) bool { return n == 1 },
			want:   []int{2, 3},
		},
		{
			name:   "remove last",
			slice:  []int{1, 2, 3},
			equals: func(n int) bool { return n == 3 },
			want:   []int{1, 2},
		},
		{
			name:   "not found",
			slice:  []int{1, 2, 3},
			equals: func(n int) bool { return n == 10 },
			want:   []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Remove(tt.slice, tt.equals)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		want  []int
	}{
		{
			name:  "with duplicates",
			slice: []int{1, 2, 2, 3, 3, 3},
			want:  []int{1, 2, 3},
		},
		{
			name:  "no duplicates",
			slice: []int{1, 2, 3},
			want:  []int{1, 2, 3},
		},
		{
			name:  "empty slice",
			slice: []int{},
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Unique(tt.slice, func(n int) int { return n })
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroupBy(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6}
	got := GroupBy(slice, func(n int) string {
		if n%2 == 0 {
			return "even"
		}
		return "odd"
	})

	if !reflect.DeepEqual(got["even"], []int{2, 4, 6}) {
		t.Errorf("GroupBy() even = %v, want [2, 4, 6]", got["even"])
	}
	if !reflect.DeepEqual(got["odd"], []int{1, 3, 5}) {
		t.Errorf("GroupBy() odd = %v, want [1, 3, 5]", got["odd"])
	}
}

func TestPartition(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6}
	matching, notMatching := Partition(slice, func(n int) bool { return n%2 == 0 })

	if !reflect.DeepEqual(matching, []int{2, 4, 6}) {
		t.Errorf("Partition() matching = %v, want [2, 4, 6]", matching)
	}
	if !reflect.DeepEqual(notMatching, []int{1, 3, 5}) {
		t.Errorf("Partition() notMatching = %v, want [1, 3, 5]", notMatching)
	}
}

// Benchmark tests
func BenchmarkFilter(b *testing.B) {
	slice := make([]int, 1000)
	for i := range slice {
		slice[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Filter(slice, func(n int) bool { return n%2 == 0 })
	}
}

func BenchmarkMap(b *testing.B) {
	slice := make([]int, 1000)
	for i := range slice {
		slice[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Map(slice, func(n int) int { return n * 2 })
	}
}

func BenchmarkFind(b *testing.B) {
	slice := make([]int, 1000)
	for i := range slice {
		slice[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Find(slice, func(n int) bool { return n == 500 })
	}
}
