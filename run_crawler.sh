#!/usr/bin/env bash
#DATABASE_URL='freshcomics:freshcomics_password@tcp(localhost:5432)/freshcomicsdb?charset=utf8&parseTime=True&loc=Local&sslmode=disable' go run main.go
VERBOSE=1 HOSTPORT='localhost:8080' DATABASE_URL='host=localhost user=freshcomics password=freshcomics_password dbname=freshcomicsdb sslmode=disable' go run main/crawler.go
