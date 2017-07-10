package crawler

import (
	"time"
	"fmt"
	"io/ioutil"
	"strings"
	"bytes"
	"regexp"
	"log"
	"net/http"

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
	nextPageUrl, err := ApplyXPathAndFilter(page, sd.NextPageXpath, "(.*)")
	if err != nil {
		return "", err
	}
	return nextPageUrl, nil
}

func GetLastCheckedSiteDef(conn *gorm.DB) (*SiteDef) {
	sd := &SiteDef{}
	q := conn.Order("last_checked DESC").First(sd)
	if q.Error != nil {
		return nil
	}
	if time.Now().Sub(sd.LastChecked) > sd.CheckInterval {
		return sd
	}
	return nil
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
	sd.LastChecked = time.Now()
	conn.Save(sd)
	start := time.Now()
	log.Printf("start crawl: %s", sd.Name)
	page, url, err := FetchStartPageAndURL(conn, sd)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("start url:", url)
	for {
		newUpdate, err := sd.NewSiteUpdateFromPage(url, page)
		if err != nil {
			log.Println("error creating new SiteUpdate from url:", err)
			break
		}

		err = conn.Create(newUpdate).Error
		if err != nil {
			log.Println("error persisting new SiteUpdate:", err)
			break
		}

		url, err = GetNextPageURL(sd, page)
		if err != nil {
			log.Println("error getting next page url:", err)
			break
		}
		page, err = FetchPage(url)
		if err != nil {
			log.Println("error fetching next page:", err)
			break
		}
	}

	elapsed := time.Now().Sub(start)
	log.Printf("End crawl: %s (%v)", sd.Name, elapsed)
	return
}

