package dao

import (
	"database/sql"
	"github.com/lib/pq"
	
	"github.com/johnstcn/freshcomics/backendsvc/models"
)

const (
	SiteDefSchema = `CREATE TABLE IF NOT EXISTS site_defs (
	id				serial	 	PRIMARY KEY,
	name 			text 		NOT NULL DEFAULT 'New SiteDef' UNIQUE,
	active			boolean		NOT NULL DEFAULT FALSE,
	nsfw			boolean		NOT NULL DEFAULT FALSE,
	start_url 		text 		NOT NULL DEFAULT 'http://example.com' UNIQUE,
	last_checked_at timestamptz	NOT NULL DEFAULT CURRENT_TIMESTAMP,
	url_template 	text 		NOT NULL DEFAULT 'http://example.com/%s' UNIQUE,
	ref_xpath 		text 		NOT NULL DEFAULT '//a[@rel="next"]/@href',
	ref_regexp 		text 		NOT NULL DEFAULT '([^/]+)/?$',
	title_xpath 	text 		NOT NULL DEFAULT '//title/text()',
	title_regexp 	text 		NOT NULL DEFAULT '(.+)'
);

CREATE INDEX IF NOT EXISTS site_defs_last_checked_at_idx ON site_defs (last_checked_at);`

	SiteDefCreateStmt = `INSERT INTO site_defs DEFAULT VALUES RETURNING id;`
	SiteDefGetStmt    = `SELECT * FROM site_defs WHERE id = $1;`
	SiteDefGetAllStmt = `SELECT * FROM site_defs ORDER BY last_checked_at ASC;`
	SiteDefUpdateStmt = `UPDATE site_defs SET (
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
)



// SiteDefManager abstracts CRUD interactions with SiteDefs.
type SiteDefManager interface {
	Create() (int64, error)
	Get(id int64) (*models.SiteDef, error)
	GetAll() (*[]models.SiteDef, error)
	Update(*models.SiteDef) error
}

// SiteDefStore handles CRUD operations for SiteDefs and is backed by a PostgreSQL database.
// It implements the SiteDefManager interface.
type SiteDefStore struct {
	db *sql.DB
}

// Create creates a new SiteDef with default values and returns its id.
func (s SiteDefStore) Create() (int64, error) {
	var newID int64
	tx, err := s.db.Begin()
	if err != nil {
		return -1, err
	}
	row := tx.QueryRow(SiteDefCreateStmt)
	err = row.Scan(&newID)
	if err != nil {
		return -1, err
	}
	err = tx.Commit()
	if err != nil {
		return -1, err
	}
	return newID, nil
}


// Get returns a SiteDef with the given id.
func (s SiteDefStore) Get(id int64) (*models.SiteDef, error) {
	var nt pq.NullTime	// database/sql can't Scan() time.Time directly.
	def := models.SiteDef{}
	row := s.db.QueryRow(SiteDefGetStmt, id)
	err := row.Scan(
		&def.ID,
		&def.Name,
		&def.Active,
		&def.NSFW,
		&def.StartURL,
		&nt,
		&def.URLTemplate,
		&def.RefXpath,
		&def.RefRegexp,
		&def.TitleXpath,
		&def.TitleRegexp,
	)
	if err != nil {
		return nil, err
	}
	def.LastCheckedAt = nt.Time
	return &def, nil
}


// GetAll returns all SiteDefs ordered by LastCheckedAt ascending.
func (s SiteDefStore) GetAll() (*[]models.SiteDef, error) {
	defs := make([]models.SiteDef, 0)
	rows, err := s.db.Query(SiteDefGetAllStmt)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var nt pq.NullTime
		def := models.SiteDef{}
		err := rows.Scan(
			&def.ID,
			&def.Name,
			&def.Active,
			&def.NSFW,
			&def.StartURL,
			&nt,
			&def.URLTemplate,
			&def.RefXpath,
			&def.RefRegexp,
			&def.TitleXpath,
			&def.TitleRegexp,
		)
		def.LastCheckedAt = nt.Time
		defs = append(defs, def)
		if err != nil {
			return nil, err
		}
	}
	return &defs, nil
}

// Update persists changes to a given SiteDef.
func (s SiteDefStore) Update(sitedef *models.SiteDef) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		SiteDefUpdateStmt,
		sitedef.Name,
		sitedef.Active,
		sitedef.NSFW,
		sitedef.StartURL,
		sitedef.LastCheckedAt,
		sitedef.URLTemplate,
		sitedef.RefXpath,
		sitedef.RefRegexp,
		sitedef.TitleXpath,
		sitedef.TitleRegexp,
		sitedef.ID,
	)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}


// NewSiteDefStore creates a new SiteDefStore.
func NewSiteDefStore(f func() *sql.DB) (*SiteDefStore, error) {
	store := &SiteDefStore{db: f()}
	_, err := store.db.Exec(SiteDefSchema)
	if err != nil {
		return nil, err
	}
	return store, nil
}
