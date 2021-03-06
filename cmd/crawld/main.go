package main

import (
	"github.com/johnstcn/freshcomics/internal/store"
	"github.com/johnstcn/freshcomics/pkg/crawld"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	cfg, err := crawld.NewConfig()
	if err != nil {
		log.WithError(err).Fatal("init crawld config")
	}
	log.SetReportCaller(cfg.LogCallerTrace)

	conn, err := sqlx.Connect("postgres", cfg.DSN)
	if err != nil {
		log.WithError(err).Fatal("could not connect to database")
	}

	pgstore, err := store.NewPGStore(conn)
	if err != nil {
		log.WithError(err).Fatal("init pgstore")
	}

	d, err := crawld.New(cfg, pgstore)
	if err != nil {
		log.WithError(err).Fatal("init crawld")
	}

	log.Fatal(d.Run())
}
