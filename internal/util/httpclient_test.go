package util

import (
	"bytes"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseCache_DoesNotLeakMutableBuffers(t *testing.T) {
	c := NewResponseCache(time.Minute, 5)
	input := []byte("original")
	c.Set("k", input)

	input[0] = 'X'
	got, ok := c.Get("k")
	if !ok || !bytes.Equal(got, []byte("original")) {
		t.Fatalf("Set retained caller buffer: got %q, ok=%v", got, ok)
	}

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 1000 {
				value, exists := c.Get("k")
				if exists && len(value) > 0 {
					value[0]++
				}
			}
		})
	}
	wg.Wait()

	got, ok = c.Get("k")
	if !ok || !bytes.Equal(got, []byte("original")) {
		t.Fatalf("Get exposed cache buffer: got %q, ok=%v", got, ok)
	}
}

func TestNewResponseCache(t *testing.T) {
	t.Parallel()
	c := NewResponseCache(time.Minute, 5)
	require.NotNil(t, c)
	assert.Equal(t, time.Minute, c.maxAge)
	assert.Equal(t, 5, c.maxSize)
}

func TestResponseCache_GetSet(t *testing.T) {
	t.Parallel()
	c := NewResponseCache(time.Minute, 5)
	c.Set("k", []byte("v"))
	got, ok := c.Get("k")
	require.True(t, ok)
	assert.Equal(t, []byte("v"), got)
}

func TestResponseCache_Get_MissReturnsFalse(t *testing.T) {
	t.Parallel()
	c := NewResponseCache(time.Minute, 5)
	_, ok := c.Get("missing")
	assert.False(t, ok)
}

func TestResponseCache_Get_ExpiredReturnsFalse(t *testing.T) {
	t.Parallel()
	c := NewResponseCache(10*time.Millisecond, 5)
	c.Set("k", []byte("v"))
	time.Sleep(50 * time.Millisecond)
	_, ok := c.Get("k")
	assert.False(t, ok)
}

func TestResponseCache_Set_EvictsOldestAtMaxSize(t *testing.T) {
	t.Parallel()
	c := NewResponseCache(time.Minute, 2)
	c.Set("a", []byte("1"))
	time.Sleep(2 * time.Millisecond)
	c.Set("b", []byte("2"))
	time.Sleep(2 * time.Millisecond)
	c.Set("c", []byte("3"))
	_, okA := c.Get("a")
	_, okC := c.Get("c")
	assert.False(t, okA, "oldest should be evicted")
	assert.True(t, okC)
}

func TestResponseCache_Cleanup(t *testing.T) {
	t.Parallel()
	c := NewResponseCache(5*time.Millisecond, 100)
	c.Set("k", []byte("v"))
	time.Sleep(20 * time.Millisecond)
	c.cleanup()
	_, ok := c.Get("k")
	assert.False(t, ok)
}

func TestGetAniListCache_Singleton(t *testing.T) {
	t.Parallel()
	a := GetAniListCache()
	b := GetAniListCache()
	assert.Same(t, a, b)
}

func TestGetSearchCache_Singleton(t *testing.T) {
	t.Parallel()
	a := GetSearchCache()
	b := GetSearchCache()
	assert.Same(t, a, b)
}

func TestNewSurfStdClient(t *testing.T) {
	t.Parallel()
	c := newSurfStdClient(5 * time.Second)
	require.NotNil(t, c)
	assert.Equal(t, 5*time.Second, c.Timeout)
}

func TestGetSharedClient_Singleton(t *testing.T) {
	t.Parallel()
	a := GetSharedClient()
	b := GetSharedClient()
	assert.Same(t, a, b)
}

func TestGetFastClient_Singleton(t *testing.T) {
	t.Parallel()
	a := GetFastClient()
	b := GetFastClient()
	assert.Same(t, a, b)
}

func TestNewFastClient_Distinct(t *testing.T) {
	t.Parallel()
	a := NewFastClient()
	b := NewFastClient()
	require.NotNil(t, a)
	require.NotNil(t, b)
	assert.NotSame(t, a, b)
}

func TestGetDownloadClient_Singleton(t *testing.T) {
	t.Parallel()
	a := GetDownloadClient()
	b := GetDownloadClient()
	assert.Same(t, a, b)
}

func TestPreWarmClients_DoesNotPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() { PreWarmClients() })
}

func TestNewWorkerPool(t *testing.T) {
	t.Parallel()
	wp := NewWorkerPool(3)
	require.NotNil(t, wp)
	assert.Equal(t, 3, wp.maxWorkers)
}

func TestWorkerPool_SubmitAndWait(t *testing.T) {
	t.Parallel()
	wp := NewWorkerPool(2)
	var n atomic.Int32
	for range 5 {
		wp.Submit(func() { n.Add(1) })
	}
	wp.Wait()
	assert.Equal(t, int32(5), n.Load())
}

func TestGetScraperPool_Singleton(t *testing.T) {
	t.Parallel()
	a := GetScraperPool()
	b := GetScraperPool()
	assert.Same(t, a, b)
}

func TestGetAPIPool_Singleton(t *testing.T) {
	t.Parallel()
	a := GetAPIPool()
	b := GetAPIPool()
	assert.Same(t, a, b)
}

func TestParallelExecute(t *testing.T) {
	t.Parallel()

	t.Run("no tasks no-op", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() { ParallelExecute(2) })
	})

	t.Run("runs all", func(t *testing.T) {
		t.Parallel()
		var n atomic.Int32
		var wg sync.WaitGroup
		wg.Add(3)
		tasks := []func(){
			func() { defer wg.Done(); n.Add(1) },
			func() { defer wg.Done(); n.Add(1) },
			func() { defer wg.Done(); n.Add(1) },
		}
		ParallelExecute(2, tasks...)
		wg.Wait()
		assert.Equal(t, int32(3), n.Load())
	})
}
