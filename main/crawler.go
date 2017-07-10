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


//func CreateTestSiteDef(conn *gorm.DB) {
//
//	sd, _ := crawler.NewSiteDef(
//		"Forest Frenemies",
//		"https://forestfrenemies.wordpress.com/comic/salad-bar/",
//		"/comic/([^/]+)",
//		"//a[@rel=\"next\"]/@href",
//		"//meta[@property=\"og:title\"]/@content",
//		"(.*)|",
//		"//meta[@property=\"article:published_time\"]/@content",
//		"(.*)",
//		"2006-01-02T15:04:05-07:00",
//	)
//	conn.Create(sd)
//}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open("postgres", dsn)
	//db.LogMode(true)
	if err != nil {
		log.Fatalln("Error connecting to database:", err)
	}
	log.Println("Connected to database")
	defer db.Close()
	db.AutoMigrate(crawler.SiteDef{}, crawler.SiteUpdate{})

	go func() {
		for {
			tick := 1 * time.Second
			next := crawler.GetLastCheckedSiteDef(db)
			if next != nil {
				crawler.Crawl(db, next)
			}
			time.Sleep(tick)
		}
	}()

}
