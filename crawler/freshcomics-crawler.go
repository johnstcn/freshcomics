package main

import (
	"time"

	"github.com/johnstcn/freshcomics/crawler/models"
	"github.com/johnstcn/freshcomics/crawler/util"
	"github.com/johnstcn/freshcomics/crawler/web"
)

var DEFAULT_CHECK_INTERVAL = 60 * time.Minute
var DEFAULT_FULL_CRAWL = false

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
				go util.Crawl(def, DEFAULT_FULL_CRAWL)
			}
		}
		time.Sleep(tick)
	}
}
