#!/usr/bin/env bash
export DEBUG=true
export FRESHCOMICS_CRAWLER_HOST=localhost
export FRESHCOMICS_CRAWLER_PORT=3001
export FRESHCOMICS_CRAWLER_DSN='host=localhost user=freshcomics password=freshcomics_password dbname=freshcomicsdb sslmode=disable'
export FRESHCOMICS_CRAWLER_CHECKINTERVALSECS=60
export FRESHCOMICS_CRAWLER_CRAWLDISPATCHSECS=1
export FRESHCOMICS_CRAWLER_BACKOFF=1,1
go run freshcomics-crawler.go