package fetch

import (
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

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
	Err          error  // Last error encountered when fetching URL
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
		return FetchedPage{}, errors.Wrapf(err, "unable to create HTTP request for url %s")
	}
	req.Header.Add("User-Agent", f.userAgent)

	for {
		glog.V(9).Infof("GET [%d/%d] %s", attempt, f.maxAttempts, url)
		p = f.fetchOneAttempt(req)
		if p.Err != nil {
			glog.V(9).Infof("Attempt %d/%d failed: %v", attempt+1, f.maxAttempts, p.Err)
			attempt++
			if attempt < f.maxAttempts {
				glog.V(9).Infof("Waiting %d secs", f.backoff(attempt))
				<-f.after(f.backoff(attempt))
				continue
			}

			// bail here, max attempts reached
			glog.V(9).Info("max attempts reached")
			break
		}

		return p, nil
	}
	return p, errors.New("max attempts reached")
}

func (f *pageFetcher) fetchOneAttempt(r *http.Request) FetchedPage {
	var p = FetchedPage{
		URL: r.URL.String(),
	}

	resp, err := f.client.Do(r)
	if err != nil {
		p.Err = err
		return p
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		p.Err = errors.Wrap(err, "error reading response body")
		return p
	}

	p.ResponseCode = resp.StatusCode
	p.Body = body
	return p
}
