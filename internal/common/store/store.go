package store

import (
	"database/sql"
	"fmt"
	"net"
	"time"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/johnstcn/freshcomics/internal/common/log"
)

const (
	sqlGetComics = `SELECT site_defs.name, site_defs.nsfw, site_updates.id, site_updates.title, site_updates.seen_at FROM site_updates JOIN site_defs ON (site_updates.site_def_id = site_defs.id) WHERE site_updates.id IN (SELECT DISTINCT ON (site_def_id) id FROM site_updates ORDER BY site_def_id, seen_at DESC) ORDER BY seen_at desc;`
	sqlCreateSiteDef = `INSERT INTO site_defs (name, active, nsfw, start_url, last_checked_at, url_template, ref_xpath, ref_regexp, title_xpath, title_regexp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id;`
	sqlGetRedirectURL = `SELECT site_updates.url FROM site_updates WHERE id = $1`
	sqlRecordClick = `INSERT INTO comic_clicks (update_id, country, region, city) VALUES ($1, $2, $3, $4);`
	sqlGetAllSiteDefs = `SELECT id, name, active, nsfw, start_url, last_checked_at, url_template, ref_xpath, ref_regexp, title_xpath, title_regexp FROM site_defs ORDER BY name ASC;`
	sqlGetAllSiteDefsActive = `SELECT id, name, active, nsfw, start_url, last_checked_at, url_template, ref_xpath, ref_regexp, title_xpath, title_regexp FROM site_defs WHERE active = TRUE ORDER BY NAME ASC;`
	sqlGetSiteDefByID = `SELECT id, name, active, nsfw, start_url, last_checked_at, url_template, ref_xpath, ref_regexp, title_xpath, title_regexp FROM site_defs WHERE id = $1;`
	sqlGetSiteDefLastChecked = `SELECT id, name, active, nsfw, start_url, last_checked_at, url_template, ref_xpath, ref_regexp, title_xpath, title_regexp  FROM site_defs ORDER BY last_checked_at ASC LIMIT 1;`
	sqlSaveSiteDef = `UPDATE site_defs SET (name, active, nsfw, start_url, last_checked_at, url_template, ref_xpath, ref_regexp, title_xpath, title_regexp) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) WHERE id = $11;`
	sqlSetSiteDefLastChecked = `UPDATE site_defs SET last_checked_at = $1 WHERE id = $2;`
	sqlCreateSiteUpdate               = `INSERT INTO site_updates (site_def_id, ref, url, title, seen_at) VALUES ($1, $2, $3, $4, $5);`
	sqlGetSiteUpdatesBySiteDefID      = `SELECT id, site_def_id, ref, url, title, seen_at FROM site_updates WHERE site_def_id = $1 ORDER BY seen_at DESC;`
	sqlGetSiteUpdateBySiteDefAndRef   = `SELECT id, site_def_id, ref, url, title, seen_at FROM site_updates WHERE site_def_id = $1 AND ref = $2;`
	sqlGetStartURLForCrawl            = `SELECT url FROM site_updates WHERE site_def_id = $1 ORDER BY seen_at DESC LIMIT 1;`
	sqlGetCrawlInfo                   = `SELECT id, site_def_id, started_at, ended_at, status, seen FROM crawl_events ORDER BY created_at DESC;`
	sqlGetCrawlInfoBySiteDefID        = `SELECT id, site_def_id, started_at, ended_at, status, seen FROM crawl_events WHERE site_def_id = $1 ORDER BY created_at DESC;`
	sqlCreateCrawlInfo                = `INSERT INTO crawl_events (site_def_id, started_at, ended_at, status, seen) VALUES ($1, $2, $3, $4, $5);`

)

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

type store struct {
	db    Conn
	geoIP IPInfoer
}

var _ Store = (*store)(nil)

func NewStore(dsn string) (Store, error) {
	var db Conn
	var err error
	for {
		db, err = sqlx.Connect("postgres", dsn)
		if err != nil {
			log.Info(errors.Wrapf(err, "could not connect to database"))
			<-time.After(1 * time.Second)
			continue
		}
		log.Info("Connected to database")
		break
	}
	log.Info("Connected to database")

	ip, err := NewIPInfoer(86400, 5)
	if err != nil {
		return nil, err
	}

	return &store{db: db, geoIP: ip}, nil
}

func (s *store) GetComics() ([]Comic, error) {
	comics := make([]Comic, 0)
	err := s.db.Select(&comics, sqlGetComics)
	if err != nil {
		fmt.Println("error fetching latest comic list:", err)
		return nil, err
	}
	return comics, nil
}

func (s *store) GetRedirectURL(updateID string) (string, error) {
	var result string
	err := s.db.Get(&result, sqlGetRedirectURL, updateID)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (s *store) RecordClick(updateID int, addr net.IP) error {
	geoLoc, err := s.geoIP.GetIPInfo(addr)
	if err != nil {
		return err
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(sqlRecordClick, updateID, geoLoc.Country, geoLoc.Region, geoLoc.City)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// CreateSiteDef persists the given SiteDef returning the id
func (s *store) CreateSiteDef(sd SiteDef) (int, error) {
	var newid int
	tx, err := s.db.Beginx()
	if err != nil {
		return -1, err
	}
	rows, err := tx.Query(sqlCreateSiteDef, sd.Name, sd.Active, sd.NSFW, sd.StartURL, sd.LastCheckedAt, sd.URLTemplate, sd.RefXpath, sd.RefRegexp, sd.TitleXpath, sd.TitleRegexp)
	if err != nil {
		return -1, err
	}
	if rows.Next() {
		rows.Scan(&newid)
	}
	err = tx.Commit()
	if err != nil {
		return -1, err
	}
	return newid, nil
}

// GetAllSiteDefs returns all SiteDefs.
func (s *store) GetAllSiteDefs(includeInactive bool) ([]SiteDef, error) {
	var err error
	defs := make([]SiteDef, 0)
	if includeInactive {
		err = s.db.Select(&defs, sqlGetAllSiteDefs)
	} else {
		err = s.db.Select(&defs, sqlGetAllSiteDefsActive)
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return defs, nil
}

// GetSiteDefByID returns a SiteDef with the given ID.
func (s *store) GetSiteDefByID(id int64) (SiteDef, error) {
	def := SiteDef{}
	err := s.db.Get(&def, sqlGetSiteDefByID, id)
	if err != nil {
		return SiteDef{}, err
	}
	return def, nil
}

// GetSiteDefLastChecked returns the SiteDef with the oldest last_checked timestamp.
func (s *store) GetSiteDefLastChecked() (SiteDef, error) {
	def := SiteDef{}
	err := s.db.Get(&def, sqlGetSiteDefLastChecked)
	if err != nil {
		return SiteDef{}, err
	}
	return def, nil
}

// SaveSiteDef persists changes to a SiteDef.
func (s *store) SaveSiteDef(sd SiteDef) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(sqlSaveSiteDef, sd.Name, sd.Active, sd.NSFW, sd.StartURL, sd.LastCheckedAt, sd.URLTemplate, sd.RefXpath, sd.RefRegexp, sd.TitleXpath, sd.TitleRegexp, sd.ID)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// SetSiteDefLastChecked sets the last_checked of a SiteDef to the given time.
func (s *store) SetSiteDefLastChecked(sd SiteDef, when time.Time) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(sqlSetSiteDefLastChecked, when, sd.ID)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// CreateSiteUpdate persists a new SiteUpdate.
func (s *store) CreateSiteUpdate(su SiteUpdate) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(sqlCreateSiteUpdate, su.SiteDefID, su.Ref, su.URL, su.Title, su.SeenAt)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// GetSiteUpdatesBySiteDefID returns all SiteUpdates related to the SiteDef.
func (s *store) GetSiteUpdatesBySiteDefID(sdid int64) ([]SiteUpdate, error) {
	var err error
	updates := make([]SiteUpdate, 0)
	err = s.db.Select(&updates, sqlGetSiteUpdatesBySiteDefID, sdid)
	if err != nil {
		return nil, err
	}
	return updates, nil
}

// GetSiteUpdateBySiteDefAndRef returns a SiteUpdate with the given ref related to the given SiteDef.
func (s *store) GetSiteUpdateBySiteDefAndRef(sdid int64, ref string) (SiteUpdate, error) {
	update := SiteUpdate{}
	err := s.db.Get(&update, sqlGetSiteUpdateBySiteDefAndRef, sdid, ref)
	if err != nil {
		return SiteUpdate{}, err
	}
	return update, nil
}

// GetStartURLForCrawl returns the URL of the latest SiteUpdate if present,
// and nil otherwise (no SiteUpdates for SiteDef).
func (s *store) GetStartURLForCrawl(sd SiteDef) (string, error) {
	var nextUrl string
	err := s.db.Get(&nextUrl, sqlGetStartURLForCrawl, sd.ID)

	if err != nil {
		log.Info("Error fetching latest URL for SiteDef:", err)
		return "", err
	}

	return nextUrl, nil
}

func (s *store) GetCrawlInfo() ([]CrawlInfo, error) {
	events := make([]CrawlInfo, 0)
	err := s.db.Select(&events, sqlGetCrawlInfo)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (s *store) GetCrawlInfoBySiteDefID(sdid int64) ([]CrawlInfo, error) {
	events := make([]CrawlInfo, 0)
	err := s.db.Select(&events, sqlGetCrawlInfoBySiteDefID, sdid)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (s *store) CreateCrawlInfo(ci CrawlInfo) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(sqlCreateCrawlInfo, ci.SiteDefID, ci.StartedAt, ci.EndedAt, ci.Status, ci.Seen)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
