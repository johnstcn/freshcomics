package store

import (
	"fmt"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"database/sql/driver"
)

var testSiteDef = SiteDef{
	Name: "Test Name",
	Active: true,
	NSFW: true,
	StartURL: "Test Start URL",
	LastCheckedAt: time.Unix(0, 0),
	URLTemplate: "Test Template",
	RefXpath: "Test Ref XPath",
	RefRegexp: "Test Ref Regexp",
	TitleXpath: "Test Title XPath",
	TitleRegexp: "Test Title Regexp",
}

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
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetComics)).WillReturnRows(rows)
	comics, err := s.store.GetComics()
	s.NotNil(comics)
	s.Len(comics, 1)
	s.EqualValues("Test Comic", comics[0].Name)
	s.NoError(err)
}

func (s *StoreTestSuite) TestGetComics_Err() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetComics)).WillReturnError(fmt.Errorf("some error"))
	comics, err := s.store.GetComics()
	s.Nil(comics)
	s.EqualError(err, "some error")
}

func (s *StoreTestSuite) TestGetRedirectURL_OK() {
	rows := sqlmock.NewRows([]string{"url"}).AddRow("http://example.com")
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetRedirectURL)).WithArgs("12345").WillReturnRows(rows)
	url, err := s.store.GetRedirectURL("12345")
	s.NoError(err)
	s.EqualValues("http://example.com", url)
}

func (s *StoreTestSuite) TestGetRedirectURL_Err() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetRedirectURL)).WithArgs("12345").WillReturnError(fmt.Errorf("some error"))
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
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlRecordClick)).WithArgs(12345, "IE", "L", "Dublin").WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit()
	err := s.store.RecordClick(12345, ip)
	s.NoError(err)
}

func (s *StoreTestSuite) TestRecordClick_InvalidIP() {
	ip := net.ParseIP("169.254.169.254")
	s.mip.On("GetIPInfo", ip).Return(GeoLoc{}, fmt.Errorf("invalid IP")).Once()
	err := s.store.RecordClick(12345, ip)
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
	err := s.store.RecordClick(12345, ip)
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
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlRecordClick)).WithArgs(12345, "IE", "L", "Dublin").WillReturnError(fmt.Errorf("some error"))
	err := s.store.RecordClick(12345, ip)
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
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlRecordClick)).WithArgs(12345, "IE", "L", "Dublin").WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit().WillReturnError(fmt.Errorf("some error"))
	err := s.store.RecordClick(12345, ip)
	s.EqualError(err, "some error")
}

func (s *StoreTestSuite) TestCreateSiteDef_OK() {
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	s.mdb.ExpectBegin()
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlCreateSiteDef)).WithArgs(testSiteDef.Name, testSiteDef.Active, testSiteDef.NSFW, testSiteDef.StartURL, testSiteDef.LastCheckedAt, testSiteDef.URLTemplate, testSiteDef.RefXpath, testSiteDef.RefRegexp, testSiteDef.TitleXpath, testSiteDef.TitleRegexp).WillReturnRows(rows)
	s.mdb.ExpectCommit()
	newid, err := s.store.CreateSiteDef(testSiteDef)
	s.NotZero(newid)
	s.NoError(err)
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
