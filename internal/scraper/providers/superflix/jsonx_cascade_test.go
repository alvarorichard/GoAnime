package superflix

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/util/jsonx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tvmazeCascadeServer serves the same two-hop TVmaze cascade the lister walks
// (lookup → episodes) on Go 1.27's in-memory test server: no TCP port is bound,
// so these tests cannot flake on port exhaustion or a sandboxed network.
func tvmazeCascadeServer(t *testing.T, lookup, episodes func(w http.ResponseWriter)) *http.Client {
	t.Helper()
	srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/lookup/shows":
			lookup(w)
		case strings.HasSuffix(r.URL.Path, "/episodes"):
			episodes(w)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	// Client() starts the in-memory server and fills in srv.URL; its transport
	// routes every address to this server, so no port is ever bound.
	client := srv.Client()
	prev := tvmazeBaseURL
	tvmazeBaseURL = srv.URL
	t.Cleanup(func() { tvmazeBaseURL = prev })
	return client
}

func writeString(s string) func(http.ResponseWriter) {
	return func(w http.ResponseWriter) { _, _ = w.Write([]byte(s)) }
}

// TestTVmazeCascade_StreamingDecodeHappyPath walks the full cascade through
// jsonx.Decode (streaming, no intermediate buffer) and checks the decoded
// result is what the buffered implementation produced.
func TestTVmazeCascade_StreamingDecodeHappyPath(t *testing.T) {
	client := tvmazeCascadeServer(t,
		writeString(`{"id":15299,"name":"The Boys"}`),
		writeString(`[
			{"season":1,"number":1,"name":"The Name of the Game","airdate":"2019-07-26"},
			{"season":1,"number":2,"name":"Cherry","airdate":"2019-07-26"},
			{"season":2,"number":1,"name":"The Big Ride","airdate":"2020-09-04"}
		]`))

	got, err := GetEpisodesFromTVmaze(context.Background(), client, "tt1190634")
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Len(t, got["1"], 2)
	assert.Len(t, got["2"], 1)
	assert.Equal(t, "Cherry", got["1"][1].Title)
}

// TestTVmazeCascade_CaseInsensitiveFieldsStillMatch proves the v1 field-matching
// semantics survived the move to encoding/json/v2. Without
// MatchCaseInsensitiveNames these payloads would decode to zero values and the
// cascade would silently report "no episodes" instead of failing loudly.
func TestTVmazeCascade_CaseInsensitiveFieldsStillMatch(t *testing.T) {
	client := tvmazeCascadeServer(t,
		writeString(`{"ID":15299,"Name":"The Boys"}`),
		writeString(`[{"Season":1,"NUMBER":1,"Name":"Pilot","AirDate":"2019-07-26"}]`))

	got, err := GetEpisodesFromTVmaze(context.Background(), client, "tt1190634")
	require.NoError(t, err)
	require.Len(t, got["1"], 1)
	assert.Equal(t, "Pilot", got["1"][0].Title)
	assert.Equal(t, "1", got["1"][0].EpiNum.String())
}

// TestTVmazeCascade_InvalidUTF8IsTolerated mirrors real scraped titles: one bad
// byte in a name must not fail the whole episode list.
func TestTVmazeCascade_InvalidUTF8IsTolerated(t *testing.T) {
	client := tvmazeCascadeServer(t,
		writeString(`{"id":15299,"name":"The Boys"}`),
		writeString("[{\"season\":1,\"number\":1,\"name\":\"Caf\xff\",\"airdate\":\"2019-07-26\"}]"))

	got, err := GetEpisodesFromTVmaze(context.Background(), client, "tt1190634")
	require.NoError(t, err)
	require.Len(t, got["1"], 1)
	assert.Contains(t, got["1"][0].Title, "Caf")
}

// TestTVmazeCascade_OversizedBodyIsRejected is the security half of the
// migration: a hostile or broken upstream that streams forever is cut off at the
// declared limit and reported as such, instead of being read into memory.
func TestTVmazeCascade_OversizedBodyIsRejected(t *testing.T) {
	client := tvmazeCascadeServer(t,
		writeString(`{"id":15299,"name":"The Boys"}`),
		func(w http.ResponseWriter) {
			_, _ = w.Write([]byte(`[{"season":1,"number":1,"name":"`))
			chunk := strings.Repeat("x", 64*1024)
			// Well past the 4 MiB cap; the decoder must stop us long before the
			// loop finishes.
			for range 200 {
				if _, err := w.Write([]byte(chunk)); err != nil {
					return
				}
			}
			_, _ = w.Write([]byte(`","airdate":"2019-07-26"}]`))
		})

	_, err := GetEpisodesFromTVmaze(context.Background(), client, "tt1190634")
	require.Error(t, err)
	assert.True(t, errors.Is(err, jsonx.ErrTooLarge),
		"expected jsonx.ErrTooLarge, got %v", err)
	// That the decoder actually stops reading at the cap (rather than draining
	// the whole stream and only then complaining) is asserted directly against
	// the reader in TestDecodeStopsReadingAtLimit.
}

// TestTVmazeCascade_LookupFailureShortCircuits keeps the cascade contract: a
// broken first hop must not attempt the second.
func TestTVmazeCascade_LookupFailureShortCircuits(t *testing.T) {
	episodesHit := false
	client := tvmazeCascadeServer(t,
		func(w http.ResponseWriter) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"boom"}`))
		},
		func(w http.ResponseWriter) {
			episodesHit = true
			_, _ = w.Write([]byte(`[]`))
		})

	_, err := GetEpisodesFromTVmaze(context.Background(), client, "tt1190634")
	require.Error(t, err)
	assert.False(t, episodesHit, "episodes endpoint must not be reached after a lookup failure")
}

// TestTVmazeCascade_MalformedJSONFailsCleanly makes sure a truncated payload
// surfaces as an error rather than a half-populated result.
func TestTVmazeCascade_MalformedJSONFailsCleanly(t *testing.T) {
	client := tvmazeCascadeServer(t,
		writeString(`{"id":15299,"name":"The Boys"}`),
		writeString(`[{"season":1,"number":1,`))

	got, err := GetEpisodesFromTVmaze(context.Background(), client, "tt1190634")
	require.Error(t, err)
	assert.Nil(t, got)
	assert.False(t, errors.Is(err, jsonx.ErrTooLarge),
		"a syntax error must not be reported as a size violation")
	assert.Contains(t, fmt.Sprint(err), "tvmaze episodes")
}
