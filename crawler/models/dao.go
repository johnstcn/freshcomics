package models

import (
	"database/sql"
	"os"

	"github.com/azer/snakecase"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/johnstcn/freshcomics/common/log"

	"strconv"
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
	stmt := `SELECT * FROM site_defs ORDER BY last_checked ASC LIMIT 1;`
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
		title_regexp,
		date_xpath,
		date_regexp,
		date_format
	) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) WHERE id = $14;`
	tx, err := d.DB.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(stmt, sd.Name, sd.Active, sd.NSFW, sd.StartURL, sd.LastChecked,
		sd.URLTemplate, sd.RefXpath, sd.RefRegexp, sd.TitleXpath, sd.TitleRegexp,
		sd.DateXpath, sd.DateRegexp, sd.DateFormat, sd.ID)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// SetSiteDefLastCheckedNow sets the last_checked of a SiteDef to now.
func (d *BackendDAO) SetSiteDefLastCheckedNow(sd *SiteDef) error {
	stmt := `UPDATE site_defs SET last_checked = CURRENT_TIMESTAMP WHERE id = $1;`
	tx, err := d.DB.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(stmt, sd.ID)
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
	stmt := `INSERT INTO site_updates (site_def_id, ref, url, title, published)
	VALUES ($1, $2, $3, $4, $5);`
	tx, err := d.DB.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(stmt, su.SiteDefID, su.Ref, su.URL, su.Title, su.Published)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// GetSiteUpdatesBySiteDef returns all SiteUpdates related to the SiteDef.
func (d *BackendDAO) GetSiteUpdatesBySiteDef(sd *SiteDef) (*[]SiteUpdate, error) {
	updates := make([]SiteUpdate, 0)
	stmt := `SELECT * FROM site_updates WHERE site_def_id = $1;`
	err := d.DB.Select(&updates, stmt, sd.ID)
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
	stmt := `SELECT URL FROM site_updates WHERE site_def_id = $1 ORDER BY Published DESC LIMIT 1;`
	err := d.DB.Get(&nextUrl, stmt, sd.ID)

	if err != nil {
		log.Info("Error fetching latest URL for SiteDef:", err)
		return "", err
	}

	return nextUrl, nil
}

func GetDAO() *BackendDAO {
	if dao == nil {
		dsn := os.Getenv("DSN")
		db := sqlx.MustConnect("postgres", dsn)
		log.Info("Connected to database")
		db.MustExec(schema)
		db.MapperFunc(snakecase.SnakeCase)
		dao = &BackendDAO{DB: db}
	}
	return dao
}
