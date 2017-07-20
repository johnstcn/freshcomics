package config

import (
	"log"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Debug 				bool	`default:"false"`
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
		log.Fatal(err.Error())
	}
}
