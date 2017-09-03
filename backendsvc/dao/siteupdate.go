package dao

import (
	"database/sql"
	"github.com/lib/pq"
	"github.com/johnstcn/freshcomics/backendsvc/models"
	"time"
)

const (
	SiteUpdateSchema = `CREATE TABLE IF NOT EXISTS site_updates (
	id 				serial		PRIMARY KEY,
	site_def_id 	integer 	REFERENCES site_defs (id) ON DELETE CASCADE,
	ref 			text		NOT NULL,
	url				text		NOT NULL UNIQUE,
	title			text		NOT NULL,
	seen_at			timestamptz	NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS site_updates_ref_idx ON site_updates (ref);
CREATE INDEX IF NOT EXISTS site_updates_seen_at_idx ON site_updates(seen_at);`
	SiteUpdateCreateStmt = `INSERT INTO site_updates (site_def_id, ref, url, title, seen_at)
	VALUES ($1, $2, $3, $4, $5);`
	SiteUpdateGetStmt = `SELECT * FROM site_updates WHERE id = $1;`
	SiteUpdateGetAllBySiteDefIDStmt = `SELECT * FROM site_updates WHERE site_def_id = $1 ORDER BY seen_at DESC;`
	SiteUpdateUpdateStmt = `UPDATE site_updates SET (ref, url, title, seen_at) = ($1, $2, $3, $4) WHERE id = $5;`
)

// SiteUpdateManager abstracts CRUD interactions with SiteUpdates.
type SiteUpdateManager interface {
	Create() (int64, error)
	Get(id int64) (*models.SiteUpdate, error)
	GetAllBySiteDefID(siteDefID int64) (*[]models.SiteUpdate, error)
	Update(*models.SiteUpdate) error
}

// SiteUpdateStore handles CRUD operations for SiteUpdates and implements SiteUpdateManager interface.
// Backed by a PostgreSQL database.
type SiteUpdateStore struct {
	db *sql.DB
}

// Create creates a new SiteUpdate and returns its id.
func (s SiteUpdateStore) Create(siteDefId int64, ref, url, title string, seen_at time.Time) (int64, error) {
	var newID int64
	tx, err := s.db.Begin()
	if err != nil {
		return -1, err
	}
	row := tx.QueryRow(SiteUpdateCreateStmt, siteDefId, ref, url, title, seen_at)
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

// Get retrieves a given SiteUpdate by id.
func (s SiteUpdateStore) Get(id int64) (*models.SiteUpdate, error) {
	var nt pq.NullTime	// database/sql can't Scan() time.Time directly.
	u := models.SiteUpdate{}
	row := s.db.QueryRow(SiteUpdateGetStmt, id)
	err := row.Scan(
		&u.ID,
		&u.SiteDefID,
		&u.Ref,
		&u.URL,
		&u.Title,
		&nt,
	)
	if err != nil {
		return nil, err
	}
	u.SeenAt = nt.Time
	return &u, nil
}

// GetAllBySiteDefID returns all SiteUpdates for a given SiteDef.
func (s SiteUpdateStore) GetAllBySiteDefID(siteDefID int64) (*[]models.SiteUpdate, error) {
	us := make([]models.SiteUpdate, 0)
	rows, err := s.db.Query(SiteUpdateGetAllBySiteDefIDStmt)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var nt pq.NullTime
		u := models.SiteUpdate{}
		err := rows.Scan(
			&u.ID,
			&u.SiteDefID,
			&u.Ref,
			&u.URL,
			&u.Title,
			&nt,
		)
		u.SeenAt = nt.Time
		us = append(us, u)
		if err != nil {
			return nil, err
		}
	}
	return &us, nil
}

// Update persists changes to a given SiteUpdate (except for ID and SiteDefID fields).
func (s SiteUpdateStore) Update(u *models.SiteUpdate) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		SiteUpdateUpdateStmt,
		u.Ref,
		u.URL,
		u.Title,
		u.SeenAt,
		u.ID,
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

// NewSiteUpdateStore creates a new SiteUpdateStore.
func NewSiteUpdateStore(f func() *sql.DB) (*SiteUpdateStore, error) {
	store := &SiteUpdateStore{db: f()}
	_, err := store.db.Exec(SiteUpdateSchema)
	if err != nil {
		return nil, err
	}
	return store, nil
}

