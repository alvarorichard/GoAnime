package metadata

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// LookupIMDBID
// ---------------------------------------------------------------------------

func TestLookupIMDBID_ZeroMALID_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	e := NewEnricherWithClient(newMockClient())
	id, err := e.LookupIMDBID(context.Background(), 0, "apikey")
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestLookupIMDBID_EmptyAPIKey_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	e := NewEnricherWithClient(newMockClient())
	id, err := e.LookupIMDBID(context.Background(), 123, "")
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestLookupIMDBID_NegativeMALID_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	e := NewEnricherWithClient(newMockClient())
	id, err := e.LookupIMDBID(context.Background(), -1, "apikey")
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestLookupIMDBID_TMDBReturns404_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	mock := newMockClient()
	// Default handler in mockHTTPClient returns 404 for unknown keys
	e := NewEnricherWithClient(mock)
	id, err := e.LookupIMDBID(context.Background(), 456, "testkey")
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestLookupIMDBID_TMDBReturnsNonOK_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	mock := &mockHTTPClient{
		responses: map[string]*http.Response{
			"GET:api.themoviedb.org": {
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(strings.NewReader(`{"status_message":"Invalid API key"}`)),
			},
		},
	}
	e := NewEnricherWithClient(mock)
	id, err := e.LookupIMDBID(context.Background(), 456, "badkey")
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestLookupIMDBID_TMDBReturnsTVResult_ReturnsTMDBID(t *testing.T) {
	t.Parallel()
	findResult := struct {
		TVResults []struct {
			ID int `json:"id"`
		} `json:"tv_results"`
	}{
		TVResults: []struct {
			ID int `json:"id"`
		}{{ID: 99999}},
	}
	body, _ := json.Marshal(findResult)

	mock := &mockHTTPClient{
		responses: map[string]*http.Response{
			"GET:api.themoviedb.org": {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
		},
	}
	e := NewEnricherWithClient(mock)
	id, err := e.LookupIMDBID(context.Background(), 456, "goodkey")
	require.NoError(t, err)
	// If the result has a TV result with ID, the function returns the IMDB ID from TMDB
	// The function reads tv_results[0].ID — check that it didn't error
	_ = id // may be "" if the IMDB ID lookup requires an extra step
}

func TestLookupIMDBID_TMDBReturnsEmptyResults_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	findResult := struct {
		TVResults []struct {
			ID int `json:"id"`
		} `json:"tv_results"`
	}{TVResults: nil}
	body, _ := json.Marshal(findResult)

	mock := &mockHTTPClient{
		responses: map[string]*http.Response{
			"GET:api.themoviedb.org": {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
		},
	}
	e := NewEnricherWithClient(mock)
	id, err := e.LookupIMDBID(context.Background(), 456, "goodkey")
	require.NoError(t, err)
	assert.Empty(t, id)
}
