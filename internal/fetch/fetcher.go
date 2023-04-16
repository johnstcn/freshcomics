package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/exp/slog"
)

// FetchedPage holds the result of fetching a page
type FetchedPage struct {
	URL          string // URL fetched
	ResponseCode int    // Response code returned
	Body         []byte // Response body
	Retries      int    // Number of retries
}

// Fetcher fetches a given URL
type Fetcher interface {
	Fetch(ctx context.Context, url string) (FetchedPage, error)
}

var _ Fetcher = (*pageFetcher)(nil)

// pageFetcher implements PageFetcher
type pageFetcher struct {
	client    *http.Client
	retries   int
	wait      time.Duration
	after     func(d time.Duration) <-chan time.Time
	userAgent string
	log       *slog.Logger
}

type Args struct {
	Client    *http.Client
	UserAgent string
	Retries   int
	Wait      time.Duration
}

// New returns a new PageFetcher
func New(a *Args) Fetcher {
	return &pageFetcher{
		client:    a.Client,
		retries:   a.Retries,
		wait:      a.Wait,
		userAgent: a.UserAgent,
		after:     time.After,
	}
}

// Fetch fetches the given URL and returns a FetchedPage
func (f *pageFetcher) Fetch(ctx context.Context, url string) (FetchedPage, error) {
	var p FetchedPage
	p.URL = url

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return FetchedPage{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Add("User-Agent", f.userAgent)

	for {
		select {
		case <-ctx.Done():
			return FetchedPage{}, ctx.Err()
		default:
			f.log.Debug("get", "retry", p.Retries, "max", f.retries, "url", url)
			code, body, err := fetchOnce(f.client, req)
			p.ResponseCode = code
			p.Body = body
			if err == nil {
				return p, nil
			}
			if p.Retries >= f.retries {
				f.log.Error("get failed after retry", "retries", f.retries, "url", url, "err", err)
				return p, err
			}
			p.Retries++
			f.log.Debug("retry", "retry", p.Retries, "max", f.retries, "url", url)
			<-time.After(f.wait)
		}
	}
}

func fetchOnce(c *http.Client, r *http.Request) (int, []byte, error) {
	resp, err := c.Do(r)
	if err != nil {
		return 0, nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("read response body: %w", err)
	}

	return resp.StatusCode, body, nil
}
