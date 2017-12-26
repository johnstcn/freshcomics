package store

import (
	"net"
	"time"

	"github.com/fiorix/freegeoip"
	"github.com/pkg/errors"
)

type GeoLoc struct {
	Country string
	Region  string
	City    string
}

type IPInfoer interface {
	GetIPInfo(addr net.IP) GeoLoc
}

type ipInfoer struct {
	geoIP *freegeoip.DB
}

var _ IPInfoer = (*ipInfoer)(nil)

func NewIPInfoer(refreshSecs, fetchTimeoutSecs int) (IPInfoer, error) {
	geoIPRefresh := time.Duration(refreshSecs) * time.Second
	geoIPFetchTimeout := time.Duration(fetchTimeoutSecs) * time.Second
	ip, err := freegeoip.OpenURL(freegeoip.MaxMindDB, geoIPRefresh, geoIPFetchTimeout)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open MaxMind geoIP db")
	}

	return &ipInfoer{
		geoIP: ip,
	}, nil
}

func (i *ipInfoer) GetIPInfo(addr net.IP) GeoLoc {
	var ipInfo freegeoip.DefaultQuery
	var g GeoLoc
	i.geoIP.Lookup(addr, &ipInfo)
	g.Country = ipInfo.Country.ISOCode
	if len(ipInfo.Region) > 0 {
		g.Region = ipInfo.Region[0].ISOCode
	}
	if len(ipInfo.City.Names) > 0 {
		g.City = ipInfo.City.Names["en"]
	}
	return g
}
