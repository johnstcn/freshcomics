package store

import (
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/johnstcn/freshcomics/internal/common/log"
	"github.com/johnstcn/freshcomics/internal/ipinfo"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	_ "github.com/lib/pq"
)

const (
	sqlGetComics             string = `SELECT site_defs.name, site_defs.nsfw, site_updates.id, site_updates.title, site_updates.seen_at FROM site_updates JOIN site_defs ON (site_updates.site_def_id = site_defs.id) WHERE site_updates.id IN (SELECT DISTINCT ON (site_def_id) id FROM site_updates ORDER BY site_def_id, seen_at DESC) ORDER BY seen_at desc;`
	sqlCreateSiteDef         string = `INSERT INTO site_defs (name, active, nsfw, start_url, url_template, ref_xpath, ref_regexp, title_xpath, title_regexp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id;`
	sqlRedirect              string = `SELECT site_updates.url FROM site_updates WHERE id = $1`
	sqlSaveClick             string = `INSERT INTO "comic_clicks" (update_id, country, region, city) VALUES ($1, $2, $3, $4);`
	sqlGetSiteDefs           string = `SELECT id, name, active, nsfw, start_url, url_template, ref_xpath, ref_regexp, title_xpath, title_regexp FROM site_defs ORDER BY name ASC;`
	sqlGetActiveSiteDefs     string = `SELECT id, name, active, nsfw, start_url, url_template, ref_xpath, ref_regexp, title_xpath, title_regexp FROM site_defs WHERE active = TRUE ORDER BY NAME ASC;`
	sqlGetSiteDef            string = `SELECT id, name, active, nsfw, start_url, url_template, ref_xpath, ref_regexp, title_xpath, title_regexp FROM site_defs WHERE id = $1;`
	sqlUpdateSiteDef         string = `UPDATE site_defs SET (name, active, nsfw, start_url, url_template, ref_xpath, ref_regexp, title_xpath, title_regexp) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) WHERE id = $11;`
	sqlCreateSiteUpdate      string = `INSERT INTO site_updates (site_def_id, ref, url, title, seen_at) VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	sqlGetSiteUpdates        string = `SELECT id, site_def_id, ref, url, title, seen_at FROM site_updates WHERE site_def_id = $1 ORDER BY seen_at DESC;`
	sqlGetSiteUpdate         string = `SELECT id, site_def_id, ref, url, title, seen_at FROM site_updates WHERE site_def_id = $1 AND ref = $2;`
	sqlGetLastURL            string = `SELECT url FROM site_updates WHERE site_def_id = $1 ORDER BY seen_at DESC LIMIT 1;`
	sqlGetCrawlInfos         string = `SELECT id, site_def_id, created_at, started_at, ended_at, error, seen FROM crawl_infos ORDER BY created_at DESC;`
	sqlGetCrawlInfo          string = `SELECT id, site_def_id, created_at, started_at, ended_at, error, seen FROM crawl_infos WHERE site_def_id = $1 ORDER BY created_at DESC;`
	sqlGetPendingCrawlInfos  string = `SELECT id, site_def_id, created_at, started_at, ended_at, error, seen FROM crawl_infos WHERE started_at IS NULL AND ended_at IS NULL ORDER BY created_at ASC;`
	sqlCreateCrawlInfo       string = `INSERT INTO crawl_infos (site_def_id) VALUES ($1) RETURNING ID;`
	sqlStartCrawlInfo        string = `UPDATE crawl_infos SET started_at = CURRENT_TIMESTAMP WHERE id = $1;`
	sqlEndCrawlInfo          string = `UPDATE crawl_infos SET (ended_at, error, seen) VALUES (CURRENT_TIMESTAMP, $2, $3) WHERE id = $1;`
)

type pgStore struct {
	db    Conn
	geoIP ipinfo.IPInfoer
}

var _ ComicStore = (*pgStore)(nil)
var _ Redirecter = (*pgStore)(nil)
var _ SiteDefStore = (*pgStore)(nil)
var _ SiteUpdateStore = (*pgStore)(nil)
var _ CrawlInfoStore = (*pgStore)(nil)

func NewPGStore(dsn string) (Store, error) {
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

	ip, err := ipinfo.NewIPInfoer(86400, 5)
	if err != nil {
		return nil, err
	}

	return &pgStore{db: db, geoIP: ip}, nil
}

// GetComics implements ComicStore.GetComics
func (s *pgStore) GetComics() ([]Comic, error) {
	comics := make([]Comic, 0)
	err := s.db.Select(&comics, sqlGetComics)
	if err != nil {
		fmt.Println("error fetching latest comic list:", err)
		return nil, err
	}
	return comics, nil
}

// Redirect implements Redirecter.Redirect
func (s *pgStore) Redirect(id SiteUpdateID) (string, error) {
	var result string
	err := s.db.Get(&result, sqlRedirect, id)
	if err != nil {
		return "", err
	}
	return result, nil
}

// ClickLogger interface methods

// CreateClickLog implements ClickLogger.CreateClickLog
func (s *pgStore) CreateClickLog(id SiteUpdateID, addr net.IP) error {
	geoLoc, err := s.geoIP.GetIPInfo(addr)
	if err != nil {
		return err
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(sqlSaveClick, id, geoLoc.Country, geoLoc.Region, geoLoc.City)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// CreateSiteDef implements SiteDefStore.CreateSiteDef
func (s *pgStore) CreateSiteDef(sd SiteDef) (SiteDefID, error) {
	var newid int
	tx, err := s.db.Beginx()
	if err != nil {
		return 0, err
	}
	rows, err := tx.Query(sqlCreateSiteDef, sd.Name, sd.Active, sd.NSFW, sd.StartURL, sd.URLTemplate, sd.RefXpath, sd.RefRegexp, sd.TitleXpath, sd.TitleRegexp)
	if err != nil {
		return 0, err
	}
	if rows.Next() {
		if err := rows.Scan(&newid); err != nil {
			return 0, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return SiteDefID(newid), nil
}

// GetSiteDefs implements SiteDefStore.GetSiteDefs
func (s *pgStore) GetSiteDefs(includeInactive bool) ([]SiteDef, error) {
	var err error
	defs := make([]SiteDef, 0)
	if includeInactive {
		err = s.db.Select(&defs, sqlGetSiteDefs)
	} else {
		err = s.db.Select(&defs, sqlGetActiveSiteDefs)
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return defs, nil
}

// GetSiteDef implements SiteDefStore.GetSiteDef
func (s *pgStore) GetSiteDef(id SiteDefID) (SiteDef, error) {
	def := SiteDef{}
	err := s.db.Get(&def, sqlGetSiteDef, id)
	if err != nil {
		return SiteDef{}, err
	}
	return def, nil
}

// UpdateSiteDef implements SiteDefStore.UpdateSiteDef
func (s *pgStore) UpdateSiteDef(sd SiteDef) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(sqlUpdateSiteDef, sd.Name, sd.Active, sd.NSFW, sd.StartURL, sd.URLTemplate, sd.RefXpath, sd.RefRegexp, sd.TitleXpath, sd.TitleRegexp, sd.ID)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// GetLastURL implements SiteDefStore.GetLastURL
func (s *pgStore) GetLastURL(sd SiteDef) (string, error) {
	var nextUrl string
	err := s.db.Get(&nextUrl, sqlGetLastURL, sd.ID)

	if err != nil {
		log.Info("Error fetching latest URL for SiteDef:", err)
		return "", err
	}

	return nextUrl, nil
}

// SiteUpdateStore methods

// CreateSiteUpdate implements SiteUpdateStore.CreateSiteUpdate
func (s *pgStore) CreateSiteUpdate(su SiteUpdate) (SiteUpdateID, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return 0, err
	}
	var newID int64
	rows, err := tx.Query(sqlCreateSiteUpdate, su.SiteDefID, su.Ref, su.URL, su.Title, su.SeenAt)
	if err != nil {
		return 0, err
	}
	if rows.Next() {
		if err := rows.Scan(&newID); err != nil {
			return 0, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return SiteUpdateID(newID), nil
}

// GetSiteUpdates implements SiteUpdateStore.GetSiteUpdates
func (s *pgStore) GetSiteUpdates(id SiteDefID) ([]SiteUpdate, error) {
	var err error
	updates := make([]SiteUpdate, 0)
	err = s.db.Select(&updates, sqlGetSiteUpdates, id)
	if err != nil {
		return nil, err
	}
	return updates, nil
}

// GetSiteUpdate implements SiteUpdateStore.GetSiteUpdate
func (s *pgStore) GetSiteUpdate(id SiteDefID, ref string) (SiteUpdate, error) {
	update := SiteUpdate{}
	err := s.db.Get(&update, sqlGetSiteUpdate, id, ref)
	if err != nil {
		return SiteUpdate{}, err
	}
	return update, nil
}

// CrawlInfoStore methods

// TODO(cian): limit
// GetCrawlInfos implements CrawlInfoStore.GetCrawlInfos
func (s *pgStore) GetCrawlInfos() ([]CrawlInfo, error) {
	infos := make([]CrawlInfo, 0)
	err := s.db.Select(&infos, sqlGetCrawlInfos)
	if err != nil {
		return nil, err
	}
	return infos, nil
}

// GetCrawlInfo implements CrawlInfoStore.GetCrawlInfo
func (s *pgStore) GetCrawlInfo(id SiteDefID) ([]CrawlInfo, error) {
	infos := make([]CrawlInfo, 0)
	err := s.db.Select(&infos, sqlGetCrawlInfo, id)
	if err != nil {
		return nil, err
	}
	return infos, nil
}

// GetPendingCrawlInfos implements CrawlinfoStore.GetPendingCrawlInfos
func (s *pgStore) GetPendingCrawlInfos() ([]CrawlInfo, error) {
	infos := make([]CrawlInfo, 0)
	err := s.db.Select(&infos, sqlGetPendingCrawlInfos)
	if err != nil {
		return nil, err
	}
	return infos, nil
}

// CreateCrawlInfo implements CrawlInfoStore.CreateCrawlInfo
func (s *pgStore) CreateCrawlInfo(id SiteDefID) (CrawlInfoID, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return 0, err
	}

	var newID int64
	rows, err := tx.Query(sqlCreateCrawlInfo, id)
	if err != nil {
		return 0, err
	}
	if rows.Next() {
		if err := rows.Scan(&newID); err != nil {
			return 0, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return CrawlInfoID(newID), nil
}

// StartCrawlInfo implements CrawlInfoStore.StartCrawlInfo
func (s *pgStore) StartCrawlInfo(id CrawlInfoID) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(sqlStartCrawlInfo, id)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// EndCrawlInfo implements CrawlInfoStore.EndCrawlInfo
func (s *pgStore) EndCrawlInfo(id CrawlInfoID, crawlErr error, seen int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	var errString string
	if crawlErr != nil {
		errString = crawlErr.Error()
	}

	_, err = tx.Exec(sqlEndCrawlInfo, id, errString, seen)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
