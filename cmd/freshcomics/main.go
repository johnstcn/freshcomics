package main

//go:generate go-bindata -prefix "web/templates/" -pkg web -o web/templates.go web/templates/

import (
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/johnstcn/freshcomics/internal/common/log"
	"github.com/johnstcn/freshcomics/internal/frontend/config"
	"github.com/johnstcn/freshcomics/internal/frontend/web"
	"github.com/johnstcn/freshcomics/internal/store"
)

func main() {
	log.Info("Starting up")
	conn, err := sqlx.Connect("postgres", config.Cfg.DSN)
	if err != nil {
		panic(err)
	}

	store, err := store.NewPGStore(conn)
	if err != nil {
		panic(err)
	}

	listenAddress := fmt.Sprintf("%s:%d", config.Cfg.Host, config.Cfg.Port)
	log.Info("Listening on", listenAddress)
	fe := web.NewFrontend(store)
	if err != nil {
		panic(err)
	}
	http.ListenAndServe(listenAddress, fe)
}
