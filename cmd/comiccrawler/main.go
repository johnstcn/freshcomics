package comiccrawler

//go:generate go-bindata -prefix "web/templates/" -pkg web -o web/templates.go web/templates/

import (
	"fmt"
	"net/http"
	"time"

	"github.com/johnstcn/freshcomics/internal/crawler/config"
	"github.com/johnstcn/freshcomics/internal/crawler/util"
	"github.com/johnstcn/freshcomics/internal/crawler/web"
	"github.com/johnstcn/freshcomics/internal/common/log"
	"github.com/johnstcn/freshcomics/internal/store"
)


func ServeAdmin(host string, port int, a *web.Admin) {
	listenAddress := fmt.Sprintf("%s:%d", host, port)
	log.Info("Listening on", listenAddress)
	log.Error(http.ListenAndServe(listenAddress, a))
}

func ScheduleCrawls(s store.Store, tick, checkInterval time.Duration) {
	for {
		now := time.Now().UTC()
		def, _ := s.GetSiteDefLastChecked()
		if def != nil {
			delta := now.Sub(def.LastCheckedAt)
			shouldCheck := delta > checkInterval
			log.Debug("def", def.Name, "last checked", def.LastCheckedAt)
			log.Debug("now", now)
			log.Debug("checkInterval:", checkInterval)
			log.Debug("delta:", delta)
			if shouldCheck {
				go util.Crawl(s, def)
			} else {
				log.Debug("nothing to schedule")
			}
		}
		log.Debug("sleeping for", tick)
		<-time.After(tick)
	}
}

func main() {
	tick := time.Duration(config.Cfg.CrawlDispatchSecs) * time.Second
	checkInterval := time.Duration(config.Cfg.CheckIntervalSecs) * time.Second
	store, err := store.NewStore(config.Cfg.DSN)
	if err != nil {
		panic(err)
	}
	admin := web.NewAdmin(store)
	go ScheduleCrawls(store, tick, checkInterval)
	ServeAdmin(config.Cfg.Host, config.Cfg.Port, admin)
}
