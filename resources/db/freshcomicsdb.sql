CREATE TABLE IF NOT EXISTS site_defs (
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

CREATE TABLE IF NOT EXISTS site_updates (
	id 				serial		PRIMARY KEY,
	site_def_id 	integer 	REFERENCES site_defs (id) ON DELETE CASCADE,
	ref 			text		NOT NULL,
	url				text		NOT NULL UNIQUE,
	title			text		NOT NULL,
	seen_at			timestamptz	NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS crawl_info (
	id				    serial		  PRIMARY KEY,
	site_def_id		integer		  REFERENCES site_defs (id) ON DELETE CASCADE,
	started_at		timestamptz	NOT NULL DEFAULT CURRENT_TIMESTAMP,
	ended_at      timestamptz,
	status        text        NOT NULL,
	seen          integer     NOT NULL DEFAULT 0
);
