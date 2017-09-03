package dao

import (
	"github.com/johnstcn/freshcomics/backendsvc/models"
	"database/sql"
)

const (
	EventSchema = `
CREATE TABLE IF NOT EXISTS events (
	id		serial		PRIMARY KEY,
	at		timestamptz	NOT NULL DEFAULT CURRENT_TIMESTAMP,
	type	text		NOT NULL,
	tags	JSONB		NOT NULL DEFAULT {}
);

CREATE INDEX IF NOT EXISTS events_type_idx ON events(type);
CREATE INDEX IF NOT EXISTS events_at_idx ON events(at);`
	EventCreateStmt = `INSERT INTO events DEFAULT VALUES;`
)

type EventManager interface {
	Create(t string, info map[string]string) (int64, error)
	Get(id int64) (*models.Event, error)
	GetByType(t string, info map[string]string) error
}

type EventStore struct {
	db *sql.DB
}

func (s EventStore) Create(t string) (int64, error) {
	var newID int64
	tx, err := s.db.Begin()
	if err != nil {
		return -1, err
	}
	row := tx.QueryRow(EventCreateStmt)
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

func (s EventStore) Get(id int64) (*models.Event, error) {
	panic("implement me")
}

func (s EventStore) GetAllByType(t string) error {
	panic("implement me")
}

func (s EventStore) Update(*models.Event) error {
	panic("implement me")
}


