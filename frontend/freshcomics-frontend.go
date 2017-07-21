package main

//go:generate go-bindata -prefix "web/templates/" -pkg web -o web/templates.go web/templates/


import (
	"github.com/johnstcn/freshcomics/common/log"
	"github.com/johnstcn/freshcomics/frontend/web"
	"github.com/johnstcn/freshcomics/frontend/config"
)

func main() {
	log.Info("Starting up")
	web.ServeFrontend(config.Cfg.Host, config.Cfg.Port)
}
