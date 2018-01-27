package store

import (
	"time"
)

type Comic struct {
	ID			int64		`db:"id"`
	Name      	string		`db:"name"`
	Title     	string		`db:"title"`
	SeenAt	 	time.Time	`db:"seen_at"`
	NSFW 		bool		`db:"nsfw"`
}

type ClickLog struct {
	ID			int64		`db:"id"`
	UpdateID	int64		`db:"update_id"`
	ClickedAt	time.Time	`db:"clicked_at"`
	Country		string		`db:"country"`
	Region		string		`db:"region"`
	City		string		`db:"city"`
}

type SiteDef struct {
	ID            int64		`db:"id"`
	Name          string	`db:"name"`
	Active        bool		`db:"active"`
	NSFW          bool		`db:"nsfw"`
	StartURL      string	`db:"start_url"`
	LastCheckedAt time.Time	`db:"last_checked_at"`
	URLTemplate   string	`db:"url_template"`
	RefXpath      string	`db:"ref_xpath"`
	RefRegexp     string	`db:"ref_regexp"`
	TitleXpath    string	`db:"title_xpath"`
	TitleRegexp   string	`db:"title_regexp"`
}

type SiteUpdate struct {
	ID        int64			`db:"id"`
	SiteDefID int64			`db:"site_def_id"`
	Ref       string		`db:"ref"`
	URL       string		`db:"url"`
	Title     string		`db:"title"`
	SeenAt    time.Time		`db:"seen_at"`
}

type CrawlInfo struct {
	ID        int64          `db:"id"`
	SiteDefID int64          `db:"site_def_id"`
	StartedAt time.Time      `db:"started_at"`
	EndedAt   time.Time      `db:"ended_at"`
	Status    string         `db:"status"`
	Seen      int            `db:"seen"`
}
