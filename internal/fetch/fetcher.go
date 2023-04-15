package fetch

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"golang.org/x/exp/slog"
)

//go:generate mockery -interface PageFetcher -package fetchtest

// Doer is an interface satisfied by http.Client
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

var _ Doer = (*http.Client)(nil)

// backoffFunc returns a time.Duration as a function of number of maxAttempts
type backoffFunc func(attempt int) time.Duration

// afterFunc is satisfied by time.After
type afterFunc func(d time.Duration) <-chan time.Time

var backoffExponential = func(attempt int) time.Duration {
	waitSecs := math.Pow(2.0, float64(attempt-1))
	return time.Duration(waitSecs) * time.Second
}

// FetchedPage holds the result of fetching a page
type FetchedPage struct {
	URL          string // URL fetched
	ResponseCode int    // Response code returned
	Body         []byte // Response body
}

// PageFetcher fetches a given URL with a number of maxAttempts
type PageFetcher interface {
	Fetch(URL string) (FetchedPage, error)
}

var _ PageFetcher = (*pageFetcher)(nil)

// pageFetcher implements PageFetcher
type pageFetcher struct {
	client      Doer
	maxAttempts int
	backoff     backoffFunc
	after       afterFunc
	userAgent   string
	log         *slog.Logger
}

// NewPageFetcher returns a new PageFetcher with exponential backoff
func NewPageFetcher(httpClient Doer, retries int, userAgent string) PageFetcher {
	return &pageFetcher{
		client:      httpClient,
		maxAttempts: retries,
		backoff:     backoffExponential,
		after:       time.After,
		userAgent:   userAgent,
	}
}

// Fetch fetches the given URL and returns a FetchedPage
func (f *pageFetcher) Fetch(URL string) (FetchedPage, error) {
	return f.fetchWithRetry(URL)
}

func (f *pageFetcher) fetchWithRetry(url string) (FetchedPage, error) {
	var p FetchedPage
	var attempt int

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return FetchedPage{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Add("User-Agent", f.userAgent)

	for {
		f.log.Debug("get", "attempt", attempt, "max", f.maxAttempts, "url", url)
		p, err = f.fetchOneAttempt(req)
		if err == nil {
			return p, nil
		}
		attempt++
		if attempt < f.maxAttempts {
			f.log.Debug("backoff", "attempt", attempt, "max", f.maxAttempts, "url", url)
			<-f.after(f.backoff(attempt))
			continue
		}

		// bail here, max attempts reached
		f.log.Error("get failed", "attempts", f.maxAttempts, "url", url, "err", err)
		break
	}

	return p, err
}

func (f *pageFetcher) fetchOneAttempt(r *http.Request) (FetchedPage, error) {
	p := FetchedPage{
		URL: r.URL.String(),
	}

	resp, err := f.client.Do(r)
	if err != nil {
		return p, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return p, fmt.Errorf("read response body: %w", err)
	}

	p.ResponseCode = resp.StatusCode
	p.Body = body
	return p, nil
}
