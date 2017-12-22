package models

import (
	"time"
	"github.com/jmoiron/sqlx/types"
)

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

type CrawlEvent struct {
	ID        int64
	SiteDefID int64
	CreatedAt time.Time
	EventType string
	EventInfo types.JSONText
}