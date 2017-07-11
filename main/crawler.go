package main

import (
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/johnstcn/freshcomics/crawler"
)

func GetDB() (*gorm.DB) {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatalln("Error connecting to database:", err)
	}
	if os.Getenv("VERBOSE") == "1" {
		db.LogMode(true)
	}
	log.Println("Connected to database")

	db.AutoMigrate(crawler.SiteDef{}, crawler.SiteUpdate{})

	return db
}

func main() {
	db := GetDB()
	defer db.Close()
	go crawler.ServeAdmin(db)
	for {
		tick := 1 * time.Second
		next := crawler.GetLastCheckedSiteDef(db)
		if next != nil {
			go crawler.Crawl(db, next)
		}
		time.Sleep(tick)
	}
}
