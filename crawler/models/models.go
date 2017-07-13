package models

import (
	"time"
)

var schema = `
CREATE TABLE IF NOT EXISTS site_defs (
	id				serial	 	PRIMARY KEY,
	name 			text 		NOT NULL DEFAULT 'New SiteDef' UNIQUE,
	active			boolean		NOT NULL DEFAULT FALSE,
	nsfw			boolean		NOT NULL DEFAULT FALSE,
	start_url 		text 		NOT NULL DEFAULT 'http://example.com' UNIQUE,
	last_checked 	timestamp	NOT NULL DEFAULT CURRENT_TIMESTAMP,
	url_template 	text 		NOT NULL DEFAULT 'http://example.com/%s' UNIQUE,
	ref_xpath 		text 		NOT NULL DEFAULT '//a[@rel="next"]/@href',
	ref_regexp 		text 		NOT NULL DEFAULT '([^/]+)/?$',
	title_xpath 	text 		NOT NULL DEFAULT '//title/text()',
	title_regexp 	text 		NOT NULL DEFAULT '(.+)',
	date_xpath 		text 		NOT NULL DEFAULT '',
	date_regexp 	text 		NOT NULL DEFAULT '(.+)',
	date_format 	text 		NOT NULL DEFAULT '',
	next_page_xpath text		NOT NULL DEFAULT '//a[@rel="next"]/@href',
	next_page_regexp text		NOT NULL DEFAULT '([^/]+)/?$'
);

CREATE INDEX IF NOT EXISTS site_defs_last_checked_idx ON site_defs (last_checked);

CREATE TABLE IF NOT EXISTS site_updates (
	id 				serial		PRIMARY KEY,
	site_def_id 	integer 	REFERENCES site_defs (id) ON DELETE CASCADE,
	ref 			text		NOT NULL,
	url				text		NOT NULL UNIQUE,
	title			text		NOT NULL,
	published		timestamp	NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS site_updates_ref_idx ON site_updates (ref);
CREATE INDEX IF NOT EXISTS site_updates_published_idx ON site_updates(published);
`

type SiteDef struct {
	ID          	int64
	Name        	string
	Active      	bool
	NSFW        	bool
	StartURL    	string
	LastChecked 	time.Time
	URLTemplate 	string
	RefXpath    	string
	RefRegexp   	string
	TitleXpath  	string
	TitleRegexp 	string
	DateXpath   	string
	DateRegexp  	string
	DateFormat  	string
	NextPageXpath 	string
	NextPageRegexp 	string
}

type SiteUpdate struct {
	ID 			int64
	SiteDefID 	int64
	Ref       	string
	URL       	string
	Title     	string
	Published   time.Time
}
