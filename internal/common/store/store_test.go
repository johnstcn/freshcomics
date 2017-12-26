package store

import (
	"fmt"
	"net"
	"time"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"database/sql/driver"
)

type StoreTestSuite struct {
	suite.Suite
	store *store
	mip   *mockIPInfoer
	mdb   sqlmock.Sqlmock
	now   func() time.Time
}

type mockIPInfoer struct {
	mock.Mock
}

func (m *mockIPInfoer) GetIPInfo(addr net.IP) (GeoLoc, error) {
	args := m.Called(addr)
	return args.Get(0).(GeoLoc), args.Error(1)
}

func (s *StoreTestSuite) SetupSuite() {
	conn, mdb, err := sqlmock.New()
	if err != nil {
		s.Fail(err.Error())
	}
	s.mdb = mdb
	s.mip = &mockIPInfoer{}
	s.store = &store{
		db:    sqlx.NewDb(conn, "sqlmock"),
		geoIP: s.mip,
	}
	s.now = func() time.Time {
		return time.Unix(0, 0)
	}
}

func (s *StoreTestSuite) TearDownTest() {
	s.NoError(s.mdb.ExpectationsWereMet())
	s.mip.AssertExpectations(s.T())
}

func (s *StoreTestSuite) TestGetComics_OK() {
	rows := sqlmock.NewRows([]string{"name", "nsfw", "id", "title", "seen_at"}).AddRow("Test Comic", false, 1, "Test Title", s.now())
	s.mdb.ExpectQuery(`SELECT site_defs.name, site_defs.nsfw, site_updates.id, site_updates.title, site_updates.seen_at FROM site_updates JOIN site_defs ON \(site_updates.site_def_id = site_defs.id\) WHERE site_updates.id IN \( SELECT DISTINCT ON \(site_def_id\) id FROM site_updates ORDER BY site_def_id, seen_at DESC \) ORDER BY seen_at desc;`).WillReturnRows(rows)
	comics, err := s.store.GetComics()
	s.NotNil(comics)
	s.EqualValues("Test Comic", comics[0].Name)
	s.NoError(err)
}

func (s *StoreTestSuite) TestGetComics_Err() {
	s.mdb.ExpectQuery(`SELECT site_defs.name, site_defs.nsfw, site_updates.id, site_updates.title, site_updates.seen_at FROM site_updates JOIN site_defs ON \(site_updates.site_def_id = site_defs.id\) WHERE site_updates.id IN \( SELECT DISTINCT ON \(site_def_id\) id FROM site_updates ORDER BY site_def_id, seen_at DESC \) ORDER BY seen_at desc;`).WillReturnError(fmt.Errorf("some error"))
	comics, err := s.store.GetComics()
	s.Nil(comics)
	s.EqualError(err, "some error")
}

func (s *StoreTestSuite) TestGetRedirectURL_OK() {
	rows := sqlmock.NewRows([]string{"url"}).AddRow("http://example.com")
	s.mdb.ExpectQuery(`SELECT site_updates.url FROM site_updates WHERE id = \$1`).WithArgs("12345").WillReturnRows(rows)
	url, err := s.store.GetRedirectURL("12345")
	s.NoError(err)
	s.EqualValues("http://example.com", url)
}

func (s *StoreTestSuite) TestGetRedirectURL_Err() {
	s.mdb.ExpectQuery(`SELECT site_updates.url FROM site_updates WHERE id = \$1`).WithArgs("12345").WillReturnError(fmt.Errorf("some error"))
	url, err := s.store.GetRedirectURL("12345")
	s.Zero(url)
	s.EqualError(err, "some error")
}

func (s *StoreTestSuite) TestRecordClick_OK() {
	ip := net.ParseIP("169.254.169.254")
	s.mip.On("GetIPInfo", ip).Return(GeoLoc{
		Country: "IE",
		Region: "L",
		City: "Dublin",
	}, nil).Once()
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(`INSERT INTO comic_clicks \(update_id, country, region, city\) VALUES \(\$1, \$2, \$3, \$4\);`).WithArgs(12345, "IE", "L", "Dublin").WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit()
	err := s.store.RecordClick("12345", ip)
	s.NoError(err)
}

func (s *StoreTestSuite) TestRecordClick_InvalidUpdateID() {
	err := s.store.RecordClick("notaninteger", nil)
	s.EqualError(err, `strconv.Atoi: parsing "notaninteger": invalid syntax`)
}

func (s *StoreTestSuite) TestRecordClick_InvalidIP() {
	ip := net.ParseIP("169.254.169.254")
	s.mip.On("GetIPInfo", ip).Return(GeoLoc{}, fmt.Errorf("invalid IP")).Once()
	err := s.store.RecordClick("12345", ip)
	s.EqualError(err, "invalid IP")
}

func (s *StoreTestSuite) TestRecordClick_ErrBeginTx() {
	ip := net.ParseIP("169.254.169.254")
	s.mip.On("GetIPInfo", ip).Return(GeoLoc{
		Country: "IE",
		Region: "L",
		City: "Dublin",
	}, nil).Once()
	s.mdb.ExpectBegin().WillReturnError(fmt.Errorf("some error"))
	err := s.store.RecordClick("12345", ip)
	s.EqualError(err, "some error")
}

func (s *StoreTestSuite) TestRecordClick_ErrExec() {
	ip := net.ParseIP("169.254.169.254")
	s.mip.On("GetIPInfo", ip).Return(GeoLoc{
		Country: "IE",
		Region: "L",
		City: "Dublin",
	}, nil).Once()
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(`INSERT INTO comic_clicks \(update_id, country, region, city\) VALUES \(\$1, \$2, \$3, \$4\);`).WithArgs(12345, "IE", "L", "Dublin").WillReturnError(fmt.Errorf("some error"))
	err := s.store.RecordClick("12345", ip)
	s.EqualError(err, "some error")
}

func (s *StoreTestSuite) TestRecordClick_ErrCommitTx() {
	ip := net.ParseIP("169.254.169.254")
	s.mip.On("GetIPInfo", ip).Return(GeoLoc{
		Country: "IE",
		Region: "L",
		City: "Dublin",
	}, nil).Once()
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(`INSERT INTO comic_clicks \(update_id, country, region, city\) VALUES \(\$1, \$2, \$3, \$4\);`).WithArgs(12345, "IE", "L", "Dublin").WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit().WillReturnError(fmt.Errorf("some error"))
	err := s.store.RecordClick("12345", ip)
	s.EqualError(err, "some error")
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
