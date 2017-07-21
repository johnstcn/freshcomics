package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/azer/snakecase"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/johnstcn/freshcomics/crawler/config"
	"github.com/johnstcn/freshcomics/common/log"
)

var dao *BackendDAO

// BackendDAO wraps a DB and provides data accessor methods for models
type BackendDAO struct {
	DB *sqlx.DB
}

// CreateSiteDef creates a new SiteDef with default values
func (d *BackendDAO) CreateSiteDef() (*SiteDef, error) {
	var newid int64
	stmt := `INSERT INTO site_defs DEFAULT VALUES RETURNING id;`
	tx, err := d.DB.Beginx()
	if err != nil {
		return nil, err
	}
	err = tx.Get(&newid, stmt)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return d.GetSiteDefByID(newid)
}

// GetAllSiteDefs returns all SiteDefs.
func (d *BackendDAO) GetAllSiteDefs(includeInactive bool) (*[]SiteDef, error) {
	var stmt string
	if includeInactive {
		stmt = `SELECT * FROM site_defs ORDER BY name ASC;`
	} else {
		stmt = `SELECT * FROM site_defs WHERE active = TRUE ORDER BY NAME ASC;`
	}
	defs := make([]SiteDef, 0)
	err := d.DB.Select(&defs, stmt)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &defs, nil
}

// GetSiteDefByID returns a SiteDef with the given ID.
func (d *BackendDAO) GetSiteDefByID(id int64) (*SiteDef, error) {
	stmt := `SELECT * FROM site_defs WHERE id = $1;`
	def := SiteDef{}
	err := d.DB.Get(&def, stmt, id)
	if err != nil {
		return nil, err
	}
	return &def, nil
}

// GetSiteDefLastChecked returns the SiteDef with the oldest last_checked timestamp.
func (d *BackendDAO) GetSiteDefLastChecked() (*SiteDef, error) {
	stmt := `SELECT * FROM site_defs ORDER BY last_checked_at ASC LIMIT 1;`
	def := SiteDef{}
	err := d.DB.Get(&def, stmt)
	if err != nil {
		return nil, err
	}
	return &def, nil
}

// SaveSiteDef persists changes to a SiteDef.
func (d *BackendDAO) SaveSiteDef(sd *SiteDef) error {
	stmt := `UPDATE site_defs SET (
		name,
		active,
		nsfw,
		start_url,
		last_checked,
		url_template,
		ref_xpath,
		ref_regexp,
		title_xpath,
		title_regexp
	) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) WHERE id = $11;`
	tx, err := d.DB.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(stmt, sd.Name, sd.Active, sd.NSFW, sd.StartURL, sd.LastCheckedAt,
		sd.URLTemplate, sd.RefXpath, sd.RefRegexp, sd.TitleXpath, sd.TitleRegexp, sd.ID)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// SetSiteDefLastCheckedNow sets the last_checked of a SiteDef to the given time.
func (d *BackendDAO) SetSiteDefLastChecked(sd *SiteDef, when time.Time) error {
	stmt := `UPDATE site_defs SET last_checked_at = $1 WHERE id = $2;`
	tx, err := d.DB.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(stmt, when, sd.ID)
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
func (d *BackendDAO) CreateSiteUpdate(su *SiteUpdate) error {
	stmt := `INSERT INTO site_updates (site_def_id, ref, url, title, seen_at)
	VALUES ($1, $2, $3, $4, $5);`
	tx, err := d.DB.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(stmt, su.SiteDefID, su.Ref, su.URL, su.Title, su.SeenAt)
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
func (d *BackendDAO) GetSiteUpdatesBySiteDefID(sdid int64, limit int) (*[]SiteUpdate, error) {
	var err error
	updates := make([]SiteUpdate, 0)
	if limit < 0 {
		stmt := `SELECT * FROM site_updates WHERE site_def_id = $1 ORDER BY seen_at DESC;`
		err = d.DB.Select(&updates, stmt, sdid)
	} else {
		stmt := `SELECT * FROM site_updates WHERE site_def_id = $1 ORDER BY seen_at DESC LIMIT $2;`
		err = d.DB.Select(&updates, stmt, sdid, limit)
	}
	if err != nil {
		return nil, err
	}
	return &updates, nil
}

// GetSiteUpdateBySiteDefAndRef returns a SiteUpdate with the given ref related to the given SiteDef.
func (d *BackendDAO) GetSiteUpdateBySiteDefAndRef(sd *SiteDef, ref string) (*SiteUpdate, error) {
	update := SiteUpdate{}
	stmt := `SELECT * FROM site_updates WHERE site_def_id = $1 AND ref = $2;`
	err := d.DB.Get(&update, stmt, strconv.FormatInt(sd.ID, 10), ref)
	if err != nil {
		return nil, err
	}
	return &update, nil
}

// GetStartURLForCrawl returns the URL of the latest SiteUpdate if present,
// and nil otherwise (no SiteUpdates for SiteDef).
func (d *BackendDAO) GetStartURLForCrawl(sd *SiteDef) (string, error) {
	var nextUrl string
	stmt := `SELECT URL FROM site_updates WHERE site_def_id = $1 ORDER BY seen_at DESC LIMIT 1;`
	err := d.DB.Get(&nextUrl, stmt, sd.ID)

	if err != nil {
		log.Info("Error fetching latest URL for SiteDef:", err)
		return "", err
	}

	return nextUrl, nil
}

func (d *BackendDAO) GetCrawlEvents(limit int) (*[]CrawlEvent, error) {
	var err error
	events := make([]CrawlEvent, 0)
	if limit < 0 {
		stmt := `SELECT * FROM crawl_events ORDER BY created_at DESC;`
		err = d.DB.Select(&events, stmt)
	} else {
		stmt := `SELECT * FROM crawl_events ORDER BY created_at DESC LIMIT $1;`
		err = d.DB.Select(&events, stmt, limit)
	}
	if err != nil {
		return nil, err
	}
	return &events, nil
}

func (d *BackendDAO) GetCrawlEventsBySiteDefID(sdid int64, limit int) (*[]CrawlEvent, error) {
	var err error
	events := make([]CrawlEvent, 0)
	if limit < 0 {
		stmt := `SELECT * FROM crawl_events WHERE site_def_id = $1 ORDER BY created_at DESC;`
		err = d.DB.Select(&events, stmt, sdid)
	} else {
		stmt := `SELECT * FROM crawl_events WHERE site_def_id = $1 ORDER BY created_at DESC LIMIT $2;`
		err = d.DB.Select(&events, stmt, sdid, limit)
	}
	if err != nil {
		return nil, err
	}
	return &events, nil
}

func (d *BackendDAO) CreateCrawlEvent(sd *SiteDef, eventType, eventInfo interface{}) error {
	info, err := json.Marshal(eventInfo)
	if err != nil {
		return errors.New("eventInfo should be in format k1=v1&k2=v2")
	}
	stmt := `INSERT INTO crawl_events (site_def_id, event_type, event_info) VALUES ($1, $2, $3);`
	tx, err := d.DB.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(stmt, sd.ID, eventType, info)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func init() {
	dsn := config.Cfg.DSN
	db := sqlx.MustConnect("postgres", dsn)
	log.Info("Connected to database")
	db.MustExec(schema)
	db.MapperFunc(snakecase.SnakeCase)
	dao = &BackendDAO{DB: db}
}

func GetDAO() *BackendDAO {
	return dao
}
