#!/usr/bin/env bash
export FRESHCOMICS_CRAWLER_DEBUG=true
export FRESHCOMICS_CRAWLER_HOST=localhost
export FRESHCOMICS_CRAWLER_PORT=3000
export FRESHCOMICS_CRAWLER_DSN='host=localhost user=freshcomics password=freshcomics_password dbname=freshcomicsdb sslmode=disable'
go run freshcomics-crawler.go