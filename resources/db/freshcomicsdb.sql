CREATE TABLE IF NOT EXISTS site_defs (
	id				      serial	 	  PRIMARY KEY,
	name 			      text 		    NOT NULL DEFAULT 'New SiteDef' UNIQUE,
	active		    	boolean		  NOT NULL DEFAULT FALSE,
	nsfw			      boolean		  NOT NULL DEFAULT FALSE,
	start_url 		  text 		    NOT NULL DEFAULT 'http://example.com' UNIQUE,
	url_template 	  text 		    NOT NULL DEFAULT 'http://example.com/%s' UNIQUE,
	ref_xpath 		  text 		    NOT NULL DEFAULT '//a[@rel="next"]/@href',
	ref_regexp 		  text 		    NOT NULL DEFAULT '([^/]+)/?$',
	title_xpath 	  text 		    NOT NULL DEFAULT '//title/text()',
	title_regexp 	  text 		    NOT NULL DEFAULT '(.+)'
);

CREATE TABLE IF NOT EXISTS site_updates (
	id 				    serial		  PRIMARY KEY,
	site_def_id 	integer 	  REFERENCES site_defs (id) ON DELETE CASCADE,
	ref 			    text		    NOT NULL,
	url				    text		    NOT NULL UNIQUE,
	title			    text		    NOT NULL,
	seen_at			  timestamptz	NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS crawl_infos (
	id				    serial		  PRIMARY KEY,
	site_def_id		integer		  REFERENCES site_defs (id) ON DELETE CASCADE,
	created_at    timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
	started_at		timestamptz,
	ended_at      timestamptz,
	error         text        NOT NULL DEFAULT '',
	seen          integer     NOT NULL DEFAULT 0
);
