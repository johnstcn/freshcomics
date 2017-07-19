package main

//go:generate go-bindata -prefix "web/templates/" -pkg web -o web/templates.go web/templates/


import (
	"os"

	"github.com/johnstcn/freshcomics/frontend/web"
)

func main() {
	web.ServeFrontend(os.Getenv("HOST"), os.Getenv("PORT"))
}
