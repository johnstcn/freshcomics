package store

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/fiorix/freegeoip"
	"github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/mock"
)

type MockLookuper struct {
	mock.Mock
}

func (m *MockLookuper) Lookup(addr net.IP, result interface{}) error {
	args := m.Called(addr, result)
	return args.Error(0)
}

type IPInfoTestSuite struct {
	suite.Suite
	ip IPInfoer
	ml *MockLookuper
}

func (s *IPInfoTestSuite) SetupSuite() {
	s.ml = &MockLookuper{}
	s.ip = &ipInfoer{
		geoIP: s.ml,
	}
}

func (s *IPInfoTestSuite) TearDownTest() {
	s.ml.AssertExpectations(s.T())
}

func (s *IPInfoTestSuite) TestNewIPInfoer_OK() {
	ip, err := NewIPInfoer(60, 10)
	s.NotNil(ip)
	s.NoError(err)
}

func (s *IPInfoTestSuite) TestNewIPInfoer_Err() {
	badOpener := func(_ string, _, _ time.Duration) (*freegeoip.DB, error) {
		return nil, fmt.Errorf("some error")
	}
	ip, err := newIPInfoer(badOpener, 60, 10)
	s.Nil(ip)
	s.EqualError(err, "Could not open MaxMind geoIP db: some error")
}

func (s *IPInfoTestSuite) TestGetIPInfo_OK() {
	s.ml.On("Lookup", mock.AnythingOfType("net.IP"), mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(1).(*freegeoip.DefaultQuery)
		result.Country.ISOCode = "IE"
		result.Region = append(result.Region, struct{
			ISOCode string `maxminddb:"iso_code"`
			Names map[string]string `maxminddb:"names"`
		}{
			ISOCode: "L",
			Names: map[string]string{},
		})
		result.City = struct{
			Names map[string]string `maxminddb:"names"`
		}{
			Names: map[string]string{"en": "Dublin"},
		}
	}).Return(nil).Once()
	g, err := s.ip.GetIPInfo(net.ParseIP("169.254.169.264"))
	s.NotZero(g)
	s.EqualValues("IE", g.Country)
	s.EqualValues("L", g.Region)
	s.EqualValues("Dublin", g.City)
	s.NoError(err)
}

func (s *IPInfoTestSuite) TestGetIPInfo_Err() {
	s.ml.On("Lookup", mock.AnythingOfType("net.IP"), mock.Anything).Return(fmt.Errorf("some error")).Once()
	g, err := s.ip.GetIPInfo(net.ParseIP("169.254.169.254"))
	s.Zero(g)
	s.EqualError(err, "some error")
}

func TestIPInfoTestSuite(t *testing.T) {
	suite.Run(t, new(IPInfoTestSuite))
}
