package main

//go:generate go-bindata -prefix "web/templates/" -pkg web -o web/templates.go web/templates/

import (
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/johnstcn/freshcomics/internal/frontend/config"
	"github.com/johnstcn/freshcomics/internal/frontend/web"
	"github.com/johnstcn/freshcomics/internal/store"
	"golang.org/x/exp/slog"
)

func main() {
	conn, err := sqlx.Connect("postgres", config.Cfg.DSN)
	if err != nil {
		panic(err)
	}

	store, err := store.NewPGStore(conn)
	if err != nil {
		panic(err)
	}

	listenAddress := fmt.Sprintf("%s:%d", config.Cfg.Host, config.Cfg.Port)
	slog.Info("listen", "host", config.Cfg.Host, "port", config.Cfg.Port)
	fe := web.NewFrontend(store)
	if err != nil {
		panic(err)
	}
	http.ListenAndServe(listenAddress, fe)
}
