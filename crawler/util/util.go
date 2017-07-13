package util

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"gopkg.in/xmlpath.v2"

	"github.com/johnstcn/freshcomics/crawler/log"
	"github.com/johnstcn/freshcomics/crawler/models"
)

// ApplyRegex returns the first match of expr in input if present.
// If not present, returns input.
func ApplyRegex(input, expr string) (string, error) {
	regex, err := regexp.Compile(expr)
	if err != nil {
		return input, errors.Wrap(err, "invalid regexp:")
	}

	match := regex.FindStringSubmatch(input)
	if match == nil {
		return input, errors.New(fmt.Sprintf("input %s does not match regexp %s", input, expr))
	}
	return match[1], nil
}

// ApplyXPath evaluates the result of xpath in the context of page, returning an empty string and an error if no match.
func ApplyXPath(page *xmlpath.Node, xpath string) (string, error) {
	xp, err := xmlpath.Compile(xpath)
	if err != nil {
		return "", err
	}

	value, ok := xp.String(page)
	if !ok {
		return "", errors.New(fmt.Sprintf("no match for xpath: %s", xpath))
	}
	return value, nil
}

// ApplyXPathAndFilter evaluates the result of xpath in the context of page and returns the first match of regex to the result.
// If no match for xpath, returns an empty string an an error.
// If no match for regex, returns the original result of evaluating the xpath.
func ApplyXPathAndFilter(page *xmlpath.Node, xpath string, regex string) (string, error) {
	value, err := ApplyXPath(page, xpath)
	if err != nil {
		return "", err
	}
	return ApplyRegex(value, regex)
}

// GetNextPageURL evaluates NextPageXpath of sd in the context of page.
func GetNextPageURL(sd *models.SiteDef, page *xmlpath.Node) (string, error) {
	nextRef, err := ApplyXPathAndFilter(page, sd.NextPageXpath, sd.NextPageRegexp)
	if err != nil {
		return "", err
	}
	if nextRef == "" {
		return nextRef, errors.New("next page ref is empty, check xpath or filter")
	}
	nextPageUrl := fmt.Sprintf(sd.URLTemplate, nextRef)
	if nextPageUrl == "" {
		return nextPageUrl, errors.New("next page url is empty, check xpath")
	}
   	return nextPageUrl, nil
}

func DecodeHTMLString(r io.Reader) (io.Reader, error) {
	utfRdr, err := charset.NewReader(r, "text/html")
	if err != nil {
		return nil, err
	}
	return utfRdr, nil
}

// FetchPage fetches and parses the given URL and returns the DOM.
func FetchPage(url string) (*xmlpath.Node, error) {
	log.Info.Println("GET", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s failed", url)
	}
	defer resp.Body.Close()

	rdr, err := DecodeHTMLString(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert response to utf8")
	}

	root, err := html.Parse(rdr)
	if err != nil {
		return nil, errors.Wrap(err, "html failed to parse")
	}

	var b bytes.Buffer
	html.Render(&b, root)
	fixedHtml := b.String()
	xmlReader := strings.NewReader(fixedHtml)
	xmlRoot, err := xmlpath.ParseHTML(xmlReader)
	if err != nil {
		return nil, errors.Wrap(err,"xmlpath failed to parse html")
	}
	return xmlRoot, nil
}

// Crawl runs an incremental crawl of sd.
func Crawl(sd *models.SiteDef, fullCrawl bool) {
	dao := models.GetDAO()
	dao.SetSiteDefLastCheckedNow(sd)
	start := time.Now()
	log.Info.Printf("start crawl: %s", sd.Name)
	page, err := FetchPage(sd.StartURL)
	pageUrl := sd.StartURL
	if err != nil {
		log.Error.Println(err)
		return
	}
	for {
		newUpdate, err := NewSiteUpdateFromPage(sd, page)
		if err != nil {
			log.Error.Println("error creating new SiteUpdate from pageUrl:", err)
			break
		}

		existingUpdate, err := dao.GetSiteUpdateBySiteDefAndRef(sd, newUpdate.Ref)
		if existingUpdate == nil {
			// does not exist -> persist
			newUpdate, err := NewSiteUpdateFromPage(sd, page)
			if err != nil {
				log.Error.Println(err)
				break
			}
			err = dao.CreateSiteUpdate(newUpdate)
			if err != nil {
				log.Error.Println(err)
				break
			}
		} else if fullCrawl {
			// exists, and we want to update -> update
			log.Info.Printf("saw existing update %s and full crawl - updating", newUpdate.Ref)
			newUpdate.ID = existingUpdate.ID
			err := dao.CreateSiteUpdate(newUpdate)
			if err != nil {
				log.Error.Println(err)
			}
		} else {
			log.Info.Printf("saw existing update %s and not full crawl - stopping", newUpdate.URL)
			// if we are not doing a full crawl
			break
		}

		// pagination
		nextPageUrl, err := GetNextPageURL(sd, page)
		if err != nil {
			log.Error.Println(err)
			break
		}
		if pageUrl == nextPageUrl {
			log.Warn.Println("loop detected, check next page xpath")
			break
		}
		pageUrl = nextPageUrl
		page, err = FetchPage(pageUrl)
		if err != nil {
			log.Error.Println("error during pagination:", err)
			break
		}
	}

	elapsed := time.Now().Sub(start)
	log.Info.Printf("End crawl: %s (%v)", sd.Name, elapsed)
	return
}

type TestCrawlResult struct {
	Error string
	NextURL string
	Result *models.SiteUpdate
}

// TestCrawl runs a test of a single URL without persisting anything
func TestCrawl(sd *models.SiteDef) *TestCrawlResult {
	res := &TestCrawlResult{}
	page, err := FetchPage(sd.StartURL)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	su, err := NewSiteUpdateFromPage(sd, page)
	res.Result = su
	if err != nil {
		res.Error = err.Error()
		return res
	}
	url, err := GetNextPageURL(sd, page)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	res.NextURL = url
	return res
}

// NewSiteUpdateFromPage attempts to create a new SiteUpdate from the given URL. Does not persist the new item.
func NewSiteUpdateFromPage(sd *models.SiteDef, page *xmlpath.Node) (*models.SiteUpdate, error) {
	ref, err := ApplyXPathAndFilter(page, sd.RefXpath, sd.RefRegexp)
	if err != nil {
		return nil, errors.Wrap(err, "error extracting ref from url")
	}

	url := fmt.Sprintf(sd.URLTemplate, ref)

	title, err := ApplyXPathAndFilter(page, sd.TitleXpath, sd.TitleRegexp)
	if err != nil {
		return nil, errors.Wrap(err,"error extracting title from page")
	}

	published := time.Now()
	if sd.DateXpath != "" && sd.DateRegexp != "" && sd.DateFormat != "" {
		publishedRaw, err := ApplyXPathAndFilter(page, sd.DateXpath, sd.DateRegexp)
		if err != nil {
			return nil, errors.Wrap(err,"error extracting date from page")
		}
		published, err = time.Parse(sd.DateFormat, publishedRaw)
		if err != nil {
			return nil, errors.Wrap(err,"error parsing date from page")
		}
	}

	su := &models.SiteUpdate{
		SiteDefID: sd.ID,
		Ref: ref,
		URL: url,
		Title: title,
		Published: published,
	}
	return su, nil
}