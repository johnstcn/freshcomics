package models

import (
	"fmt"
	"net"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	"github.com/azer/snakecase"
	"github.com/fiorix/freegeoip"

	"github.com/johnstcn/freshcomics/common/log"
	"github.com/johnstcn/freshcomics/frontend/config"
)

var dao *FrontendDAO

type FrontendDAO struct {
	DB *sqlx.DB
	GeoIP *freegeoip.DB
}

func (d *FrontendDAO) GetComics() (*[]Comic, error) {
	comics := make([]Comic, 0)
	// TODO optimize this beast
	stmt := `SELECT site_defs.name, site_defs.nsfw, site_updates.id, site_updates.title, site_updates.seen_at
FROM site_updates JOIN site_defs ON (site_updates.site_def_id = site_defs.id)
WHERE site_updates.id IN (
  SELECT DISTINCT ON (site_def_id) id
  FROM site_updates
  ORDER BY site_def_id, seen_at DESC
) ORDER BY seen_at desc;`
	err := d.DB.Select(&comics, stmt)
	if err != nil {
		fmt.Println("error fetching latest comic list:", err)
		return nil, err
	}
	return &comics, nil
}

func (d *FrontendDAO) GetRedirectURL(updateID string) (string, error) {
	var result string
	stmt := `SELECT site_updates.url FROM site_updates WHERE id = $1`
	err := d.DB.Get(&result, stmt, updateID)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *FrontendDAO) getIPInfo(addr net.IP) (country, region, city string){
	var ipInfo freegeoip.DefaultQuery
	d.GeoIP.Lookup(addr, &ipInfo)
	country = ipInfo.Country.ISOCode
	if len(ipInfo.Region) > 0 {
		region = ipInfo.Region[0].ISOCode
	}
	if len(ipInfo.City.Names) > 0 {
		city = ipInfo.City.Names["en"]
	}
	log.Debug("GeoIP:", addr.String(), "->", country, region, city)
	return
}

func (d *FrontendDAO) RecordClick(updateID string, addr net.IP) error {
	uid, err := strconv.Atoi(updateID)
	if err != nil {
		log.Error("invalid updateID:", err)
	}
	stmt := `INSERT INTO comic_clicks (update_id, country, region, city) VALUES ($1, $2, $3, $4);`

	tx, err := d.DB.Beginx()
	if err != nil {
		log.Error(err)
	}
	country, region, city := d.getIPInfo(addr)
	_, err = tx.Exec(stmt, uid, country, region, city)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func init() {
	dsn := config.Cfg.DSN
	var db *sqlx.DB
	var err error
	for {
		db, err = sqlx.Connect("postgres", dsn)
		if err != nil {
			log.Info(err)
			<-time.After(1 * time.Second)
			continue
		}
		log.Info("Connected to database")
		break
	}
	log.Info("Connected to database")
	db.MapperFunc(snakecase.SnakeCase)
	db.MustExec(schema)

	geoIPRefresh := time.Duration(config.Cfg.GeoIPRefreshSecs) * time.Second
	geoIPFetchTimeout := time.Duration(config.Cfg.GeoIPFetchTimeoutSecs) * time.Second
	ip, err := freegeoip.OpenURL(freegeoip.MaxMindDB, geoIPRefresh, geoIPFetchTimeout)
	if err != nil {
		log.Error("Could not open MaxMind GeoIP DB:", err)
	}

	dao = &FrontendDAO{DB: db, GeoIP: ip}
}

func GetDAO() *FrontendDAO {
	return dao
}