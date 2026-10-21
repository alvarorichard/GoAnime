package animedrive

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnimeDriveSearchContext_UsesCallerBudget(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		delay    time.Duration
		deadline time.Duration
		wantTime time.Duration
		wantErr  bool
	}{
		{name: "slow response within search deadline", delay: 9 * time.Second, deadline: 12 * time.Second, wantTime: 9 * time.Second},
		{name: "upstream exceeds search deadline", delay: 13 * time.Second, deadline: 12 * time.Second, wantTime: 12 * time.Second, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				requests := 0
				srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests++
					select {
					case <-r.Context().Done():
						return
					case <-time.After(tt.delay):
						_, _ = fmt.Fprint(w, `<div class="result-item"><article><div class="title"><a href="/anime/frieren/">Frieren</a></div></article></div>`)
					}
				}))
				httpClient := srv.Client()
				client := NewClientForTest(srv.URL)
				// Keep the production fast-client timeout while replacing only its
				// network transport with the in-memory test server.
				httpClient.Timeout = client.client.Timeout
				client.SetHTTPClientForTest(httpClient)
				client.maxRetries = 2
				ctx, cancel := context.WithTimeout(context.Background(), tt.deadline)
				defer cancel()

				start := time.Now()
				results, err := client.SearchAnimeContext(ctx, "frieren")
				if tt.wantErr {
					require.ErrorIs(t, err, context.DeadlineExceeded)
				} else {
					require.NoError(t, err)
					require.Len(t, results, 1)
					assert.Equal(t, "Frieren", results[0].Name)
				}
				synctest.Wait()
				assert.Equal(t, tt.wantTime, time.Since(start))
				assert.Equal(t, 1, requests, "do not restart a request before the caller's deadline")
				assert.Equal(t, 8*time.Second, client.client.Timeout, "keep the shared client unchanged")
			})
		})
	}
}

func TestAnimeDriveSearch_WithoutDeadlineKeepsClientTimeout(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}))
		httpClient := srv.Client()
		client := NewClientForTest(srv.URL)
		httpClient.Timeout = client.client.Timeout
		client.SetHTTPClientForTest(httpClient)

		start := time.Now()
		_, err := client.SearchAnime("frieren")
		require.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Equal(t, 8*time.Second, time.Since(start))
	})
}
