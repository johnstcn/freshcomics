package store

import (
	"net"
	"time"

	"github.com/jmoiron/sqlx"
)

//go:generate mockery -interface Store -package storetest

type Store interface {
	GetComics() ([]Comic, error)
	GetRedirectURL(suID string) (string, error)
	RecordClick(updateID int, addr net.IP) error
	CreateSiteDef(SiteDef) (int, error)
	GetAllSiteDefs(includeInactive bool) ([]SiteDef, error)
	SaveSiteDef(sd SiteDef) error
	GetSiteDefByID(id int64) (SiteDef, error)
	GetSiteDefLastChecked() (SiteDef, error)
	SetSiteDefLastChecked(sd SiteDef, when time.Time) error
	CreateSiteUpdate(su SiteUpdate) error
	GetSiteUpdatesBySiteDefID(sdid int64) ([]SiteUpdate, error)
	GetSiteUpdateBySiteDefAndRef(sdid int64, ref string) (SiteUpdate, error)
	GetStartURLForCrawl(sd SiteDef) (string, error)
	GetCrawlInfoBySiteDefID(sdid int64) ([]CrawlInfo, error)
	GetCrawlInfo() ([]CrawlInfo, error)
	CreateCrawlInfo(ci CrawlInfo) error
}

type Conn interface {
	Beginx() (*sqlx.Tx, error)
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
}

var _ Conn = (*sqlx.DB)(nil)
