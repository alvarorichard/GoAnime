// Package util provides shared HTTP client with connection pooling and caching
package util

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"
)

// SharedHTTPClient is a global HTTP client with connection pooling
// optimized for high-performance concurrent requests
var (
	sharedClient     *http.Client
	sharedClientOnce sync.Once

	// FastClient is optimized for quick API requests with shorter timeouts
	fastClient     *http.Client
	fastClientOnce sync.Once
)

// httpClientConfig holds configuration for creating optimized HTTP clients
type httpClientConfig struct {
	timeout             time.Duration
	maxIdleConns        int
	maxIdleConnsPerHost int
	maxConnsPerHost     int
	idleConnTimeout     time.Duration
	tlsHandshakeTimeout time.Duration
	expectContinue      time.Duration
	keepAlive           time.Duration
	dialTimeout         time.Duration
}

// defaultConfig returns optimized default configuration
func defaultConfig() httpClientConfig {
	return httpClientConfig{
		timeout:             30 * time.Second,
		maxIdleConns:        200, // Increased for more parallel requests
		maxIdleConnsPerHost: 20,  // Doubled for better concurrency
		maxConnsPerHost:     50,  // Increased significantly
		idleConnTimeout:     120 * time.Second,
		tlsHandshakeTimeout: 5 * time.Second,
		expectContinue:      1 * time.Second,
		keepAlive:           30 * time.Second,
		dialTimeout:         5 * time.Second,
	}
}

// fastConfig returns configuration optimized for quick requests
func fastConfig() httpClientConfig {
	return httpClientConfig{
		timeout:             15 * time.Second,
		maxIdleConns:        150, // Increased for parallel scraper requests
		maxIdleConnsPerHost: 25,  // More connections per host
		maxConnsPerHost:     40,  // Allow more concurrent connections
		idleConnTimeout:     90 * time.Second,
		tlsHandshakeTimeout: 5 * time.Second,
		expectContinue:      500 * time.Millisecond,
		keepAlive:           30 * time.Second,
		dialTimeout:         5 * time.Second,
	}
}

// createTransport creates an optimized HTTP transport with the given config
func createTransport(cfg httpClientConfig) *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   cfg.dialTimeout,
			KeepAlive: cfg.keepAlive,
		}).DialContext,
		MaxIdleConns:          cfg.maxIdleConns,
		MaxIdleConnsPerHost:   cfg.maxIdleConnsPerHost,
		MaxConnsPerHost:       cfg.maxConnsPerHost,
		IdleConnTimeout:       cfg.idleConnTimeout,
		TLSHandshakeTimeout:   cfg.tlsHandshakeTimeout,
		ExpectContinueTimeout: cfg.expectContinue,
		ForceAttemptHTTP2:     true,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
}

// GetSharedClient returns the shared HTTP client with connection pooling.
// This client is optimized for general use with reasonable timeouts.
func GetSharedClient() *http.Client {
	sharedClientOnce.Do(func() {
		cfg := defaultConfig()
		sharedClient = &http.Client{
			Transport: createTransport(cfg),
			Timeout:   cfg.timeout,
		}
	})
	return sharedClient
}

// GetFastClient returns an HTTP client optimized for quick API requests.
// Use this for lightweight API calls where speed is critical.
func GetFastClient() *http.Client {
	fastClientOnce.Do(func() {
		cfg := fastConfig()
		fastClient = &http.Client{
			Transport: createTransport(cfg),
			Timeout:   cfg.timeout,
		}
	})
	return fastClient
}

// ResponseCache provides a simple in-memory cache for API responses
type ResponseCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	maxAge  time.Duration
	maxSize int
}

type cacheEntry struct {
	data      []byte
	timestamp time.Time
}

// NewResponseCache creates a new response cache with the specified max age and size
func NewResponseCache(maxAge time.Duration, maxSize int) *ResponseCache {
	cache := &ResponseCache{
		entries: make(map[string]*cacheEntry, maxSize),
		maxAge:  maxAge,
		maxSize: maxSize,
	}
	// Start background cleanup goroutine
	go cache.cleanupLoop()
	return cache
}

// Get retrieves a cached response if it exists and is not expired
func (c *ResponseCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Since(entry.timestamp) > c.maxAge {
		return nil, false
	}

	return entry.data, true
}

// Set stores a response in the cache
func (c *ResponseCache) Set(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple eviction: if at max size, remove oldest entry
	if len(c.entries) >= c.maxSize {
		var oldestKey string
		var oldestTime time.Time
		first := true
		for k, v := range c.entries {
			if first || v.timestamp.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.timestamp
				first = false
			}
		}
		if oldestKey != "" {
			delete(c.entries, oldestKey)
		}
	}

	c.entries[key] = &cacheEntry{
		data:      data,
		timestamp: time.Now(),
	}
}

// cleanupLoop periodically removes expired entries
func (c *ResponseCache) cleanupLoop() {
	ticker := time.NewTicker(c.maxAge / 2)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes expired entries
func (c *ResponseCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.Sub(entry.timestamp) > c.maxAge {
			delete(c.entries, key)
		}
	}
}

// Global caches for different API responses
var (
	// AniListCache caches AniList API responses (5 minute TTL)
	AniListCache     *ResponseCache
	aniListCacheOnce sync.Once

	// SearchCache caches search results (2 minute TTL)
	SearchCache     *ResponseCache
	searchCacheOnce sync.Once
)

// GetAniListCache returns the global AniList cache
func GetAniListCache() *ResponseCache {
	aniListCacheOnce.Do(func() {
		AniListCache = NewResponseCache(5*time.Minute, 200) // Increased cache size
	})
	return AniListCache
}

// GetSearchCache returns the global search cache
func GetSearchCache() *ResponseCache {
	searchCacheOnce.Do(func() {
		SearchCache = NewResponseCache(2*time.Minute, 100) // Increased cache size
	})
	return SearchCache
}

// WorkerPool provides a safe way to run multiple goroutines with a limit
type WorkerPool struct {
	maxWorkers int
	semaphore  chan struct{}
}

// NewWorkerPool creates a new worker pool with the specified max concurrent workers
func NewWorkerPool(maxWorkers int) *WorkerPool {
	return &WorkerPool{
		maxWorkers: maxWorkers,
		semaphore:  make(chan struct{}, maxWorkers),
	}
}

// Submit submits a task to the worker pool
// It will block if all workers are busy until one becomes available
func (wp *WorkerPool) Submit(task func()) {
	wp.semaphore <- struct{}{} // Acquire
	go func() {
		defer func() { <-wp.semaphore }() // Release
		task()
	}()
}

// Wait waits for all submitted tasks to complete
func (wp *WorkerPool) Wait() {
	// Fill the semaphore to ensure all workers are done
	for i := 0; i < wp.maxWorkers; i++ {
		wp.semaphore <- struct{}{}
	}
	// Release them
	for i := 0; i < wp.maxWorkers; i++ {
		<-wp.semaphore
	}
}

// Global worker pools for different use cases
var (
	// ScraperPool is used for concurrent scraper operations
	ScraperPool     *WorkerPool
	scraperPoolOnce sync.Once

	// APIPool is used for concurrent API requests
	APIPool     *WorkerPool
	apiPoolOnce sync.Once
)

// GetScraperPool returns the global scraper worker pool (10 workers)
func GetScraperPool() *WorkerPool {
	scraperPoolOnce.Do(func() {
		ScraperPool = NewWorkerPool(10)
	})
	return ScraperPool
}

// GetAPIPool returns the global API worker pool (15 workers)
func GetAPIPool() *WorkerPool {
	apiPoolOnce.Do(func() {
		APIPool = NewWorkerPool(15)
	})
	return APIPool
}

// ParallelExecute executes multiple functions in parallel with a worker limit
// Returns when all functions complete. Safe for concurrent use.
func ParallelExecute(maxWorkers int, tasks ...func()) {
	if len(tasks) == 0 {
		return
	}

	// Use min of maxWorkers and task count
	workers := maxWorkers
	if len(tasks) < workers {
		workers = len(tasks)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, workers)

	for _, task := range tasks {
		wg.Add(1)
		task := task // Capture for goroutine
		go func() {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release
			task()
		}()
	}

	wg.Wait()
}
