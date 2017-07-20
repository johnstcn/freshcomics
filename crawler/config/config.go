package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"

	"github.com/johnstcn/freshcomics/common/log"
)

type Config struct {
	Host 				string 	`default:"localhost"`
	Port 				int 	`default:"3000"`
	CheckIntervalSecs 	int 	`default:"600"`
	CrawlDispatchSecs 	int 	`default:"10"`
	DSN                 string  `default:"host=localhost user=freshcomics password=freshcomics_password dbname=freshcomicsdb sslmode=disable"`
	Backoff				[]int	`default:"1,10,30"`
}

var Cfg Config

func init() {
	err := envconfig.Process("freshcomics_crawler", &Cfg)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	log.Info("[config] Host:", Cfg.Host)
	log.Info("[config] Port:", Cfg.Port)
	log.Info("[config] CheckIntervalSecs:", Cfg.CheckIntervalSecs)
	log.Info("[config] CrawlDispatchSecs:", Cfg.CrawlDispatchSecs)
	log.Info("[config] Backoff:", Cfg.Backoff)
}
