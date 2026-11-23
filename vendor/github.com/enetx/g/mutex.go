package g

import "sync"

// Mutex is a mutual exclusion lock that protects a value of type T.
// Unlike sync.Mutex, it binds the protected data to the lock itself,
// making it impossible to access the data without holding the lock.
//
// A Mutex must not be copied after first use: it embeds a sync.Mutex and a
// copy would protect a different value than the original (go vet's copylocks
// analyzer flags such copies). Always pass a *Mutex, never a Mutex by value.
type Mutex[T any] struct {
	mu  sync.Mutex
	val T
}

// MutexGuard provides access to the value protected by a Mutex.
// The guard must be explicitly unlocked when done.
//
// A guard holds pointers into its owning Mutex; copying the guard and calling
// Unlock on more than one copy unlocks the same underlying lock twice, which
// panics. Use each guard exactly once and do not copy it.
type MutexGuard[T any] struct {
	mu  *sync.Mutex
	val *T
}

// NewMutex creates a new Mutex containing the given value.
func NewMutex[T any](value T) *Mutex[T] { return &Mutex[T]{val: value} }

// Lock acquires the mutex and returns a guard that provides access to the protected value.
// The caller must call Unlock on the guard when done (typically via defer).
func (m *Mutex[T]) Lock() MutexGuard[T] {
	m.mu.Lock()
	return MutexGuard[T]{mu: &m.mu, val: &m.val}
}

// With acquires the mutex, calls fn with a pointer to the protected value,
// and releases the mutex when fn returns.
// This is a convenience method that eliminates the need for manual Lock/Unlock management.
func (m *Mutex[T]) With(fn func(*T)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fn(&m.val)
}

// TryLock attempts to acquire the mutex without blocking.
// Returns Some(guard) if successful, None if the mutex is already locked.
func (m *Mutex[T]) TryLock() Option[MutexGuard[T]] {
	if m.mu.TryLock() {
		return Some(MutexGuard[T]{mu: &m.mu, val: &m.val})
	}

	return None[MutexGuard[T]]()
}

// Get returns a copy of the protected value.
func (g MutexGuard[T]) Get() T { return *g.val }

// Set replaces the protected value with a new one.
func (g MutexGuard[T]) Set(value T) { *g.val = value }

// Deref returns a pointer to the protected value for direct manipulation.
func (g MutexGuard[T]) Deref() *T { return g.val }

// Unlock releases the mutex. Must be called when done with the guard.
func (g MutexGuard[T]) Unlock() { g.mu.Unlock() }
