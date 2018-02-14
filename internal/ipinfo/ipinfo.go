package ipinfo

import (
	"net"
	"time"

	"github.com/fiorix/freegeoip"
	"github.com/pkg/errors"
)

//go:generate mockery -interface IPInfoer -package ipinfotest

type GeoLoc struct {
	Country string
	Region  string
	City    string
}

type IPInfoer interface {
	GetIPInfo(addr net.IP) (GeoLoc, error)
}

type ipInfoer struct {
	geoIP Lookuper
}

type Lookuper interface {
	Lookup(addr net.IP, result interface{}) error
}

type urlOpener func(url string, refreshSecs, fetchTimeoutSecs time.Duration) (*freegeoip.DB, error)

var _ IPInfoer = (*ipInfoer)(nil)
var _ Lookuper = (*freegeoip.DB)(nil)

func NewIPInfoer(refreshSecs, fetchTimeoutSecs int) (IPInfoer, error) {
	return newIPInfoer(freegeoip.OpenURL, refreshSecs, fetchTimeoutSecs)
}

func newIPInfoer(openURL urlOpener, refreshSecs, fetchTimeoutSecs int) (IPInfoer, error) {
	geoIPRefresh := time.Duration(refreshSecs) * time.Second
	geoIPFetchTimeout := time.Duration(fetchTimeoutSecs) * time.Second
	ip, err := openURL(freegeoip.MaxMindDB, geoIPRefresh, geoIPFetchTimeout)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open MaxMind geoIP db")
	}

	return &ipInfoer{
		geoIP: ip,
	}, nil
}

func (i *ipInfoer) GetIPInfo(addr net.IP) (GeoLoc, error) {
	var ipInfo freegeoip.DefaultQuery
	var g GeoLoc
	err := i.geoIP.Lookup(addr, &ipInfo)
	if err != nil {
		return GeoLoc{}, err
	}
	g.Country = ipInfo.Country.ISOCode
	if len(ipInfo.Region) > 0 {
		g.Region = ipInfo.Region[0].ISOCode
	}
	if len(ipInfo.City.Names) > 0 {
		g.City = ipInfo.City.Names["en"]
	}
	return g, nil
}
