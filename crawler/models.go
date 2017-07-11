package crawler

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"gopkg.in/xmlpath.v2"
)

type SiteUpdate struct {
	gorm.Model
	SiteDefID 	uint
	Ref       	string	`sql:"index"`
	URL       	string
	Title     	string
	Published   int64 	`sql:"index"`
}

type SiteDef struct {
	gorm.Model
	Name        string	`sql:"unique;index"`
	StartURL    string
	LastChecked int64	`sql:"index"`
	PagTemplate string
	RefXpath    string
	RefRegexp   string
	TitleXpath  string
	TitleRegexp string
	DateXpath   string
	DateRegexp  string
	DateFormat  string
}

// NewSiteDef is a convenience method for creating a new SiteDef
func NewSiteDef(name, startURL, pagTemplate, refXpath, refRegexp, titleXpath, titleRegexp, dateXpath, dateRegexp, dateFormat string) (*SiteDef, error) {
	if name == "" {
		return nil, errors.New("SiteDef name can't be empty")
	}
	if startURL == "" {
		return nil, errors.New("SiteDef start url can't be empty")
	}
	if refRegexp == "" {
		return nil, errors.New("SiteDef ref regexp can't be empty")
	}
	if pagTemplate == "" {
		return nil, errors.New("SiteDef next page xpath can't be empty")
	}
	if titleXpath == "" {
		titleXpath = "//title/text()"
	}
	if titleRegexp == "" {
		titleRegexp = "(.*)"
	}

	sd := &SiteDef{
		Name:        name,
		StartURL:    startURL,
		LastChecked: 0,
		PagTemplate: pagTemplate,
		RefXpath:    refXpath,
		RefRegexp:   refRegexp,
		TitleXpath:  titleXpath,
		TitleRegexp: titleRegexp,
		DateXpath:   dateXpath,
		DateRegexp:  dateRegexp,
		DateFormat:  dateFormat,
	}
	return sd, nil
}

func (sd *SiteDef) SetLastChecked(value string) {
	t , _ := time.Parse("2006-01-02T15:04:05", value)
	sd.LastChecked = t.Unix()
}

// GetRefFromURL runs the RefRegexp of sd on url and returns the result.
func (sd *SiteDef) GetRefFromURL(url string) (string, error) {
	filtered, err := ApplyRegex(url, sd.RefRegexp)
	if err != nil {
		return url, errors.Wrap(err, "ref regex did not match")
	}
	return filtered, nil
}

// NewSiteUpdateFromPage attempts to create a new SiteUpdate from the given URL. Does not persist the new item.
func (sd *SiteDef) NewSiteUpdateFromPage(url string, page *xmlpath.Node) (*SiteUpdate, error) {
	ref, err := sd.GetRefFromURL(url)
	if err != nil {
		return nil, errors.Wrap(err, "error extracting ref from url")
	}

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

	su := &SiteUpdate{
		SiteDefID: sd.ID,
		Ref: ref,
		URL: url,
		Title: title,
		Published: published.Unix(),
	}
	return su, nil
}