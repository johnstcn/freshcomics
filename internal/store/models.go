package store

import (
	"time"
)

type ComicID int64
type ClickLogID int64
type SiteDefID int64
type SiteUpdateID int64
type CrawlInfoID int64

type Comic struct {
	ID     ComicID   `db:"id"`
	Name   string    `db:"name"`
	Title  string    `db:"title"`
	SeenAt time.Time `db:"seen_at"`
	NSFW   bool      `db:"nsfw"`
}

type ClickLog struct {
	ID        ClickLogID   `db:"id"`
	UpdateID  SiteUpdateID `db:"update_id"`
	ClickedAt time.Time    `db:"clicked_at"`
	Country   string       `db:"country"`
	Region    string       `db:"region"`
	City      string       `db:"city"`
}

type SiteDef struct {
	ID            SiteDefID `db:"id"`
	Name          string    `db:"name"`
	Active        bool      `db:"active"`
	NSFW          bool      `db:"nsfw"`
	StartURL      string    `db:"start_url"`
	URLTemplate   string    `db:"url_template"`
	RefXpath      string    `db:"ref_xpath"`
	RefRegexp     string    `db:"ref_regexp"`
	TitleXpath    string    `db:"title_xpath"`
	TitleRegexp   string    `db:"title_regexp"`
}

type SiteUpdate struct {
	ID        SiteUpdateID `db:"id"`
	SiteDefID SiteDefID    `db:"site_def_id"`
	Ref       string       `db:"ref"`
	URL       string       `db:"url"`
	Title     string       `db:"title"`
	SeenAt    time.Time    `db:"seen_at"`
}

type CrawlInfo struct {
	ID        CrawlInfoID `db:"id"`
	SiteDefID SiteDefID   `db:"site_def_id"`
	URL       string      `db:"url"`
	StartedAt time.Time   `db:"started_at"`
	EndedAt   time.Time   `db:"ended_at"`
	Error     string      `db:"error"`
	Seen      int         `db:"seen"`
}
