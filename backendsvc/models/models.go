package models

import (
	"time"
	"github.com/jmoiron/sqlx/types"
)

var schema = `
CREATE TABLE IF NOT EXISTS events (
	id		serial		PRIMARY KEY,
	at		timestamptz	NOT NULL DEFAULT CURRENT_TIMESTAMP,
	type	text		NOT NULL,
	info	JSONB		NOT NULL
);

CREATE INDEX IF NOT EXISTS events_at_idx ON events(at);
CREATE INDEX IF NOT EXISTS events_type_idx ON events(type);
`

// SiteDef is a collection of metadata about a comic site.
type SiteDef struct {
	ID            int64
	Name          string
	Active        bool
	NSFW          bool
	StartURL      string
	LastCheckedAt time.Time
	URLTemplate   string
	RefXpath      string
	RefRegexp     string
	TitleXpath    string
	TitleRegexp   string
}

type SiteUpdate struct {
	ID        int64
	SiteDefID int64
	Ref       string
	URL       string
	Title     string
	SeenAt    time.Time
}

type Event struct {
	ID   	int64
	At 		time.Time
	Type 	string
	Tags   	types.JSONText
}