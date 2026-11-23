package g

import "sync"

// RwLock is a reader-writer lock that protects a value of type T.
// It allows multiple readers or a single writer at any point in time.
// Unlike sync.RWMutex, it binds the protected data to the lock itself,
// making it impossible to access the data without holding the lock.
//
// An RwLock must not be copied after first use: it embeds a sync.RWMutex and a
// copy would protect a different value than the original (go vet's copylocks
// analyzer flags such copies). Always pass an *RwLock, never an RwLock by value.
type RwLock[T any] struct {
	mu  sync.RWMutex
	val T
}

// RwLockReadGuard provides read-only access to the value protected by an RwLock.
// Multiple read guards can exist simultaneously.
//
// A guard holds pointers into its owning RwLock; copying the guard and calling
// Unlock on more than one copy releases the same read lock twice, which panics.
// Use each guard exactly once and do not copy it.
type RwLockReadGuard[T any] struct {
	mu  *sync.RWMutex
	val *T
}

// RwLockWriteGuard provides exclusive read-write access to the value protected by an RwLock.
// Only one write guard can exist at a time, and no read guards can coexist with it.
//
// A guard holds pointers into its owning RwLock; copying the guard and calling
// Unlock on more than one copy releases the same write lock twice, which panics.
// Use each guard exactly once and do not copy it.
type RwLockWriteGuard[T any] struct {
	mu  *sync.RWMutex
	val *T
}

// NewRwLock creates a new RwLock containing the given value.
func NewRwLock[T any](value T) *RwLock[T] { return &RwLock[T]{val: value} }

// RWith acquires a read lock, calls fn with a copy of the protected value,
// and releases the lock when fn returns.
// The value is passed by copy to prevent accidental mutation under a read lock.
func (r *RwLock[T]) RWith(fn func(T)) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn(r.val)
}

// With acquires a write lock, calls fn with a pointer to the protected value,
// and releases the lock when fn returns.
// This is a convenience method that eliminates the need for manual Lock/Unlock management.
func (r *RwLock[T]) With(fn func(*T)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fn(&r.val)
}

// Read acquires a read lock and returns a guard that provides read-only access.
// Multiple goroutines can hold read locks simultaneously.
// The caller must call Unlock on the guard when done (typically via defer).
func (r *RwLock[T]) Read() RwLockReadGuard[T] {
	r.mu.RLock()
	return RwLockReadGuard[T]{mu: &r.mu, val: &r.val}
}

// Write acquires a write lock and returns a guard that provides exclusive access.
// No other readers or writers can access the value while this guard exists.
// The caller must call Unlock on the guard when done (typically via defer).
func (r *RwLock[T]) Write() RwLockWriteGuard[T] {
	r.mu.Lock()
	return RwLockWriteGuard[T]{mu: &r.mu, val: &r.val}
}

// TryRead attempts to acquire a read lock without blocking.
// Returns Some(guard) if successful, None if a write lock is held.
func (r *RwLock[T]) TryRead() Option[RwLockReadGuard[T]] {
	if r.mu.TryRLock() {
		return Some(RwLockReadGuard[T]{mu: &r.mu, val: &r.val})
	}

	return None[RwLockReadGuard[T]]()
}

// TryWrite attempts to acquire a write lock without blocking.
// Returns Some(guard) if successful, None if any lock is held.
func (r *RwLock[T]) TryWrite() Option[RwLockWriteGuard[T]] {
	if r.mu.TryLock() {
		return Some(RwLockWriteGuard[T]{mu: &r.mu, val: &r.val})
	}

	return None[RwLockWriteGuard[T]]()
}

// Get returns a copy of the protected value.
func (g RwLockReadGuard[T]) Get() T { return *g.val }

// Deref returns a pointer to the protected value for direct access.
// Note: modifying through this pointer would be a logic error.
func (g RwLockReadGuard[T]) Deref() *T { return g.val }

// Unlock releases the read lock. Must be called when done with the guard.
func (g RwLockReadGuard[T]) Unlock() { g.mu.RUnlock() }

// Get returns a copy of the protected value.
func (g RwLockWriteGuard[T]) Get() T { return *g.val }

// Set replaces the protected value with a new one.
func (g RwLockWriteGuard[T]) Set(value T) { *g.val = value }

// Deref returns a pointer to the protected value for direct manipulation.
func (g RwLockWriteGuard[T]) Deref() *T { return g.val }

// Unlock releases the write lock. Must be called when done with the guard.
func (g RwLockWriteGuard[T]) Unlock() { g.mu.Unlock() }
