package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/johnstcn/freshcomics/internal/common/log"
)

type Store interface {
	GetComics() ([]Comic, error)
	GetRedirectURL(suID string) (string, error)
	RecordClick(updateID string, addr net.IP) error
	CreateSiteDef() (*SiteDef, error)
	GetAllSiteDefs(includeInactive bool) (*[]SiteDef, error)
	SaveSiteDef(sd *SiteDef) error
	GetSiteDefByID(id int64) (*SiteDef, error)
	GetSiteDefLastChecked() (*SiteDef, error)
	SetSiteDefLastChecked(sd *SiteDef, when time.Time) error
	CreateSiteUpdate(su *SiteUpdate) error
	GetSiteUpdatesBySiteDefID(sdid int64, limit int) (*[]SiteUpdate, error)
	GetSiteUpdateBySiteDefAndRef(sd *SiteDef, ref string) (*SiteUpdate, error)
	CreateCrawlEvent(sd *SiteDef, eventType, eventInfo interface{}) error
	GetStartURLForCrawl(sd *SiteDef) (string, error)
	GetCrawlEventsBySiteDefID(sdid int64, limit int) (*[]CrawlEvent, error)
	GetCrawlEvents(limit int) (*[]CrawlEvent, error)
}

type Conn interface {
	Beginx() (*sqlx.Tx, error)
	Get(dest interface{}, query string, args ...interface{}) error
	MapperFunc(mf func(string) string)
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
	// TODO optimize this beast
	stmt := `SELECT site_defs.name, site_defs.nsfw, site_updates.id, site_updates.title, site_updates.seen_at
FROM site_updates JOIN site_defs ON (site_updates.site_def_id = site_defs.id)
WHERE site_updates.id IN (
  SELECT DISTINCT ON (site_def_id) id
  FROM site_updates
  ORDER BY site_def_id, seen_at DESC
) ORDER BY seen_at desc;`
	err := s.db.Select(&comics, stmt)
	if err != nil {
		fmt.Println("error fetching latest comic list:", err)
		return nil, err
	}
	return comics, nil
}

func (s *store) GetRedirectURL(updateID string) (string, error) {
	var result string
	stmt := `SELECT site_updates.url FROM site_updates WHERE id = $1`
	err := s.db.Get(&result, stmt, updateID)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (s *store) RecordClick(updateID string, addr net.IP) error {
	uid, err := strconv.Atoi(updateID)
	if err != nil {
		log.Error("invalid updateID:", err)
	}
	stmt := `INSERT INTO comic_clicks (update_id, country, region, city) VALUES ($1, $2, $3, $4);`

	tx, err := s.db.Beginx()
	if err != nil {
		log.Error(err)
	}
	ipinfo, err := s.geoIP.GetIPInfo(addr)
	if err != nil {
		log.Error("unable to lookup IP %s: %v", addr, err)
	}
	_, err = tx.Exec(stmt, uid, ipinfo.Country, ipinfo.Region, ipinfo.City)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// CreateSiteDef creates a new SiteDef with default values
func (s *store) CreateSiteDef() (*SiteDef, error) {
	var newid int64
	stmt := `INSERT INTO site_defs DEFAULT VALUES RETURNING id;`
	tx, err := s.db.Beginx()
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
	return s.GetSiteDefByID(newid)
}

// GetAllSiteDefs returns all SiteDefs.
func (s *store) GetAllSiteDefs(includeInactive bool) (*[]SiteDef, error) {
	var stmt string
	if includeInactive {
		stmt = `SELECT * FROM site_defs ORDER BY name ASC;`
	} else {
		stmt = `SELECT * FROM site_defs WHERE active = TRUE ORDER BY NAME ASC;`
	}
	defs := make([]SiteDef, 0)
	err := s.db.Select(&defs, stmt)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &defs, nil
}

// GetSiteDefByID returns a SiteDef with the given ID.
func (s *store) GetSiteDefByID(id int64) (*SiteDef, error) {
	stmt := `SELECT * FROM site_defs WHERE id = $1;`
	def := SiteDef{}
	err := s.db.Get(&def, stmt, id)
	if err != nil {
		return nil, err
	}
	return &def, nil
}

// GetSiteDefLastChecked returns the SiteDef with the oldest last_checked timestamp.
func (s *store) GetSiteDefLastChecked() (*SiteDef, error) {
	stmt := `SELECT * FROM site_defs ORDER BY last_checked_at ASC LIMIT 1;`
	def := SiteDef{}
	err := s.db.Get(&def, stmt)
	if err != nil {
		return nil, err
	}
	return &def, nil
}

// SaveSiteDef persists changes to a SiteDef.
func (s *store) SaveSiteDef(sd *SiteDef) error {
	stmt := `UPDATE site_defs SET (
		name,
		active,
		nsfw,
		start_url,
		last_checked_at,
		url_template,
		ref_xpath,
		ref_regexp,
		title_xpath,
		title_regexp
	) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) WHERE id = $11;`
	tx, err := s.db.Beginx()
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
func (s *store) SetSiteDefLastChecked(sd *SiteDef, when time.Time) error {
	stmt := `UPDATE site_defs SET last_checked_at = $1 WHERE id = $2;`
	tx, err := s.db.Beginx()
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
func (s *store) CreateSiteUpdate(su *SiteUpdate) error {
	stmt := `INSERT INTO site_updates (site_def_id, ref, url, title, seen_at)
	VALUES ($1, $2, $3, $4, $5);`
	tx, err := s.db.Beginx()
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
func (s *store) GetSiteUpdatesBySiteDefID(sdid int64, limit int) (*[]SiteUpdate, error) {
	var err error
	updates := make([]SiteUpdate, 0)
	if limit < 0 {
		stmt := `SELECT * FROM site_updates WHERE site_def_id = $1 ORDER BY seen_at DESC;`
		err = s.db.Select(&updates, stmt, sdid)
	} else {
		stmt := `SELECT * FROM site_updates WHERE site_def_id = $1 ORDER BY seen_at DESC LIMIT $2;`
		err = s.db.Select(&updates, stmt, sdid, limit)
	}
	if err != nil {
		return nil, err
	}
	return &updates, nil
}

// GetSiteUpdateBySiteDefAndRef returns a SiteUpdate with the given ref related to the given SiteDef.
func (s *store) GetSiteUpdateBySiteDefAndRef(sd *SiteDef, ref string) (*SiteUpdate, error) {
	update := SiteUpdate{}
	stmt := `SELECT * FROM site_updates WHERE site_def_id = $1 AND ref = $2;`
	err := s.db.Get(&update, stmt, strconv.FormatInt(sd.ID, 10), ref)
	if err != nil {
		return nil, err
	}
	return &update, nil
}

// GetStartURLForCrawl returns the URL of the latest SiteUpdate if present,
// and nil otherwise (no SiteUpdates for SiteDef).
func (s *store) GetStartURLForCrawl(sd *SiteDef) (string, error) {
	var nextUrl string
	stmt := `SELECT URL FROM site_updates WHERE site_def_id = $1 ORDER BY seen_at DESC LIMIT 1;`
	err := s.db.Get(&nextUrl, stmt, sd.ID)

	if err != nil {
		log.Info("Error fetching latest URL for SiteDef:", err)
		return "", err
	}

	return nextUrl, nil
}

func (s *store) GetCrawlEvents(limit int) (*[]CrawlEvent, error) {
	var err error
	events := make([]CrawlEvent, 0)
	if limit < 0 {
		stmt := `SELECT * FROM crawl_events ORDER BY created_at DESC;`
		err = s.db.Select(&events, stmt)
	} else {
		stmt := `SELECT * FROM crawl_events ORDER BY created_at DESC LIMIT $1;`
		err = s.db.Select(&events, stmt, limit)
	}
	if err != nil {
		return nil, err
	}
	return &events, nil
}

func (s *store) GetCrawlEventsBySiteDefID(sdid int64, limit int) (*[]CrawlEvent, error) {
	var err error
	events := make([]CrawlEvent, 0)
	if limit < 0 {
		stmt := `SELECT * FROM crawl_events WHERE site_def_id = $1 ORDER BY created_at DESC;`
		err = s.db.Select(&events, stmt, sdid)
	} else {
		stmt := `SELECT * FROM crawl_events WHERE site_def_id = $1 ORDER BY created_at DESC LIMIT $2;`
		err = s.db.Select(&events, stmt, sdid, limit)
	}
	if err != nil {
		return nil, err
	}
	return &events, nil
}

func (s *store) CreateCrawlEvent(sd *SiteDef, eventType, eventInfo interface{}) error {
	info, err := json.Marshal(eventInfo)
	if err != nil {
		return errors.New("eventInfo should be in format k1=v1&k2=v2")
	}
	stmt := `INSERT INTO crawl_events (site_def_id, event_type, event_info) VALUES ($1, $2, $3);`
	tx, err := s.db.Beginx()
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
