package pool

import (
	"sync"
	"testing"
)

func TestTokenPool_GetPut(t *testing.T) {
	pool := NewTokenPool()

	// Test Get returns a slice with expected capacity
	slice := pool.Get()
	if cap(slice) != 64 {
		t.Errorf("TokenPool.Get() returned slice with capacity %d, want 64", cap(slice))
	}

	// Test Put clears the slice
	slice = append(slice, []byte("test")...)
	pool.Put(slice)
	slice = pool.Get()
	if len(slice) != 0 {
		t.Errorf("TokenPool.Get() after Put() returned slice with length %d, want 0", len(slice))
	}
}

func TestTokenPool_GetPutString(t *testing.T) {
	pool := NewTokenPool()

	// Test GetString returns an empty string
	str := pool.GetString()
	if str != "" {
		t.Errorf("TokenPool.GetString() = %q, want empty string", str)
	}

	// Test PutString
	pool.PutString("test")
	str = pool.GetString()
	if str != "" {
		t.Errorf("TokenPool.GetString() after PutString() = %q, want empty string", str)
	}
}

func TestTokenPool_Concurrency(t *testing.T) {
	pool := NewTokenPool()
	var wg sync.WaitGroup
	numGoroutines := 100

	// Test concurrent Get/Put operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			slice := pool.Get()
			slice = append(slice, []byte("test")...)
			pool.Put(slice)
		}()
	}
	wg.Wait()

	// Test concurrent GetString/PutString operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			pool.PutString("test")
			_ = pool.GetString()
		}()
	}
	wg.Wait()
}

func BenchmarkTokenPool_GetPut(b *testing.B) {
	pool := NewTokenPool()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		slice := pool.Get()
		pool.Put(slice)
	}
}

func BenchmarkTokenPool_GetPutString(b *testing.B) {
	pool := NewTokenPool()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool.PutString("test")
		_ = pool.GetString()
	}
}

func BenchmarkTokenPool_Parallel(b *testing.B) {
	pool := NewTokenPool()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slice := pool.Get()
			pool.Put(slice)
		}
	})
}
