package crawler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"gopkg.in/xmlpath.v2"
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

func AlreadyExists(conn *gorm.DB, sd *SiteDef, ref string) bool {
	var c int
	conn.Model(&SiteUpdate{}).Where("site_def_id = ?", sd.ID).Where("ref = ?", ref).Count(&c)
	return c > 0
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
func GetNextPageURL(sd *SiteDef, page *xmlpath.Node) (string, error) {
	nextRef, err := ApplyXPathAndFilter(page, sd.RefXpath, sd.RefRegexp)
	if err != nil {
		return "", err
	}
	if nextRef == "" {
		return nextRef, errors.New("next page ref is empty, check xpath or filter")
	}
	nextPageUrl := fmt.Sprintf(sd.PagTemplate, nextRef)
	if nextPageUrl == "" {
		return nextPageUrl, errors.New("next page url is empty, check xpath")
	}
   	return nextPageUrl, nil
}

func GetLastCheckedSiteDef(conn *gorm.DB) (*SiteDef) {
	sd := &SiteDef{}
	now := time.Now().UTC().Unix()
	q := conn.Where("? - last_checked > 3600", now).Order("last_checked ASC").First(sd)
	if q.Error != nil {
		return nil
	}
	return sd
}

// GetLastCrawledURL returns the URL of the newest SiteUpdate of sd, if present.
func GetLastCrawledURL(conn *gorm.DB, sd *SiteDef) (string, error) {
	lastUpdate := SiteUpdate{}
	err := conn.Where("site_def_id = ?", sd.ID).Order("created_at DESC").First(&lastUpdate).Error
	if err == gorm.ErrRecordNotFound {
		return "", err
	}
	return lastUpdate.URL, nil
}

// FetchPage fetches and parses the given URL and returns the DOM.
func FetchPage(url string) (*xmlpath.Node, error) {
	log.Println("GET", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s failed", url)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	reader := strings.NewReader(string(body))
	root, err := html.Parse(reader)
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


// FetchStartPageAndURL fetches and parses the start URL for a crawl.
// For the initial crawl we use the start URL of sd.
// Subsequent crawls use the URL of the last created SiteUpdate for sd.
func FetchStartPageAndURL(conn *gorm.DB, sd *SiteDef) (*xmlpath.Node, string, error) {
	var firstCrawl bool
	lastUrl, err := GetLastCrawledURL(conn, sd)
	if err != nil {
		log.Println("initial crawl")
		lastUrl = sd.StartURL
		firstCrawl = true
	}
	lastPage, err := FetchPage(lastUrl)
	if err != nil {
		return nil, "", errors.Wrap(err, "error fetching start page")
	}

	if firstCrawl {
		return lastPage, lastUrl, nil
	}

	nextPageURL, err := GetNextPageURL(sd, lastPage)
	if err != nil {
		return nil, "", errors.Wrap(err, "error fetching next page url")
	}

	nextPage, err := FetchPage(nextPageURL)
	if err != nil {
		return nil, "", errors.Wrap(err, "error fetching next page")
	}

	return nextPage, nextPageURL, nil
}

// Crawl runs an incremental crawl of sd.
func Crawl(conn *gorm.DB, sd *SiteDef) {
	sd.LastChecked = time.Now().Unix()
	conn.Save(sd)
	start := time.Now()
	log.Printf("start crawl: %s", sd.Name)
	page, pageUrl, err := FetchStartPageAndURL(conn, sd)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("start pageUrl:", pageUrl)
	for {
		newUpdate, err := sd.NewSiteUpdateFromPage(pageUrl, page)
		if err != nil {
			log.Println("error creating new SiteUpdate from pageUrl:", err)
			break
		}

		if !AlreadyExists(conn, sd, newUpdate.Ref) {
			err = conn.Create(newUpdate).Error
			if err != nil {
				log.Println("error persisting new SiteUpdate:", err)
				break
			}
		} else {
			log.Printf("not persisting update %s - already seen", newUpdate.URL)
		}


		nextPageUrl, err := GetNextPageURL(sd, page)
		if err != nil {
			log.Println("error getting next page pageUrl:", err)
			break
		}
		if pageUrl == nextPageUrl {
			log.Println("loop detected, check next page xpath")
			break
		}
		pageUrl = nextPageUrl
		page, err = FetchPage(pageUrl)
		if err != nil {
			log.Println("error fetching next page:", err)
			break
		}
	}

	elapsed := time.Now().Sub(start)
	log.Printf("End crawl: %s (%v)", sd.Name, elapsed)
	return
}

type TestCrawlResult struct {
	Sucess bool
	Error error
	NextURL string
	Result *SiteUpdate
}

// TestCrawl runs a test of a single URL without persisting anything
func TestCrawl(sd *SiteDef) *TestCrawlResult {
	res := &TestCrawlResult{}
	page, err := FetchPage(sd.StartURL)
	if err != nil {
		res.Error = err
		return res
	}
	su, err := sd.NewSiteUpdateFromPage(sd.StartURL, page)
	res.Result = su
	if err != nil {
		res.Error = err
		return res
	}
	url, err := GetNextPageURL(sd, page)
	if err != nil {
		res.Error = err
		return res
	}
	res.NextURL = url
	return res
}