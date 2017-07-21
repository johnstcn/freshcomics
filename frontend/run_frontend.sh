#!/usr/bin/env bash
export DEBUG=1
export FRESHCOMICS_FRONTEND_HOST=localhost
export FRESHCOMICS_FRONTEND_PORT=8081
export FRESHCOMICS_FRONTEND_DSN='host=localhost user=freshcomics password=freshcomics_password dbname=freshcomicsdb sslmode=disable'
go run freshcomics-frontend.go