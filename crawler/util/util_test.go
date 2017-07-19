package util

import (
	"time"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/xmlpath.v2"

	"github.com/johnstcn/freshcomics/crawler/models"
)

func TestApplyRegex(t *testing.T) {
	result, err := ApplyRegex("foo bar", "(.+)bar")
	assert.Nil(t, err)
	assert.Equal(t, "foo", result)
}

func TestApplyRegexInvalidExpr(t *testing.T) {
	result, err := ApplyRegex("foo", "(bar")
	assert.NotNil(t, err)
	assert.Equal(t, "foo", result)
}

func TestApplyRegexNilMatch(t *testing.T) {
	result, err := ApplyRegex("foo", "(baz)")
	assert.NotNil(t, err)
	assert.Equal(t, "foo", result)
}

func TestApplyXPath(t *testing.T) {
	page, _ := xmlpath.ParseHTML(strings.NewReader(`<html><body>test </body></html>`))
	result, err := ApplyXPath(page, `//body/text()`)
	assert.Nil(t, err)
	assert.Equal(t, "test ", result)
}

func TestApplyXPathInvalidXPath(t *testing.T) {
	page, _ := xmlpath.ParseHTML(strings.NewReader(`<html><body>test </body></html>`))
	result, err := ApplyXPath(page, `!£(*£$&!`)
	assert.NotNil(t, err)
	assert.Equal(t, "", result)
}

func TestApplyXPathNoMatch(t *testing.T) {
	page, _ := xmlpath.ParseHTML(strings.NewReader(`<html><body>test </body></html>`))
	result, err := ApplyXPath(page, `//div/text()`)
	assert.NotNil(t, err)
	assert.Equal(t, "", result)
}

func TestApplyXPathAndFilter(t *testing.T) {
	page, _ := xmlpath.ParseHTML(strings.NewReader(`<html><body>test </body></html>`))
	result, err := ApplyXPathAndFilter(page, `//body/text()`, `(.+)`)
	assert.Nil(t, err)
	assert.Equal(t, "test", result)
}

func TestApplyXPathAndFilterNoMatch(t *testing.T) {
	page, _ := xmlpath.ParseHTML(strings.NewReader(`<html><body>test </body></html>`))
	result, err := ApplyXPathAndFilter(page, `//div/text()`, `(.+)`)
	assert.NotNil(t, err)
	assert.Equal(t, "", result)
}

func TestGetNextPageURL(t *testing.T) {
	page, _ := xmlpath.ParseHTML(strings.NewReader(`<html><body><a href="/next"></a></body></html>`))
	def := &models.SiteDef{
		RefXpath: `//a/@href`,
		RefRegexp: `/(.+)`,
		URLTemplate: "http://example.com/%s",
	}
	result, err := GetNextPageURL(def, page)
	assert.Nil(t, err)
	assert.Equal(t, "http://example.com/next", result)
}

func TestGetNextPageURLNoMatch(t *testing.T) {
	page, _ := xmlpath.ParseHTML(strings.NewReader(`<html><body><a href="/next"></a></body></html>`))
	def := &models.SiteDef{
		RefXpath: `//a/@href`,
		RefRegexp: `/(.+)`,
		URLTemplate: "http://example.com/%s",
	}
	result, err := GetNextPageURL(def, page)
	assert.Nil(t, err)
	assert.Equal(t, "http://example.com/next", result)
}

func TestDecodeHTMLString(t *testing.T) {
	mojibake := `<html><body>文字化け</body></html>`
	r1 := strings.NewReader(mojibake)
	r2, err := DecodeHTMLString(r1)
	res, _ := ioutil.ReadAll(r2)
	assert.Nil(t, err)
	assert.Equal(t, mojibake, string(res))
}

func TestFetchPage(t *testing.T) {
	go func() {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html><body>test</body></html>`))
		}
		http.HandleFunc("/", testHandler)
		http.ListenAndServe("localhost:8888", http.DefaultServeMux)
	}()
	time.Sleep(100 * time.Millisecond)
	_, err := FetchPage("http://localhost:8888")
	assert.Nil(t, err)
}

func TestFetchPageGetFail(t *testing.T) {
	_, err := FetchPage("http://localhost:9999")
	assert.NotNil(t, err)
}