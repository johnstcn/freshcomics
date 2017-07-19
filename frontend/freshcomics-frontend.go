package main

//go:generate go-bindata -prefix "web/templates/" -pkg web -o web/templates.go web/templates/


import (
	"os"

	"github.com/johnstcn/freshcomics/frontend/web"
	"github.com/johnstcn/freshcomics/common/log"
)

func main() {
	log.Info("Starting up")
	web.ServeFrontend(os.Getenv("HOST"), os.Getenv("PORT"))
}
