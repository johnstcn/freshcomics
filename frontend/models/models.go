package models

import (
	"time"
)

type Comic struct {
	ID			int64
	Name      	string
	Title     	string
	SeenAt	 	time.Time
	NSFW 		bool
}

type ClickLog struct {
	ID			int64
	UpdateID	int64
	ClickedAt	time.Time
	Country		string
	Region		string
	City		string
	Zip			string
}

var schema = `CREATE TABLE IF NOT EXISTS comic_clicks (
	id			serial		PRIMARY KEY,
	update_id	integer		REFERENCES site_updates (id) ON DELETE CASCADE,
	clicked_at  timestamptz	NOT NULL DEFAULT CURRENT_TIMESTAMP,
	country		text		NOT NULL,
	region		text		NOT NULL,
	city		text		NOT NULL
);

CREATE INDEX IF NOT EXISTS comic_clicks_country_idx ON comic_clicks (country);
CREATE INDEX IF NOT EXISTS comic_clicks_region_idx ON comic_clicks (region);
`