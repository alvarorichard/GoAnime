package player

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatchProgressAggregatesEpisodeChildren(t *testing.T) {
	parent := &model{}
	ep1 := parent.childProgress("episode-1", 100)
	ep2 := parent.childProgress("episode-2", 200)

	ep1.addProgressReceived(99)
	ep2.addProgressReceived(99)

	parent.mu.Lock()
	received := parent.received
	total := parent.totalBytes
	parent.mu.Unlock()

	assert.Equal(t, int64(198), received)
	assert.Equal(t, int64(300), total)
	assert.InDelta(t, 0.66, float64(received)/float64(total), 0.001)
}

func TestBatchProgressChildTotalAdjustsGlobalTotalWithoutResettingReceived(t *testing.T) {
	parent := &model{}
	ep1 := parent.childProgress("episode-1", 100)
	ep2 := parent.childProgress("episode-2", 200)

	ep1.addProgressReceived(50)
	ep2.addProgressReceived(100)
	ep1.setProgressTotal(120)

	parent.mu.Lock()
	received := parent.received
	total := parent.totalBytes
	parent.mu.Unlock()

	assert.Equal(t, int64(150), received)
	assert.Equal(t, int64(320), total)
}

func TestBatchProgressAbsoluteEpisodeUpdatesDoNotResetGlobalProgress(t *testing.T) {
	parent := &model{}
	ep1 := parent.childProgress("episode-1", 100)
	ep2 := parent.childProgress("episode-2", 100)

	ep1.setProgressReceived(90)
	ep2.setProgressReceived(10)
	ep2.setProgressReceived(80)
	ep1.setProgressReceived(95)

	parent.mu.Lock()
	received := parent.received
	total := parent.totalBytes
	parent.mu.Unlock()

	assert.Equal(t, int64(175), received)
	assert.Equal(t, int64(200), total)
}

func TestBatchProgressFallbackCanResetOnlyCurrentEpisode(t *testing.T) {
	parent := &model{}
	ep1 := parent.childProgress("episode-1", 100)
	ep2 := parent.childProgress("episode-2", 100)

	ep1.addProgressReceived(70)
	ep2.addProgressReceived(40)
	ep2.resetProgressReceived()
	ep2.addProgressReceived(25)

	parent.mu.Lock()
	received := parent.received
	total := parent.totalBytes
	parent.mu.Unlock()

	assert.Equal(t, int64(95), received)
	assert.Equal(t, int64(200), total)
}
