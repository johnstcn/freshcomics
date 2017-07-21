package config

import (
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"

	"github.com/johnstcn/freshcomics/common/config"
)

type FrontendConfig struct {
	config.Config
	RefreshIntervalSecs 	int 	`default:"60"`
}

var Cfg FrontendConfig

func init() {
	err := envconfig.Process("freshcomics_frontend", &Cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("[config] Debug:", os.Getenv("DEBUG") != "")
	log.Println("[config] Host:", Cfg.Host)
	log.Println("[config] Port:", Cfg.Port)
	log.Println("[config] RefreshIntervalSecs:", Cfg.RefreshIntervalSecs)
	log.Println("[config] DSN:", Cfg.DSN)
}
