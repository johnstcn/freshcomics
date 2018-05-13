package store

import (
	"database/sql/driver"
	"fmt"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/johnstcn/freshcomics/internal/ipinfo"
	"github.com/johnstcn/freshcomics/internal/ipinfo/ipinfotest"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var testSiteDefA = SiteDef{
	ID:            1,
	Name:          "Test Name",
	Active:        true,
	NSFW:          true,
	StartURL:      "Test Start URL",
	LastCheckedAt: time.Unix(1, 0),
	URLTemplate:   "Test Template",
	RefXpath:      "Test Ref XPath",
	RefRegexp:     "Test Ref Regexp",
	TitleXpath:    "Test Title XPath",
	TitleRegexp:   "Test Title Regexp",
}

var testSiteDefB = SiteDef{
	ID:            1,
	Name:          "Test Name Other",
	Active:        false,
	NSFW:          false,
	StartURL:      "Test Start URL Other",
	LastCheckedAt: time.Unix(0, 0),
	URLTemplate:   "Test Template Other",
	RefXpath:      "Test Ref XPath Other",
	RefRegexp:     "Test Ref Regexp Other",
	TitleXpath:    "Test Title XPath Other",
	TitleRegexp:   "Test Title Regexp Other",
}

var testSiteUpdateA = SiteUpdate{
	ID:        1,
	SiteDefID: 1,
	Ref:       "Test Ref",
	URL:       "Test URL",
	Title:     "Test Title",
	SeenAt:    time.Unix(0, 0),
}

var testCrawlInfoA = CrawlInfo{
	ID:        1,
	SiteDefID: 1,
	StartedAt: time.Unix(0, 0),
	EndedAt:   time.Unix(1, 0),
	Seen:      1,
	Status:    CrawlStatusOK,
}

var testError = fmt.Errorf("some error")

type PGStoreTestSuite struct {
	suite.Suite
	store *pgStore
	mip   *ipinfotest.IPInfoer
	mdb   sqlmock.Sqlmock
	now   func() time.Time
}

func (s *PGStoreTestSuite) SetupSuite() {
	conn, mdb, err := sqlmock.New()
	if err != nil {
		s.Fail(err.Error())
	}
	s.mdb = mdb
	s.mip = &ipinfotest.IPInfoer{}
	s.store = &pgStore{
		db:    sqlx.NewDb(conn, "sqlmock"),
		geoIP: s.mip,
	}
	s.now = func() time.Time {
		return time.Unix(1234, 0)
	}
}

func (s *PGStoreTestSuite) TearDownTest() {
	s.NoError(s.mdb.ExpectationsWereMet())
	s.mip.AssertExpectations(s.T())
}

func (s *PGStoreTestSuite) TestGetComics_OK() {
	rows := sqlmock.NewRows([]string{"name", "nsfw", "id", "title", "seen_at"}).AddRow("Test Comic", false, 1, "Test Title", s.now())
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetComics)).WillReturnRows(rows)
	comics, err := s.store.GetComics()
	s.NotNil(comics)
	s.Len(comics, 1)
	s.EqualValues("Test Comic", comics[0].Name)
	s.NoError(err)
}

func (s *PGStoreTestSuite) TestGetComics_Err() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetComics)).WillReturnError(testError)
	comics, err := s.store.GetComics()
	s.Nil(comics)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestGetRedirectURL_OK() {
	rows := sqlmock.NewRows([]string{"url"}).AddRow("http://example.com")
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetRedirectURL)).WithArgs("12345").WillReturnRows(rows)
	url, err := s.store.GetRedirectURL("12345")
	s.NoError(err)
	s.EqualValues("http://example.com", url)
}

func (s *PGStoreTestSuite) TestGetRedirectURL_Err() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetRedirectURL)).WithArgs("12345").WillReturnError(testError)
	url, err := s.store.GetRedirectURL("12345")
	s.Zero(url)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestRecordClick_OK() {
	ip := net.ParseIP("169.254.169.254")
	s.mip.On("GetIPInfo", ip).Return(ipinfo.GeoLoc{
		Country: "IE",
		Region:  "L",
		City:    "Dublin",
	}, nil).Once()
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlRecordClick)).WithArgs(12345, "IE", "L", "Dublin").WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit()
	err := s.store.RecordClick(12345, ip)
	s.NoError(err)
}

func (s *PGStoreTestSuite) TestRecordClick_InvalidIP() {
	ip := net.ParseIP("169.254.169.254")
	s.mip.On("GetIPInfo", ip).Return(ipinfo.GeoLoc{}, testError).Once()
	err := s.store.RecordClick(12345, ip)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestRecordClick_ErrBeginTx() {
	ip := net.ParseIP("169.254.169.254")
	s.mip.On("GetIPInfo", ip).Return(ipinfo.GeoLoc{
		Country: "IE",
		Region:  "L",
		City:    "Dublin",
	}, nil).Once()
	s.mdb.ExpectBegin().WillReturnError(testError)
	err := s.store.RecordClick(12345, ip)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestRecordClick_ErrExec() {
	ip := net.ParseIP("169.254.169.254")
	s.mip.On("GetIPInfo", ip).Return(ipinfo.GeoLoc{
		Country: "IE",
		Region:  "L",
		City:    "Dublin",
	}, nil).Once()
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlRecordClick)).WithArgs(12345, "IE", "L", "Dublin").WillReturnError(testError)
	err := s.store.RecordClick(12345, ip)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestRecordClick_ErrCommitTx() {
	ip := net.ParseIP("169.254.169.254")
	s.mip.On("GetIPInfo", ip).Return(ipinfo.GeoLoc{
		Country: "IE",
		Region:  "L",
		City:    "Dublin",
	}, nil).Once()
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlRecordClick)).WithArgs(12345, "IE", "L", "Dublin").WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit().WillReturnError(testError)
	err := s.store.RecordClick(12345, ip)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestCreateSiteDef_OK() {
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	s.mdb.ExpectBegin()
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlCreateSiteDef)).WithArgs(testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.LastCheckedAt, testSiteDefA.URLTemplate, testSiteDefA.RefXpath, testSiteDefA.RefRegexp, testSiteDefA.TitleXpath, testSiteDefA.TitleRegexp).WillReturnRows(rows)
	s.mdb.ExpectCommit()
	newid, err := s.store.CreateSiteDef(testSiteDefA)
	s.NotZero(newid)
	s.NoError(err)
}

func (s *PGStoreTestSuite) TestCreateSiteDef_ErrBegin() {
	s.mdb.ExpectBegin().WillReturnError(testError)
	newid, err := s.store.CreateSiteDef(testSiteDefA)
	s.EqualValues(-1, newid)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestCreateSiteDef_ErrQuery() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlCreateSiteDef)).WithArgs(testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.LastCheckedAt, testSiteDefA.URLTemplate, testSiteDefA.RefXpath, testSiteDefA.RefRegexp, testSiteDefA.TitleXpath, testSiteDefA.TitleRegexp).WillReturnError(testError)
	newid, err := s.store.CreateSiteDef(testSiteDefA)
	s.EqualValues(-1, newid)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestCreateSiteDef_ErrCommit() {
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	s.mdb.ExpectBegin()
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlCreateSiteDef)).WithArgs(testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.LastCheckedAt, testSiteDefA.URLTemplate, testSiteDefA.RefXpath, testSiteDefA.RefRegexp, testSiteDefA.TitleXpath, testSiteDefA.TitleRegexp).WillReturnRows(rows)
	s.mdb.ExpectCommit().WillReturnError(testError)
	newid, err := s.store.CreateSiteDef(testSiteDefA)
	s.EqualValues(-1, newid)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestGetAllSiteDefs_OK() {
	rows := sqlmock.NewRows([]string{"id", "name", "active", "nsfw", "start_url", "last_checked_at", "url_template", "ref_xpath", "ref_regexp", "title_xpath", "title_regexp"})
	rows.AddRow(testSiteDefA.ID, testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.LastCheckedAt, testSiteDefA.URLTemplate, testSiteDefA.RefXpath, testSiteDefA.RefRegexp, testSiteDefA.TitleXpath, testSiteDefA.TitleRegexp)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetAllSiteDefsActive)).WillReturnRows(rows)
	defs, err := s.store.GetAllSiteDefs(false)
	s.NoError(err)
	s.Len(defs, 1)
	s.EqualValues(testSiteDefA, defs[0])
}

func (s *PGStoreTestSuite) TestGetAllSiteDefsInActive_OK() {
	rows := sqlmock.NewRows([]string{"id", "name", "active", "nsfw", "start_url", "last_checked_at", "url_template", "ref_xpath", "ref_regexp", "title_xpath", "title_regexp"})
	rows.AddRow(testSiteDefA.ID, testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.LastCheckedAt, testSiteDefA.URLTemplate, testSiteDefA.RefXpath, testSiteDefA.RefRegexp, testSiteDefA.TitleXpath, testSiteDefA.TitleRegexp)
	rows.AddRow(testSiteDefB.ID, testSiteDefB.Name, testSiteDefB.Active, testSiteDefB.NSFW, testSiteDefB.StartURL, testSiteDefB.LastCheckedAt, testSiteDefB.URLTemplate, testSiteDefB.RefXpath, testSiteDefB.RefRegexp, testSiteDefB.TitleXpath, testSiteDefB.TitleRegexp)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetAllSiteDefs)).WillReturnRows(rows)
	defs, err := s.store.GetAllSiteDefs(true)
	s.NoError(err)
	s.Len(defs, 2)
	s.EqualValues(testSiteDefA, defs[0])
	s.EqualValues(testSiteDefB, defs[1])
}

func (s *PGStoreTestSuite) TestGetAllSiteDefsNoRows_OK() {
	rows := sqlmock.NewRows([]string{"id", "name", "active", "nsfw", "start_url", "last_checked_at", "url_template", "ref_xpath", "ref_regexp", "title_xpath", "title_regexp"})
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetAllSiteDefs)).WillReturnRows(rows)
	defs, err := s.store.GetAllSiteDefs(true)
	s.NoError(err)
	s.Len(defs, 0)
}

func (s *PGStoreTestSuite) TestGetAllSiteDefs_Err() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetAllSiteDefs)).WillReturnError(testError)
	defs, err := s.store.GetAllSiteDefs(true)
	s.EqualError(err, "some error")
	s.Nil(defs)
}

func (s *PGStoreTestSuite) TestGetSiteDefByID_OK() {
	rows := sqlmock.NewRows([]string{"id", "name", "active", "nsfw", "start_url", "last_checked_at", "url_template", "ref_xpath", "ref_regexp", "title_xpath", "title_regexp"})
	rows.AddRow(testSiteDefA.ID, testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.LastCheckedAt, testSiteDefA.URLTemplate, testSiteDefA.RefXpath, testSiteDefA.RefRegexp, testSiteDefA.TitleXpath, testSiteDefA.TitleRegexp)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteDefByID)).WithArgs(1).WillReturnRows(rows)
	def, err := s.store.GetSiteDefByID(1)
	s.NoError(err)
	s.EqualValues(testSiteDefA, def)
}

func (s *PGStoreTestSuite) TestGetSiteDefByID_Err() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteDefByID)).WillReturnError(testError)
	def, err := s.store.GetSiteDefByID(1)
	s.EqualError(err, "some error")
	s.Zero(def)
}

func (s *PGStoreTestSuite) TestGetSiteDefLastChecked_OK() {
	rows := sqlmock.NewRows([]string{"id", "name", "active", "nsfw", "start_url", "last_checked_at", "url_template", "ref_xpath", "ref_regexp", "title_xpath", "title_regexp"})
	rows.AddRow(testSiteDefA.ID, testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.LastCheckedAt, testSiteDefA.URLTemplate, testSiteDefA.RefXpath, testSiteDefA.RefRegexp, testSiteDefA.TitleXpath, testSiteDefA.TitleRegexp)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteDefLastChecked)).WillReturnRows(rows)
	def, err := s.store.GetSiteDefLastChecked()
	s.NoError(err)
	s.EqualValues(testSiteDefA, def)
}

func (s *PGStoreTestSuite) TestGetSiteDefLastChecked_Err() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteDefLastChecked)).WillReturnError(testError)
	def, err := s.store.GetSiteDefLastChecked()
	s.EqualError(err, "some error")
	s.Zero(def)
}

func (s *PGStoreTestSuite) TestSaveSiteDef_OK() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlSaveSiteDef)).WithArgs(testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.LastCheckedAt, testSiteDefA.URLTemplate, testSiteDefA.RefXpath, testSiteDefA.RefRegexp, testSiteDefA.TitleXpath, testSiteDefA.TitleRegexp, testSiteDefA.ID).WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit()
	err := s.store.SaveSiteDef(testSiteDefA)
	s.NoError(err)
}

func (s *PGStoreTestSuite) TestSaveSiteDef_ErrBegin() {
	s.mdb.ExpectBegin().WillReturnError(testError)
	err := s.store.SaveSiteDef(testSiteDefA)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestSaveSiteDef_ErrExec() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlSaveSiteDef)).WithArgs(testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.LastCheckedAt, testSiteDefA.URLTemplate, testSiteDefA.RefXpath, testSiteDefA.RefRegexp, testSiteDefA.TitleXpath, testSiteDefA.TitleRegexp, testSiteDefA.ID).WillReturnError(testError)
	err := s.store.SaveSiteDef(testSiteDefA)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestSaveSiteDef_ErrCommit() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlSaveSiteDef)).WithArgs(testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.LastCheckedAt, testSiteDefA.URLTemplate, testSiteDefA.RefXpath, testSiteDefA.RefRegexp, testSiteDefA.TitleXpath, testSiteDefA.TitleRegexp, testSiteDefA.ID).WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit().WillReturnError(testError)
	err := s.store.SaveSiteDef(testSiteDefA)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestSetSiteDefLastChecked_OK() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlSetSiteDefLastChecked)).WithArgs(s.now(), 1).WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit()
	err := s.store.SetSiteDefLastChecked(testSiteDefA, s.now())
	s.NoError(err)
}

func (s *PGStoreTestSuite) TestSetSiteDefLastChecked_ErrBegin() {
	s.mdb.ExpectBegin().WillReturnError(testError)
	err := s.store.SetSiteDefLastChecked(testSiteDefA, s.now())
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestSetSiteDefLastChecked_ErrExec() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlSetSiteDefLastChecked)).WithArgs(s.now(), 1).WillReturnError(testError)
	err := s.store.SetSiteDefLastChecked(testSiteDefA, s.now())
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestSetSiteDefLastChecked_ErrCommit() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlSetSiteDefLastChecked)).WithArgs(s.now(), 1).WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit().WillReturnError(testError)
	err := s.store.SetSiteDefLastChecked(testSiteDefA, s.now())
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestCreateSiteUpdate_OK() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlCreateSiteUpdate)).WithArgs(testSiteUpdateA.SiteDefID, testSiteUpdateA.Ref, testSiteUpdateA.URL, testSiteUpdateA.Title, testSiteUpdateA.SeenAt).WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit()
	err := s.store.CreateSiteUpdate(testSiteUpdateA)
	s.NoError(err)
}

func (s *PGStoreTestSuite) TestCreateSiteUpdate_ErrBegin() {
	s.mdb.ExpectBegin().WillReturnError(testError)
	err := s.store.CreateSiteUpdate(testSiteUpdateA)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestCreateSiteUpdate_ErrExec() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlCreateSiteUpdate)).WithArgs(testSiteUpdateA.SiteDefID, testSiteUpdateA.Ref, testSiteUpdateA.URL, testSiteUpdateA.Title, testSiteUpdateA.SeenAt).WillReturnError(testError)
	err := s.store.CreateSiteUpdate(testSiteUpdateA)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestCreateSiteUpdate_ErrCommit() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlCreateSiteUpdate)).WithArgs(testSiteUpdateA.SiteDefID, testSiteUpdateA.Ref, testSiteUpdateA.URL, testSiteUpdateA.Title, testSiteUpdateA.SeenAt).WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit().WillReturnError(testError)
	err := s.store.CreateSiteUpdate(testSiteUpdateA)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestGetSiteUpdatesBySiteDefID_OK() {
	rows := sqlmock.NewRows([]string{"id", "site_def_id", "ref", "url", "title", "seen_at"})
	rows.AddRow(testSiteUpdateA.ID, testSiteUpdateA.SiteDefID, testSiteUpdateA.Ref, testSiteUpdateA.URL, testSiteUpdateA.Title, testSiteUpdateA.SeenAt)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteUpdatesBySiteDefID)).WithArgs(testSiteUpdateA.SiteDefID).WillReturnRows(rows)
	updates, err := s.store.GetSiteUpdatesBySiteDefID(testSiteUpdateA.SiteDefID)
	s.NoError(err)
	s.Len(updates, 1)
	s.EqualValues(updates[0], testSiteUpdateA)
}

func (s *PGStoreTestSuite) TestGetSiteUpdatesBySiteDefID_OKNoRows() {
	rows := sqlmock.NewRows([]string{"id", "site_def_id", "ref", "url", "title", "seen_at"})
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteUpdatesBySiteDefID)).WithArgs(testSiteUpdateA.SiteDefID).WillReturnRows(rows)
	updates, err := s.store.GetSiteUpdatesBySiteDefID(testSiteDefB.ID)
	s.NoError(err)
	s.Len(updates, 0)
}

func (s *PGStoreTestSuite) TestGetSiteUpdatesBySiteDefID_ErrQuery() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteUpdatesBySiteDefID)).WithArgs(testSiteUpdateA.SiteDefID).WillReturnError(testError)
	updates, err := s.store.GetSiteUpdatesBySiteDefID(testSiteUpdateA.SiteDefID)
	s.EqualError(err, "some error")
	s.Len(updates, 0)
}

func (s *PGStoreTestSuite) TestGetSiteUpdateBySiteDefAndRef_OK() {
	rows := sqlmock.NewRows([]string{"id", "site_def_id", "ref", "url", "title", "seen_at"})
	rows.AddRow(testSiteUpdateA.ID, testSiteUpdateA.SiteDefID, testSiteUpdateA.Ref, testSiteUpdateA.URL, testSiteUpdateA.Title, testSiteUpdateA.SeenAt)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteUpdateBySiteDefAndRef)).WillReturnRows(rows)
	su, err := s.store.GetSiteUpdateBySiteDefAndRef(testSiteDefA.ID, testSiteUpdateA.Ref)
	s.NoError(err)
	s.EqualValues(testSiteUpdateA, su)
}

func (s *PGStoreTestSuite) TestGetSiteUpdateBySiteDefAndRef_ErrQuery() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteUpdateBySiteDefAndRef)).WillReturnError(testError)
	su, err := s.store.GetSiteUpdateBySiteDefAndRef(testSiteDefA.ID, testSiteUpdateA.Ref)
	s.EqualError(err, "some error")
	s.Zero(su)
}

func (s *PGStoreTestSuite) TestGetStartURLForCrawl_OK() {
	rows := sqlmock.NewRows([]string{"url"})
	rows.AddRow(testSiteUpdateA.URL)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetStartURLForCrawl)).WillReturnRows(rows)
	url, err := s.store.GetStartURLForCrawl(testSiteDefA)
	s.NoError(err)
	s.EqualValues(testSiteUpdateA.URL, url)
}

func (s *PGStoreTestSuite) TestGetStartURLForCrawl_ErrQuery() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetStartURLForCrawl)).WillReturnError(testError)
	url, err := s.store.GetStartURLForCrawl(testSiteDefA)
	s.Zero(url)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestGetCrawlInfo_OK() {
	rows := sqlmock.NewRows([]string{"id", "site_def_id", "started_at", "ended_at", "status", "seen"})
	rows.AddRow(testCrawlInfoA.ID, testCrawlInfoA.SiteDefID, testCrawlInfoA.StartedAt, testCrawlInfoA.EndedAt, testCrawlInfoA.Status, testCrawlInfoA.Seen)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetCrawlInfo)).WillReturnRows(rows)
	ci, err := s.store.GetCrawlInfo()
	s.NoError(err)
	s.Len(ci, 1)
	s.EqualValues(ci[0], testCrawlInfoA)
}

func (s *PGStoreTestSuite) TestGetCrawlInfo_ErrQuery() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetCrawlInfo)).WillReturnError(testError)
	ci, err := s.store.GetCrawlInfo()
	s.Len(ci, 0)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestGetCrawlInfoBySiteDefID_OK() {
	rows := sqlmock.NewRows([]string{"id", "site_def_id", "started_at", "ended_at", "status", "seen"})
	rows.AddRow(testCrawlInfoA.ID, testCrawlInfoA.SiteDefID, testCrawlInfoA.StartedAt, testCrawlInfoA.EndedAt, testCrawlInfoA.Status, testCrawlInfoA.Seen)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetCrawlInfoBySiteDefID)).WillReturnRows(rows)
	ci, err := s.store.GetCrawlInfoBySiteDefID(1)
	s.NoError(err)
	s.Len(ci, 1)
	s.EqualValues(ci[0], testCrawlInfoA)
}

func (s *PGStoreTestSuite) TestGetCrawlInfoBySiteDefID_ErrQuery() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetCrawlInfoBySiteDefID)).WillReturnError(testError)
	ci, err := s.store.GetCrawlInfoBySiteDefID(1)
	s.Len(ci, 0)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestCreateCrawlInfo_OK() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlCreateCrawlInfo)).WithArgs(testCrawlInfoA.SiteDefID, testCrawlInfoA.StartedAt, testCrawlInfoA.EndedAt, testCrawlInfoA.Status, testCrawlInfoA.Seen).WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit()
	err := s.store.CreateCrawlInfo(testCrawlInfoA)
	s.NoError(err)
}

func (s *PGStoreTestSuite) TestCreateCrawlInfo_ErrBegin() {
	s.mdb.ExpectBegin().WillReturnError(testError)
	err := s.store.CreateCrawlInfo(testCrawlInfoA)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestCreateCrawlInfo_ErrExec() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlCreateCrawlInfo)).WithArgs(testCrawlInfoA.SiteDefID, testCrawlInfoA.StartedAt, testCrawlInfoA.EndedAt, testCrawlInfoA.Status, testCrawlInfoA.Seen).WillReturnError(testError)
	err := s.store.CreateCrawlInfo(testCrawlInfoA)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestCreateCrawlInfo_ErrCommit() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlCreateCrawlInfo)).WithArgs(testCrawlInfoA.SiteDefID, testCrawlInfoA.StartedAt, testCrawlInfoA.EndedAt, testCrawlInfoA.Status, testCrawlInfoA.Seen).WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit().WillReturnError(testError)
	err := s.store.CreateCrawlInfo(testCrawlInfoA)
	s.EqualError(err, "some error")
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(PGStoreTestSuite))
}