package store

import (
	"github.com/jmoiron/sqlx"

	_ "github.com/golang/mock/mockgen/model"
)

//go:generate mockgen -destination mocks/store.go . Store

type Store interface {
	ComicStore
	Redirecter
	SiteDefStore
	SiteUpdateStore
	CrawlInfoStore
}

type ComicStore interface {
	// GetComics returns the latest comics
	GetComics() ([]Comic, error)
}

type Redirecter interface {
	// Redirect returns the URL for the given SiteUpdateID
	Redirect(id SiteUpdateID) (string, error)
}

type SiteDefStore interface {
	// GetSiteDefs returns all active SiteDefs. If includeInactive is true, returns all SiteDefs.
	GetSiteDefs(includeInactive bool) ([]SiteDef, error)
	// GetSiteDef returns the SiteDef with the given SiteDefID
	GetSiteDef(id SiteDefID) (SiteDef, error)
	// CreateSiteDef persists the given SiteDef returning the id
	CreateSiteDef(sd SiteDef) (SiteDefID, error)
	// UpdateSiteDef updates the given SiteDef
	UpdateSiteDef(sd SiteDef) error
	// GetLastURL returns the last URL seen for the given SiteDef.
	GetLastURL(id SiteDefID) (string, error)
}

type SiteUpdateStore interface {
	// CreateSiteUpdate persists the given SiteUpdate returning the id
	CreateSiteUpdate(su SiteUpdate) (SiteUpdateID, error)
	// GetSiteUpdates returns all SiteUpdates for the given SiteDefID
	GetSiteUpdates(id SiteDefID) ([]SiteUpdate, error)
	// GetSiteUpdate gets a single SiteUpdate from the SiteDefID and the ref
	GetSiteUpdate(id SiteDefID, ref string) (SiteUpdate, bool, error)
}

type CrawlInfoStore interface {
	// GetCrawlInfos returns all CrawlInfos
	GetCrawlInfos() ([]CrawlInfo, error)
	// GetCrawlInfo returns all CrawlInfos for the given SiteDefID
	GetCrawlInfo(id SiteDefID) ([]CrawlInfo, error)
	// GetPendingCrawlInfos returns all CrawlInfos where started_at and ended_at is null
	GetPendingCrawlInfos() ([]CrawlInfo, error)
	// CreateCrawlInfo creates a new CrawlInfo for the given SiteDefID and url with default fields returning the id
	CreateCrawlInfo(id SiteDefID, url string) (CrawlInfoID, error)
	// StartCrawlInfo sets started_at to the current time for the given CrawlInfoID
	StartCrawlInfo(id CrawlInfoID) error
	// EndCrawlInfo sets ended_at to the current timestamp for the given CrawlInfoID and sets error and seen to the given values
	EndCrawlInfo(id CrawlInfoID, crawlErr error, seen int) error
}

type Conn interface {
	Beginx() (*sqlx.Tx, error)
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
}

var _ Conn = (*sqlx.DB)(nil)
