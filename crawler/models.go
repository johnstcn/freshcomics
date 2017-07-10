package crawler

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"gopkg.in/xmlpath.v2"
)

const (
	DEFAULT_CHECK_INTERVAL = 60 * time.Minute
)

type SiteDef struct {
	gorm.Model
	Name          string		 `sql:"unique;index"`
	StartURL      string
	LastChecked   time.Time		 `sql:"index"`
	CheckInterval time.Duration
	RefRegexp     string
	NextPageXpath string
	TitleXpath    string
	TitleRegexp   string
	DateXpath     string
	DateRegexp    string
	DateFormat    string
}

type SiteUpdate struct {
	gorm.Model
	SiteDefID uint
	Ref       string	`sql:"unique;index"`
	URL       string
	Title     string
	Published time.Time `sql:"index"`
}

func NewSiteDef(name, startURL, refRegexp, nextPageXpath, titleXpath, titleRegexp, dateXpath, dateRegexp, dateFormat string) (*SiteDef, error) {
	if name == "" {
		return nil, errors.New("SiteDef name can't be empty")
	}
	if startURL == "" {
		return nil, errors.New("SiteDef start url can't be empty")
	}
	if refRegexp == "" {
		return nil, errors.New("SiteDef ref regexp can't be empty")
	}
	if nextPageXpath == "" {
		return nil, errors.New("SiteDef next page xpath can't be empty")
	}
	if titleXpath == "" {
		titleXpath = "//title/text()"
	}
	if titleRegexp == "" {
		titleRegexp = "(.*)"
	}

	sd := &SiteDef{
		Name: name,
		StartURL: startURL,
		LastChecked: time.Unix(0, 0),
		CheckInterval: DEFAULT_CHECK_INTERVAL,
		RefRegexp: refRegexp,
		NextPageXpath: nextPageXpath,
		TitleXpath: titleXpath,
		TitleRegexp: titleRegexp,
		DateXpath: dateXpath,
		DateRegexp: dateRegexp,
		DateFormat: dateFormat,
	}
	return sd, nil
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
		Published: published,
	}
	return su, nil
}