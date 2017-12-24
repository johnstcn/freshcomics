package main

//go:generate go-bindata -prefix "web/templates/" -pkg web -o web/templates.go web/templates/


import (
	"fmt"
	"net/http"

	"github.com/johnstcn/freshcomics/common/log"
	"github.com/johnstcn/freshcomics/frontend/web"
	"github.com/johnstcn/freshcomics/frontend/config"
	"github.com/johnstcn/freshcomics/common/store"
)

func main() {
	log.Info("Starting up")
	store, err := store.NewStore(config.Cfg.DSN)
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
