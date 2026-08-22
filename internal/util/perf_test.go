package util

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withPerfEnabled(t *testing.T) {
	t.Helper()
	prev := PerfEnabled
	PerfEnabled = true
	t.Cleanup(func() { PerfEnabled = prev })
}

func TestGetPerfTracker_Singleton(t *testing.T) {
	a := GetPerfTracker()
	b := GetPerfTracker()
	assert.Same(t, a, b)
	require.NotNil(t, a)
}

func TestStartTimer_DisabledReturnsNil(t *testing.T) {
	prev := PerfEnabled
	PerfEnabled = false
	t.Cleanup(func() { PerfEnabled = prev })
	assert.Nil(t, StartTimer("x"))
}

func TestStartTimer_EnabledMeasures(t *testing.T) {
	withPerfEnabled(t)
	timer := StartTimer("TestStartTimer_EnabledMeasures")
	require.NotNil(t, timer)
	time.Sleep(5 * time.Millisecond)
	d := timer.Stop()
	assert.Greater(t, d, time.Duration(0))
}

func TestTimer_Stop_DisabledReturnsZero(t *testing.T) {
	prev := PerfEnabled
	PerfEnabled = false
	t.Cleanup(func() { PerfEnabled = prev })
	timer := &Timer{name: "x", start: time.Now()}
	assert.Equal(t, time.Duration(0), timer.Stop())
}

func TestTimer_StopAndLog(t *testing.T) {
	withPerfEnabled(t)
	timer := StartTimer("TestTimer_StopAndLog")
	require.NotNil(t, timer)
	d := timer.StopAndLog()
	assert.GreaterOrEqual(t, d, time.Duration(0))
}

func TestPerfTracker_Record(t *testing.T) {
	pt := &PerfTracker{metrics: make(map[string]*PerfMetric), counters: make(map[string]*int64)}
	pt.Record("op", 100*time.Millisecond)
	pt.Record("op", 200*time.Millisecond)
	metrics := pt.GetMetrics()
	require.Contains(t, metrics, "op")
	assert.Equal(t, int64(2), metrics["op"].Count)
	assert.Equal(t, 300*time.Millisecond, metrics["op"].TotalTime)
}

func TestPerfTracker_IncrementAndGetCounter(t *testing.T) {
	pt := &PerfTracker{metrics: make(map[string]*PerfMetric), counters: make(map[string]*int64)}
	pt.IncrementCounter("hits")
	pt.IncrementCounter("hits")
	pt.IncrementCounter("hits")
	assert.Equal(t, int64(3), pt.GetCounter("hits"))
	assert.Equal(t, int64(0), pt.GetCounter("missing"))
}

func TestPerfTracker_GetMetrics_ReturnsCopies(t *testing.T) {
	pt := &PerfTracker{metrics: make(map[string]*PerfMetric), counters: make(map[string]*int64)}
	pt.Record("op", time.Millisecond)
	metricsCopy := pt.GetMetrics()
	metricsCopy["op"].Count = 999
	require.NotEqual(t, int64(999), pt.GetMetrics()["op"].Count)
}

func TestPerfTracker_GetUptime(t *testing.T) {
	pt := &PerfTracker{started: time.Now().Add(-2 * time.Second), counters: map[string]*int64{}, metrics: map[string]*PerfMetric{}}
	assert.GreaterOrEqual(t, pt.GetUptime(), 2*time.Second)
}

func TestPerfTracker_Reset(t *testing.T) {
	pt := &PerfTracker{metrics: make(map[string]*PerfMetric), counters: make(map[string]*int64)}
	pt.Record("op", time.Millisecond)
	pt.IncrementCounter("c")
	pt.Reset()
	assert.Empty(t, pt.GetMetrics())
	assert.Equal(t, int64(0), pt.GetCounter("c"))
}

func TestPerfTracker_PrintReport_DisabledNoOp(t *testing.T) {
	prev := PerfEnabled
	PerfEnabled = false
	t.Cleanup(func() { PerfEnabled = prev })
	pt := GetPerfTracker()
	assert.NotPanics(t, func() { pt.PrintReport() })
}

func TestTimeFunc_DisabledStillRuns(t *testing.T) {
	prev := PerfEnabled
	PerfEnabled = false
	t.Cleanup(func() { PerfEnabled = prev })
	ran := false
	TimeFunc("x", func() { ran = true })
	assert.True(t, ran)
}

func TestTimeFunc_EnabledRecords(t *testing.T) {
	withPerfEnabled(t)
	ran := false
	TimeFunc("TestTimeFunc_EnabledRecords", func() { ran = true })
	assert.True(t, ran)
}

func TestTimeFuncWithResult(t *testing.T) {
	withPerfEnabled(t)
	got := TimeFuncWithResult("TestTimeFuncWithResult", func() int { return 42 })
	assert.Equal(t, 42, got)
}

func TestTimeFuncWithResult_Disabled(t *testing.T) {
	prev := PerfEnabled
	PerfEnabled = false
	t.Cleanup(func() { PerfEnabled = prev })
	assert.Equal(t, "x", TimeFuncWithResult("y", func() string { return "x" }))
}

func TestTimeFuncWithError(t *testing.T) {
	withPerfEnabled(t)
	v, err := TimeFuncWithError("TestTimeFuncWithError", func() (int, error) {
		return 7, errors.New("boom")
	})
	assert.Equal(t, 7, v)
	assert.Error(t, err)
}

func TestTimeFuncWithError_Disabled(t *testing.T) {
	prev := PerfEnabled
	PerfEnabled = false
	t.Cleanup(func() { PerfEnabled = prev })
	v, err := TimeFuncWithError("x", func() (string, error) { return "ok", nil })
	assert.Equal(t, "ok", v)
	assert.NoError(t, err)
}

func TestPerf_RecordsWhenEnabled(t *testing.T) {
	withPerfEnabled(t)
	pt := GetPerfTracker()
	pt.Reset()
	Perf("TestPerf_RecordsWhenEnabled", time.Now().Add(-100*time.Millisecond))
	metrics := pt.GetMetrics()
	require.Contains(t, metrics, "TestPerf_RecordsWhenEnabled")
}

func TestPerf_DisabledNoRecord(t *testing.T) {
	prev := PerfEnabled
	PerfEnabled = false
	t.Cleanup(func() { PerfEnabled = prev })
	pt := GetPerfTracker()
	pt.Reset()
	Perf("nope", time.Now())
	assert.Empty(t, pt.GetMetrics())
}

func TestPerfCount(t *testing.T) {
	withPerfEnabled(t)
	pt := GetPerfTracker()
	pt.Reset()
	PerfCount("hits")
	PerfCount("hits")
	assert.Equal(t, int64(2), pt.GetCounter("hits"))
}

func TestPerfTracker_ConcurrentCounters(t *testing.T) {
	pt := &PerfTracker{metrics: make(map[string]*PerfMetric), counters: make(map[string]*int64)}
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() { pt.IncrementCounter("hits") })
	}
	wg.Wait()
	assert.Equal(t, int64(100), pt.GetCounter("hits"))
}

func TestPrintReport_EnabledWithData(t *testing.T) {
	withPerfEnabled(t)

	pt := &PerfTracker{
		metrics:  make(map[string]*PerfMetric),
		started:  time.Now().Add(-10 * time.Second),
		counters: make(map[string]*int64),
	}

	// Slow entry (TotalTime > 5s) → perfSlowStyle for total
	pt.Record("SlowOperation", 6*time.Second)
	// Fast entry (TotalTime < 500ms) → perfFastStyle for total
	pt.Record("FastOp", 100*time.Millisecond)
	// Medium entry (500ms ≤ TotalTime ≤ 5s) → perfValueStyle for total
	pt.Record("MediumOperation", 2*time.Second)
	// Long name (> 38 chars) → truncation branch
	pt.Record("VeryLongOperationNameThatExceedsThirtyEightCharactersLimit", 300*time.Millisecond)
	// Slow average (avg > 2s): two records → avg = 3s
	pt.Record("SlowAvgOp", 3*time.Second)
	pt.Record("SlowAvgOp", 3*time.Second)
	// Fast average (avg < 200ms): two records → avg = 50ms
	pt.Record("FastAvgOp", 50*time.Millisecond)
	pt.Record("FastAvgOp", 50*time.Millisecond)
	// Medium average (200ms ≤ avg ≤ 2s)
	pt.Record("MedAvgOp", 500*time.Millisecond)
	pt.Record("MedAvgOp", 500*time.Millisecond)
	// Counter branch
	pt.IncrementCounter("cache_hits")
	pt.IncrementCounter("cache_hits")

	assert.NotPanics(t, func() { pt.PrintReport() })
}

func TestPrintReport_EnabledEmpty(t *testing.T) {
	withPerfEnabled(t)

	pt := &PerfTracker{
		metrics:  make(map[string]*PerfMetric),
		started:  time.Now(),
		counters: make(map[string]*int64),
	}

	// No metrics, no counters — exercises early-return paths for both sections.
	assert.NotPanics(t, func() { pt.PrintReport() })
}
