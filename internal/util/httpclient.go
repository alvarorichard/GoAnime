// Package util provides shared HTTP client with connection pooling and caching
package util

import (
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/enetx/surf"
)

// SharedHTTPClient is a global HTTP client with Chrome TLS fingerprint
// via enetx/surf for anti-bot bypass and connection pooling
var (
	sharedClient     *http.Client
	sharedClientOnce sync.Once

	// FastClient is optimized for quick API requests with shorter timeouts
	fastClient     *http.Client
	fastClientOnce sync.Once

	// downloadClient is optimized for large file downloads with long timeout
	downloadClient     *http.Client
	downloadClientOnce sync.Once
)

// newSurfStdClient creates a *net/http.Client backed by surf with Chrome
// browser impersonation and the given timeout.
func newSurfStdClient(timeout time.Duration) *http.Client {
	client := surf.NewClient().
		Builder().
		Impersonate().Chrome().
		Timeout(timeout).
		Build().
		Unwrap().
		Std()
	return client
}

// GetSharedClient returns the shared HTTP client with Chrome TLS fingerprint.
// This client is optimized for general use with reasonable timeouts.
func GetSharedClient() *http.Client {
	sharedClientOnce.Do(func() {
		sharedClient = newSurfStdClient(20 * time.Second)
	})
	return sharedClient
}

// GetFastClient returns an HTTP client optimized for quick API requests.
// Uses Chrome TLS fingerprint for anti-bot bypass.
func GetFastClient() *http.Client {
	fastClientOnce.Do(func() {
		fastClient = newSurfStdClient(8 * time.Second)
	})
	return fastClient
}

// NewFastClient creates a NEW fast HTTP client with its own connection pool.
// Use this instead of GetFastClient when the caller will be used concurrently
// with other clients (e.g., scrapers running in parallel goroutines) to avoid
// data races in the underlying http2 transport.
func NewFastClient() *http.Client {
	return newSurfStdClient(8 * time.Second)
}

// GetDownloadClient returns an HTTP client optimized for large file downloads.
// Uses Chrome TLS fingerprint for anti-bot bypass with a 5-minute timeout.
func GetDownloadClient() *http.Client {
	downloadClientOnce.Do(func() {
		downloadClient = newSurfStdClient(5 * time.Minute)
	})
	return downloadClient
}

// PreWarmClients triggers background initialization of the shared surf HTTP
// clients so that the first real request doesn't pay the Chrome TLS setup cost.
// Call this as early as possible (e.g., in main before user input).
func PreWarmClients() {
	go func() { GetFastClient() }()
	go func() { GetSharedClient() }()
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
		AniListCache = NewResponseCache(10*time.Minute, 500)
	})
	return AniListCache
}

// GetSearchCache returns the global search cache
func GetSearchCache() *ResponseCache {
	searchCacheOnce.Do(func() {
		SearchCache = NewResponseCache(3*time.Minute, 300)
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
	workers := min(len(tasks), maxWorkers)

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

// knownHosts are API hosts that we know will be contacted, used for connection pre-warming.
var knownHosts = []string{
	"graphql.anilist.co:443",
	"api.jikan.moe:443",
	"allanime.day:443",
	"animefire.io:443",
	"animesdrive.online:443",
	"flixhq.to:443",
	"sflix.to:443",
	"9animetv.to:443",
	"kitsu.io:443",
}

var preWarmOnce sync.Once

// PreWarmConnections initiates background DNS resolution and TLS handshakes
// for known API hosts. Call this early (e.g., at startup) so that by the time
// the first real request is made, connections are already pooled.
func PreWarmConnections() {
	preWarmOnce.Do(func() {
		client := GetFastClient()
		for _, host := range knownHosts {
			go func() {
				// GET request with short timeout — triggers full TCP+TLS handshake
				// to populate the connection pool. HEAD may not establish full TLS.
				req, err := http.NewRequest("GET", "https://"+host, nil)
				if err != nil {
					return
				}
				req.Header.Set("User-Agent", "GoAnime/1.0")
				resp, err := client.Do(req) // #nosec G107
				if err != nil {
					// DNS or connect failure is fine — this is best-effort
					Debugf("Pre-warm %s: %v", host, err)
					return
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				Debugf("Pre-warmed connection to %s", host)
			}()
		}
	})
}
