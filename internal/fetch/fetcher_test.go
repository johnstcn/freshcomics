package fetch

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/johnstcn/freshcomics/internal/testutil/slogtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type mockClient struct {
	mock.Mock
}

func (c *mockClient) Do(r *http.Request) (*http.Response, error) {
	args := c.Called(r)
	return args.Get(0).(*http.Response), args.Error(1)
}

// mockAfterer helps verify correct backoff
type mockAfterer struct {
	mock.Mock
}

func (m *mockAfterer) after(d time.Duration) <-chan time.Time {
	m.Called(d)
	ch := make(chan time.Time)
	go func() {
		ch <- time.Time{}
	}()
	return ch
}

type badReader struct{}

func (br *badReader) Read([]byte) (int, error) {
	return 0, fmt.Errorf("could not read")
}

var _ io.Reader = (*badReader)(nil)

type PageFetcherTestSuite struct {
	suite.Suite
	client  *mockClient
	afterer *mockAfterer
	fetcher *pageFetcher
}

func TestPageFetcherTestSuite(t *testing.T) {
	suite.Run(t, new(PageFetcherTestSuite))
}

func (s *PageFetcherTestSuite) SetupSuite() {
	s.client = &mockClient{}
	s.afterer = &mockAfterer{}
	s.fetcher = &pageFetcher{
		client:      s.client,
		maxAttempts: 2,
		backoff:     backoffExponential,
		after:       s.afterer.after,
		log:         slogtest.New(s.T()),
	}
}

func (s *PageFetcherTestSuite) TearDownTest() {
	s.client.AssertExpectations(s.T())
	s.afterer.AssertExpectations(s.T())
}

func (s *PageFetcherTestSuite) TestBackoff() {
	s.Equal(0*time.Second, backoffExponential(0))
	s.Equal(1*time.Second, backoffExponential(1))
	s.Equal(2*time.Second, backoffExponential(2))
	s.Equal(4*time.Second, backoffExponential(3))
	s.Equal(8*time.Second, backoffExponential(4))
}

func (s *PageFetcherTestSuite) TestNewPageFetcher() {
	pf := NewPageFetcher(http.DefaultClient,
		1, "test")
	s.NotNil(pf)
	s.Implements((*PageFetcher)(nil), pf)
}

func (s *PageFetcherTestSuite) TestFetch_OK() {
	testUrl := "http://test.url/path"
	s.client.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("body")),
	}, nil).Once()

	p, err := s.fetcher.Fetch(testUrl)

	s.NoError(err)
	s.Equal(http.StatusOK, p.ResponseCode)
	s.Equal(testUrl, p.URL)
	s.EqualValues([]byte("body"), p.Body)
}

func (s *PageFetcherTestSuite) TestFetchOneAttempt_Err_Do() {
	testUrl := "http://test.url/path"
	testReq, _ := http.NewRequest(http.MethodGet, testUrl, nil)
	s.client.On("Do", mock.AnythingOfType("*http.Request")).Return((*http.Response)(nil), errors.New("client error")).Once()

	p, err := s.fetcher.fetchOneAttempt(testReq)

	s.EqualError(err, "client error")
	s.Zero(p.ResponseCode)
	s.Equal(testUrl, p.URL)
	s.Empty(p.Body)
}

func (s *PageFetcherTestSuite) TestFetchOneAttempt_Err_Read() {
	testUrl := "http://test.url/path"
	testReq, _ := http.NewRequest(http.MethodGet, testUrl, nil)
	s.client.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(&badReader{}),
	}, nil).Once()

	p, err := s.fetcher.fetchOneAttempt(testReq)

	s.EqualError(err, "read response body: could not read")
	s.Zero(p.ResponseCode)
	s.Equal(testUrl, p.URL)
	s.Empty(p.Body)
}

func (s *PageFetcherTestSuite) TestFetchWithRetry_Retry_OK() {
	testUrl := "http://test.url/path"
	s.client.On("Do", mock.AnythingOfType("*http.Request")).Return((*http.Response)(nil), fmt.Errorf("try again")).Once()
	s.afterer.On("after", 1*time.Second).Return().Once()
	s.client.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("body")),
	}, nil).Once()

	p, err := s.fetcher.fetchWithRetry(testUrl)

	s.NoError(err)
	s.Equal(http.StatusOK, p.ResponseCode)
	s.Equal(testUrl, p.URL)
	s.EqualValues([]byte("body"), p.Body)
}

func (s *PageFetcherTestSuite) TestFetchWithRetry_Retry_Err() {
	testUrl := "http://test.url/path"
	s.client.On("Do", mock.AnythingOfType("*http.Request")).Return((*http.Response)(nil), fmt.Errorf("try again")).Once()
	s.afterer.On("after", 1*time.Second).Return().Once()
	s.client.On("Do", mock.AnythingOfType("*http.Request")).Return((*http.Response)(nil), fmt.Errorf("give up already")).Once()

	p, err := s.fetcher.fetchWithRetry(testUrl)

	s.EqualError(err, "give up already")
	s.Equal(testUrl, p.URL)
	s.Zero(p.ResponseCode)
	s.Zero(p.Body)
}
