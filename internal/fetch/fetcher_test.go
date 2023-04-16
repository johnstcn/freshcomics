package fetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johnstcn/freshcomics/internal/testutil/slogtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPageFetcher(t *testing.T) {
	t.Parallel()

	t.Run("New", func(t *testing.T) {
		t.Parallel()
		p := New(&Args{})
		assert.NotNil(t, p)
	})

	t.Run("OK", func(t *testing.T) {
		t.Parallel()
		var (
			ctx, cancel = context.WithCancel(context.Background())
			useragent   = "testing"
			body        = "body"
			handler     = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, useragent, r.Header.Get("User-Agent"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(body))
			})
			srv    = httptest.NewServer(handler)
			client = srv.Client()
		)
		t.Cleanup(cancel)
		t.Cleanup(srv.Close)

		pf := &pageFetcher{
			client:    client,
			retries:   0,
			wait:      1,
			userAgent: useragent,
			log:       slogtest.New(t),
		}

		p, err := pf.Fetch(ctx, srv.URL)
		require.NoError(t, err)
		require.Equal(t, srv.URL, p.URL)
		require.Equal(t, body, string(p.Body))
		require.Equal(t, http.StatusOK, p.ResponseCode)
		require.Zero(t, p.Retries)
	})

	t.Run("RetryOK", func(t *testing.T) {
		t.Parallel()
		var (
			ctx, cancel = context.WithCancel(context.Background())
			useragent   = "testing"
			body        = "body"
			calls       = 0
			handler     = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if calls == 0 {
					// Fake an invalid body
					w.Header().Set("Content-Length", "1")
				}
				calls++
				assert.Equal(t, useragent, r.Header.Get("User-Agent"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(body))
			})
			srv    = httptest.NewServer(handler)
			client = srv.Client()
		)
		t.Cleanup(srv.Close)
		t.Cleanup(cancel)

		pf := &pageFetcher{
			client:    client,
			retries:   1,
			wait:      1,
			userAgent: useragent,
			log:       slogtest.New(t),
		}

		p, err := pf.Fetch(ctx, srv.URL)
		require.NoError(t, err)
		assert.Equal(t, srv.URL, p.URL)
		assert.Equal(t, body, string(p.Body))
		assert.Equal(t, http.StatusOK, p.ResponseCode)
		assert.Equal(t, 1, p.Retries)
		assert.Equal(t, 2, calls)
	})

	t.Run("RetryErr", func(t *testing.T) {
		t.Parallel()
		var (
			ctx, cancel = context.WithCancel(context.Background())
			useragent   = "testing"
			handler     = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Fake an invalid body
				w.Header().Set("Content-Length", "1")
			})
			srv    = httptest.NewServer(handler)
			client = srv.Client()
		)
		t.Cleanup(srv.Close)
		t.Cleanup(cancel)

		pf := &pageFetcher{
			client:    client,
			retries:   1,
			wait:      1,
			userAgent: useragent,
			log:       slogtest.New(t),
		}

		p, err := pf.Fetch(ctx, srv.URL)
		require.Error(t, err)
		assert.Equal(t, 1, p.Retries)
	})
}
