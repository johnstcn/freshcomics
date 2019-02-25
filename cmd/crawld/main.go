package main

import (
	"log"

	"github.com/johnstcn/freshcomics/pkg/crawld"
)

func main() {
	cfg, err := crawld.NewConfig()
	if err != nil {
		log.Fatal("init crawld config:", err)
	}

	d, err := crawld.New(cfg)
	if err != nil {
		log.Fatal("init crawld:", err)
	}

	log.Fatal(d.Run())
}
