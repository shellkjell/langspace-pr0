package pool

import (
	"sync"
)

// TokenPool provides a pool of reusable token objects
type TokenPool struct {
	pool sync.Pool
}

// NewTokenPool creates a new TokenPool instance
func NewTokenPool() *TokenPool {
	return &TokenPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 64) // Default initial capacity
			},
		},
	}
}

// Get retrieves a byte slice from the pool
func (p *TokenPool) Get() []byte {
	return p.pool.Get().([]byte)
}

// Put returns a byte slice to the pool
func (p *TokenPool) Put(b []byte) {
	b = b[:0] // Clear the slice but keep capacity
	p.pool.Put(b)
}

// GetString retrieves a string from the pool
func (p *TokenPool) GetString() string {
	return string(p.Get())
}

// PutString returns a string to the pool
func (p *TokenPool) PutString(s string) {
	p.Put([]byte(s))
}
