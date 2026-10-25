package api

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		in        float64
		precision int
		want      float64
	}{
		{"zero precision", 1.7, 0, 2},
		{"one precision", 1.45, 1, 1.5},
		{"two precision", 1.234, 2, 1.23},
		{"two precision round up", 1.235, 2, 1.24},
		{"negative", -1.5, 0, -1},
		{"already round", 5.0, 2, 5.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.InDelta(t, tt.want, RoundTime(tt.in, tt.precision), 0.0001)
		})
	}
}

func TestParseAniSkipResponse(t *testing.T) {
	t.Parallel()

	t.Run("empty response errors", func(t *testing.T) {
		t.Parallel()
		ep := &models.Episode{}
		err := ParseAniSkipResponse("", ep, 0)
		assert.Error(t, err)
	})

	t.Run("malformed json errors", func(t *testing.T) {
		t.Parallel()
		ep := &models.Episode{}
		err := ParseAniSkipResponse("not json", ep, 0)
		assert.Error(t, err)
	})

	t.Run("not found errors", func(t *testing.T) {
		t.Parallel()
		body := `{"found":false,"results":[]}`
		ep := &models.Episode{}
		err := ParseAniSkipResponse(body, ep, 0)
		assert.Error(t, err)
	})

	t.Run("populates op and ed", func(t *testing.T) {
		t.Parallel()
		body := `{"found":true,"results":[
			{"skip_type":"op","interval":{"start_time":10.5,"end_time":90.4}},
			{"skip_type":"ed","interval":{"start_time":1200.0,"end_time":1320.0}}
		]}`
		ep := &models.Episode{}
		err := ParseAniSkipResponse(body, ep, 0)
		require.NoError(t, err)
		assert.Equal(t, 11, ep.SkipTimes.Op.Start)
		assert.Equal(t, 90, ep.SkipTimes.Op.End)
		assert.Equal(t, 1200, ep.SkipTimes.Ed.Start)
		assert.Equal(t, 1320, ep.SkipTimes.Ed.End)
	})

	t.Run("unknown skip type ignored", func(t *testing.T) {
		t.Parallel()
		body := `{"found":true,"results":[
			{"skip_type":"mystery","interval":{"start_time":1.0,"end_time":2.0}}
		]}`
		ep := &models.Episode{}
		err := ParseAniSkipResponse(body, ep, 0)
		require.NoError(t, err)
		assert.Equal(t, 0, ep.SkipTimes.Op.Start)
		assert.Equal(t, 0, ep.SkipTimes.Ed.Start)
	})
}

func TestGetAniSkipData_NetworkErrorOnDisallowedHost(t *testing.T) {
	t.Parallel()
	_, err := GetAniSkipData(-1, 0)
	assert.Error(t, err, "expect error for invalid request (network or non-200)")
}

func TestGetAndParseAniSkipData_PropagatesFetchError(t *testing.T) {
	t.Parallel()
	ep := &models.Episode{}
	err := GetAndParseAniSkipData(-1, -1, ep)
	assert.Error(t, err)
}
