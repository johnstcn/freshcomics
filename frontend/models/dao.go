package models

import (
	"os"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	"github.com/azer/snakecase"

	"github.com/johnstcn/freshcomics/common/log"
	"fmt"
)

var dao *FrontendDAO

type FrontendDAO struct {
	DB *sqlx.DB
}

func GetDAO() *FrontendDAO {
	if dao == nil {
		dsn := os.Getenv("DATABASE_URL")
		db := sqlx.MustConnect("postgres", dsn)
		log.Info("Connected to database")
		db.MapperFunc(snakecase.SnakeCase)
		dao = &FrontendDAO{DB: db}
	}
	return dao
}

func (d *FrontendDAO) GetComics() (*[]Comic, error) {
	comics := make([]Comic, 0)
	stmt := `SELECT DISTINCT ON (site_defs.id) site_defs.name, site_updates.url, site_updates.title, site_updates.published
FROM site_defs INNER JOIN site_updates ON site_defs.id = site_updates.site_def_id
ORDER BY site_defs.id, site_updates.published DESC;`
	err := d.DB.Select(&comics, stmt)
	if err != nil {
		fmt.Println("error fetching latest comic list:", err)
		return nil, err
	}
	return &comics, nil
}