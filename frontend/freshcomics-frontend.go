package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/azer/snakecase"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/johnstcn/freshcomics/common/log"
)

var db *sqlx.DB

type comic struct {
	Name      string
	Title     string
	URL       string
	Published time.Time
}

func getDB() {
	dsn := os.Getenv("DATABASE_URL")
	db = sqlx.MustConnect("postgres", dsn)
	log.Info("connected to database")
	db.MapperFunc(snakecase.SnakeCase)
}

func getComics() (*[]comic, error) {
	comics := make([]comic, 0)
	stmt := `SELECT DISTINCT ON (site_defs.id) site_defs.name, site_updates.url, site_updates.title, site_updates.published
    FROM site_defs INNER JOIN site_updates ON site_defs.id = site_updates.site_def_id
    ORDER BY site_defs.id, site_updates.published DESC;`
	err := db.Select(&comics, stmt)
	if err != nil {
		fmt.Println("error fetching latest comic list:", err)
		return nil, err
	}
	return &comics, nil
}

func comicsHandler(resp http.ResponseWriter, req *http.Request) {
	comics, _ := getComics()
	enc := json.NewEncoder(resp)
	enc.SetIndent("", "\t")
	enc.Encode(comics)
}

func init() {
	getDB()
}

func main() {
	listenAddress := os.Getenv("HOSTPORT")
	http.HandleFunc("/", comicsHandler)
	http.ListenAndServe(listenAddress, http.DefaultServeMux)
}
