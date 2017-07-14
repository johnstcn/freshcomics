package main

//go:generate go-bindata -prefix "web/templates/" -pkg web -o web/templates.go web/templates/

import (
	"time"

	"github.com/johnstcn/freshcomics/crawler/models"
	"github.com/johnstcn/freshcomics/crawler/util"
	"github.com/johnstcn/freshcomics/crawler/web"
)

var DEFAULT_CHECK_INTERVAL = 60 * time.Minute

func main() {
	dao := models.GetDAO()
	defer dao.DB.Close()
	go web.ServeAdmin()
	for {
		tick := 1 * time.Second
		def, _ := dao.GetSiteDefLastChecked()
		if def != nil {
			shouldCheck := time.Now().Sub(def.LastChecked) > DEFAULT_CHECK_INTERVAL
			if shouldCheck {
				go util.Crawl(def)
			}
		}
		time.Sleep(tick)
	}
}
