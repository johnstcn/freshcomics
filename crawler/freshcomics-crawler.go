package main

//go:generate go-bindata -prefix "web/templates/" -pkg web -o web/templates.go web/templates/

import (
	"time"

	"github.com/johnstcn/freshcomics/crawler/config"
	"github.com/johnstcn/freshcomics/crawler/models"
	"github.com/johnstcn/freshcomics/crawler/util"
	"github.com/johnstcn/freshcomics/crawler/web"
	"github.com/johnstcn/freshcomics/common/log"
)



func main() {
	tick := time.Duration(config.Cfg.CrawlDispatchSecs) * time.Second
	checkInterval := time.Duration(config.Cfg.CheckIntervalSecs) * time.Second
	dao := models.GetDAO()
	defer dao.DB.Close()
	go web.ServeAdmin(config.Cfg.Host, config.Cfg.Port)
	for {
		now := time.Now().UTC()
		def, _ := dao.GetSiteDefLastChecked()
		if def != nil {
			delta := now.Sub(def.LastCheckedAt)
			shouldCheck := delta > checkInterval
			log.Debug("def", def.Name, "last checked", def.LastCheckedAt)
			log.Debug("now", now)
			log.Debug("checkInterval:", checkInterval)
			log.Debug("delta:", delta)
			if shouldCheck {
				go util.Crawl(def)
			} else {
				log.Debug("nothing to schedule")
			}
		}
		time.Sleep(tick)
		log.Debug("sleeping for", tick)
	}
}
