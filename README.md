# FreshComics

FreshComics crawls webcomics frequently so that folks who follow them can stay up to date.

It is written in [GoLang](http://golang.org) and consists of a frontend and a crawler (backend).

The frontend is quite barebones and just lists the most recently seen comics for all known sites. This updates periodically in the background, configurable via environment variable.

The crawler backend exposes a web UI to easily manage definitions for crawling comic sites. Crawl frequency and backoff are configurable via environment variables.

## Dependencies

### Local:
 * Vagrant/Virtualbox
 * Ansible 2.x

### System:
 * Systemd
 * Postgresql 9.6
 * Nginx

### Golang:
 * github.com/azer/snakecase
 * github.com/dustin/go-humanize
 * github.com/fiorix/freegeoip
 * github.com/gorilla/mux
 * github.com/jmoiron/sqlx
 * github.com/kelseyhightower/envconfig
 * github.com/tdewolff/minify
 * gopkg.in/xmlpath.v2
 * golang.org/x/net
 
## Local Setup

The following will bring up a Ubuntu 16.04 VM in Vagrant with a private IP of `192.168.12.34` running the frontend and crawler via `systemd`.

 * `vagrant up`
 * `cd deploy/staging`
 * `ansible-playbook staging.yml`

You can then visit the crawler UI at http://admin.freshcomics.192.168.12.34.xip.io and the frontend at http://freshcomics.192.168.12.34.xip.io.

