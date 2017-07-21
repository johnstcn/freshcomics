package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Debug               bool 	`default:"false"`
	Host 				string 	`default:"0.0.0.0"`
	Port 				int 	`default:"3000"`
	DSN                 string  `default:"host=localhost user=freshcomics password=freshcomics_password dbname=freshcomicsdb sslmode=disable"`
}

var Cfg Config

func init() {
	err := envconfig.Process("freshcomics_", &Cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("[config] Debug:", Cfg.Debug)
	log.Println("[config] Host:", Cfg.Host)
	log.Println("[config] Port:", Cfg.Port)
}
