package main

import (
	"github.com/johnstcn/freshcomics/pkg/crawld"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg, err := crawld.NewConfig()
	if err != nil {
		log.WithError(err).Fatal("init crawld config")
	}

	d, err := crawld.New(cfg)
	if err != nil {
		log.WithError(err).Fatal("init crawld")
	}

	log.Fatal(d.Run())
}
