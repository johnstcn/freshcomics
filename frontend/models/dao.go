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
		dsn := os.Getenv("DSN")
		db := sqlx.MustConnect("postgres", dsn)
		log.Info("Connected to database")
		db.MapperFunc(snakecase.SnakeCase)
		dao = &FrontendDAO{DB: db}
	}
	return dao
}

func (d *FrontendDAO) GetComics() (*[]Comic, error) {
	comics := make([]Comic, 0)
	// TODO optimize this beast
	stmt := `SELECT site_defs.name, site_defs.nsfw, site_updates.title, site_updates.url, site_updates.published
FROM site_updates JOIN site_defs ON (site_updates.site_def_id = site_defs.id)
WHERE site_updates.id IN (
  SELECT DISTINCT ON (site_def_id) id
  FROM site_updates
  ORDER BY site_def_id, published DESC
) ORDER BY published desc;`
	err := d.DB.Select(&comics, stmt)
	if err != nil {
		fmt.Println("error fetching latest comic list:", err)
		return nil, err
	}
	return &comics, nil
}