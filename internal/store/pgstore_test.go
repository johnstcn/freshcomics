package store

import (
	"database/sql/driver"
	"fmt"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/lib/pq"

	"github.com/johnstcn/freshcomics/internal/ipinfo"
	"github.com/johnstcn/freshcomics/internal/ipinfo/ipinfotest"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var testSiteDefA = SiteDef{
	ID:            SiteDefID(1),
	Name:          "Test Name",
	Active:        true,
	NSFW:          true,
	StartURL:      "Test Start URL",
	URLTemplate:   "Test Template",
	NextPageXPath: "Test Ref XPath",
	RefRegexp:     "Test Ref Regexp",
	TitleXPath:    "Test Title XPath",
	TitleRegexp:   "Test Title Regexp",
}

var testSiteDefB = SiteDef{
	ID:            SiteDefID(2),
	Name:          "Test Name Other",
	Active:        false,
	NSFW:          false,
	StartURL:      "Test Start URL Other",
	URLTemplate:   "Test Template Other",
	NextPageXPath: "Test Ref XPath Other",
	RefRegexp:     "Test Ref Regexp Other",
	TitleXPath:    "Test Title XPath Other",
	TitleRegexp:   "Test Title Regexp Other",
}

var testSiteUpdateA = SiteUpdate{
	ID:        SiteUpdateID(1),
	SiteDefID: SiteDefID(1),
	Ref:       "Test Ref",
	URL:       "Test URL",
	Title:     "Test Title",
	SeenAt:    time.Unix(0, 0),
}

var testCrawlInfoA = CrawlInfo{
	ID:        CrawlInfoID(1),
	SiteDefID: SiteDefID(1),
	URL:       "http://example.com",
	StartedAt: pq.NullTime{
		Time:  time.Unix(0, 0).UTC(),
		Valid: true,
	},
	EndedAt: pq.NullTime{
		Time:  time.Unix(1, 0).UTC(),
		Valid: true,
	},
	Seen:  1,
	Error: "",
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
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlRedirect)).WithArgs(testSiteUpdateA.ID).WillReturnRows(rows)
	url, err := s.store.Redirect(testSiteUpdateA.ID)
	s.NoError(err)
	s.EqualValues("http://example.com", url)
}

func (s *PGStoreTestSuite) TestGetRedirectURL_Err() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlRedirect)).WithArgs(testSiteUpdateA.ID).WillReturnError(testError)
	url, err := s.store.Redirect(testSiteUpdateA.ID)
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
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlSaveClick)).WithArgs(12345, "IE", "L", "Dublin").WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit()
	err := s.store.CreateClickLog(12345, ip)
	s.NoError(err)
}

func (s *PGStoreTestSuite) TestRecordClick_InvalidIP() {
	ip := net.ParseIP("169.254.169.254")
	s.mip.On("GetIPInfo", ip).Return(ipinfo.GeoLoc{}, testError).Once()
	err := s.store.CreateClickLog(12345, ip)
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
	err := s.store.CreateClickLog(12345, ip)
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
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlSaveClick)).WithArgs(12345, "IE", "L", "Dublin").WillReturnError(testError)
	err := s.store.CreateClickLog(12345, ip)
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
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlSaveClick)).WithArgs(12345, "IE", "L", "Dublin").WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit().WillReturnError(testError)
	err := s.store.CreateClickLog(12345, ip)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestCreateSiteDef_OK() {
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	s.mdb.ExpectBegin()
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlCreateSiteDef)).WithArgs(testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.URLTemplate, testSiteDefA.NextPageXPath, testSiteDefA.RefRegexp, testSiteDefA.TitleXPath, testSiteDefA.TitleRegexp).WillReturnRows(rows)
	s.mdb.ExpectCommit()
	newID, err := s.store.CreateSiteDef(testSiteDefA)
	s.EqualValues(1, newID)
	s.NoError(err)
}

func (s *PGStoreTestSuite) TestCreateSiteDef_ErrBegin() {
	s.mdb.ExpectBegin().WillReturnError(testError)
	newID, err := s.store.CreateSiteDef(testSiteDefA)
	s.Zero(newID)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestCreateSiteDef_ErrQuery() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlCreateSiteDef)).WithArgs(testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.URLTemplate, testSiteDefA.NextPageXPath, testSiteDefA.RefRegexp, testSiteDefA.TitleXPath, testSiteDefA.TitleRegexp).WillReturnError(testError)
	newID, err := s.store.CreateSiteDef(testSiteDefA)
	s.Zero(newID)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestCreateSiteDef_ErrCommit() {
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	s.mdb.ExpectBegin()
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlCreateSiteDef)).WithArgs(testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.URLTemplate, testSiteDefA.NextPageXPath, testSiteDefA.RefRegexp, testSiteDefA.TitleXPath, testSiteDefA.TitleRegexp).WillReturnRows(rows)
	s.mdb.ExpectCommit().WillReturnError(testError)
	newID, err := s.store.CreateSiteDef(testSiteDefA)
	s.Zero(newID)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestGetAllSiteDefs_OK() {
	rows := sqlmock.NewRows([]string{"id", "name", "active", "nsfw", "start_url", "url_template", "next_page_xpath", "ref_regexp", "title_xpath", "title_regexp"})
	rows.AddRow(testSiteDefA.ID, testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.URLTemplate, testSiteDefA.NextPageXPath, testSiteDefA.RefRegexp, testSiteDefA.TitleXPath, testSiteDefA.TitleRegexp)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetActiveSiteDefs)).WillReturnRows(rows)
	defs, err := s.store.GetSiteDefs(false)
	s.NoError(err)
	s.Len(defs, 1)
	s.EqualValues(testSiteDefA, defs[0])
}

func (s *PGStoreTestSuite) TestGetAllSiteDefsInActive_OK() {
	rows := sqlmock.NewRows([]string{"id", "name", "active", "nsfw", "start_url", "url_template", "next_page_xpath", "ref_regexp", "title_xpath", "title_regexp"})
	rows.AddRow(testSiteDefA.ID, testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.URLTemplate, testSiteDefA.NextPageXPath, testSiteDefA.RefRegexp, testSiteDefA.TitleXPath, testSiteDefA.TitleRegexp)
	rows.AddRow(testSiteDefB.ID, testSiteDefB.Name, testSiteDefB.Active, testSiteDefB.NSFW, testSiteDefB.StartURL, testSiteDefB.URLTemplate, testSiteDefB.NextPageXPath, testSiteDefB.RefRegexp, testSiteDefB.TitleXPath, testSiteDefB.TitleRegexp)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteDefs)).WillReturnRows(rows)
	defs, err := s.store.GetSiteDefs(true)
	s.NoError(err)
	s.Len(defs, 2)
	s.EqualValues(testSiteDefA, defs[0])
	s.EqualValues(testSiteDefB, defs[1])
}

func (s *PGStoreTestSuite) TestGetAllSiteDefsNoRows_OK() {
	rows := sqlmock.NewRows([]string{"id", "name", "active", "nsfw", "start_url", "url_template", "next_page_xpath", "ref_regexp", "title_xpath", "title_regexp"})
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteDefs)).WillReturnRows(rows)
	defs, err := s.store.GetSiteDefs(true)
	s.NoError(err)
	s.Len(defs, 0)
}

func (s *PGStoreTestSuite) TestGetAllSiteDefs_Err() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteDefs)).WillReturnError(testError)
	defs, err := s.store.GetSiteDefs(true)
	s.EqualError(err, "some error")
	s.Nil(defs)
}

func (s *PGStoreTestSuite) TestGetSiteDefByID_OK() {
	rows := sqlmock.NewRows([]string{"id", "name", "active", "nsfw", "start_url", "url_template", "next_page_xpath", "ref_regexp", "title_xpath", "title_regexp"})
	rows.AddRow(testSiteDefA.ID, testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.URLTemplate, testSiteDefA.NextPageXPath, testSiteDefA.RefRegexp, testSiteDefA.TitleXPath, testSiteDefA.TitleRegexp)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteDef)).WithArgs(1).WillReturnRows(rows)
	def, err := s.store.GetSiteDef(1)
	s.NoError(err)
	s.EqualValues(testSiteDefA, def)
}

func (s *PGStoreTestSuite) TestGetSiteDefByID_Err() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteDef)).WillReturnError(testError)
	def, err := s.store.GetSiteDef(1)
	s.EqualError(err, "some error")
	s.Zero(def)
}

func (s *PGStoreTestSuite) TestSaveSiteDef_OK() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlUpdateSiteDef)).WithArgs(testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.URLTemplate, testSiteDefA.NextPageXPath, testSiteDefA.RefRegexp, testSiteDefA.TitleXPath, testSiteDefA.TitleRegexp, testSiteDefA.ID).WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit()
	err := s.store.UpdateSiteDef(testSiteDefA)
	s.NoError(err)
}

func (s *PGStoreTestSuite) TestSaveSiteDef_ErrBegin() {
	s.mdb.ExpectBegin().WillReturnError(testError)
	err := s.store.UpdateSiteDef(testSiteDefA)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestSaveSiteDef_ErrExec() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlUpdateSiteDef)).WithArgs(testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.URLTemplate, testSiteDefA.NextPageXPath, testSiteDefA.RefRegexp, testSiteDefA.TitleXPath, testSiteDefA.TitleRegexp, testSiteDefA.ID).WillReturnError(testError)
	err := s.store.UpdateSiteDef(testSiteDefA)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestSaveSiteDef_ErrCommit() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlUpdateSiteDef)).WithArgs(testSiteDefA.Name, testSiteDefA.Active, testSiteDefA.NSFW, testSiteDefA.StartURL, testSiteDefA.URLTemplate, testSiteDefA.NextPageXPath, testSiteDefA.RefRegexp, testSiteDefA.TitleXPath, testSiteDefA.TitleRegexp, testSiteDefA.ID).WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit().WillReturnError(testError)
	err := s.store.UpdateSiteDef(testSiteDefA)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestCreateSiteUpdate_OK() {
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	s.mdb.ExpectBegin()
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlCreateSiteUpdate)).WithArgs(testSiteUpdateA.SiteDefID, testSiteUpdateA.Ref, testSiteUpdateA.URL, testSiteUpdateA.Title, testSiteUpdateA.SeenAt).WillReturnRows(rows)
	s.mdb.ExpectCommit()
	newID, err := s.store.CreateSiteUpdate(testSiteUpdateA)
	s.NoError(err)
	s.EqualValues(1, newID)
}

func (s *PGStoreTestSuite) TestCreateSiteUpdate_ErrBegin() {
	s.mdb.ExpectBegin().WillReturnError(testError)
	newID, err := s.store.CreateSiteUpdate(testSiteUpdateA)
	s.EqualError(err, "some error")
	s.Zero(newID)
}

func (s *PGStoreTestSuite) TestCreateSiteUpdate_ErrExec() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlCreateSiteUpdate)).WithArgs(testSiteUpdateA.SiteDefID, testSiteUpdateA.Ref, testSiteUpdateA.URL, testSiteUpdateA.Title, testSiteUpdateA.SeenAt).WillReturnError(testError)
	newID, err := s.store.CreateSiteUpdate(testSiteUpdateA)
	s.EqualError(err, "some error")
	s.Zero(newID)
}

func (s *PGStoreTestSuite) TestCreateSiteUpdate_ErrCommit() {
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	s.mdb.ExpectBegin()
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlCreateSiteUpdate)).WithArgs(testSiteUpdateA.SiteDefID, testSiteUpdateA.Ref, testSiteUpdateA.URL, testSiteUpdateA.Title, testSiteUpdateA.SeenAt).WillReturnRows(rows)
	s.mdb.ExpectCommit().WillReturnError(testError)
	newID, err := s.store.CreateSiteUpdate(testSiteUpdateA)
	s.EqualError(err, "some error")
	s.Zero(newID)
}

func (s *PGStoreTestSuite) TestGetSiteUpdates_OK() {
	rows := sqlmock.NewRows([]string{"id", "site_def_id", "ref", "url", "title", "seen_at"})
	rows.AddRow(testSiteUpdateA.ID, testSiteUpdateA.SiteDefID, testSiteUpdateA.Ref, testSiteUpdateA.URL, testSiteUpdateA.Title, testSiteUpdateA.SeenAt)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteUpdates)).WithArgs(testSiteUpdateA.SiteDefID).WillReturnRows(rows)
	updates, err := s.store.GetSiteUpdates(testSiteUpdateA.SiteDefID)
	s.NoError(err)
	s.Len(updates, 1)
	s.EqualValues(updates[0], testSiteUpdateA)
}

func (s *PGStoreTestSuite) TestGetSiteUpdates_OKNoRows() {
	rows := sqlmock.NewRows([]string{"id", "site_def_id", "ref", "url", "title", "seen_at"})
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteUpdates)).WithArgs(testSiteUpdateA.SiteDefID).WillReturnRows(rows)
	updates, err := s.store.GetSiteUpdates(testSiteDefA.ID)
	s.NoError(err)
	s.Len(updates, 0)
}

func (s *PGStoreTestSuite) TestGetSiteUpdates_ErrQuery() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteUpdates)).WithArgs(testSiteUpdateA.SiteDefID).WillReturnError(testError)
	updates, err := s.store.GetSiteUpdates(testSiteUpdateA.SiteDefID)
	s.EqualError(err, "some error")
	s.Len(updates, 0)
}

func (s *PGStoreTestSuite) TestGetSiteUpdate_OK() {
	rows := sqlmock.NewRows([]string{"id", "site_def_id", "ref", "url", "title", "seen_at"})
	rows.AddRow(testSiteUpdateA.ID, testSiteUpdateA.SiteDefID, testSiteUpdateA.Ref, testSiteUpdateA.URL, testSiteUpdateA.Title, testSiteUpdateA.SeenAt)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteUpdate)).WillReturnRows(rows)
	su, err := s.store.GetSiteUpdate(testSiteDefA.ID, testSiteUpdateA.Ref)
	s.NoError(err)
	s.EqualValues(testSiteUpdateA, su)
}

func (s *PGStoreTestSuite) TestGetSiteUpdate_ErrQuery() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetSiteUpdate)).WillReturnError(testError)
	su, err := s.store.GetSiteUpdate(testSiteDefA.ID, testSiteUpdateA.Ref)
	s.EqualError(err, "some error")
	s.Zero(su)
}

func (s *PGStoreTestSuite) TestGetLastURL_OK() {
	rows := sqlmock.NewRows([]string{"url"})
	rows.AddRow(testSiteUpdateA.URL)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetLastURL)).WithArgs(testSiteDefA.ID).WillReturnRows(rows)
	url, err := s.store.GetLastURL(testSiteDefA.ID)
	s.NoError(err)
	s.EqualValues(testSiteUpdateA.URL, url)
}

func (s *PGStoreTestSuite) TestGetLastURL_ErrQuery() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetLastURL)).WithArgs(testSiteDefA.ID).WillReturnError(testError)
	url, err := s.store.GetLastURL(testSiteDefA.ID)
	s.Zero(url)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestGetCrawlInfos_OK() {
	rows := sqlmock.NewRows([]string{"id", "site_def_id", "url", "started_at", "ended_at", "error", "seen"})
	rows.AddRow(testCrawlInfoA.ID, testCrawlInfoA.SiteDefID, testCrawlInfoA.URL, testCrawlInfoA.StartedAt.Time, testCrawlInfoA.EndedAt.Time, testCrawlInfoA.Error, testCrawlInfoA.Seen)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetCrawlInfos)).WillReturnRows(rows)
	ci, err := s.store.GetCrawlInfos()
	s.NoError(err)
	s.Len(ci, 1)
	s.EqualValues(ci[0], testCrawlInfoA)
}

func (s *PGStoreTestSuite) TestGetCrawlInfos_ErrQuery() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetCrawlInfos)).WillReturnError(testError)
	ci, err := s.store.GetCrawlInfos()
	s.Len(ci, 0)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestGetCrawlInfo_OK() {
	rows := sqlmock.NewRows([]string{"id", "site_def_id", "url", "started_at", "ended_at", "error", "seen"})
	rows.AddRow(testCrawlInfoA.ID, testCrawlInfoA.SiteDefID, testCrawlInfoA.URL, testCrawlInfoA.StartedAt.Time, testCrawlInfoA.EndedAt.Time, testCrawlInfoA.Error, testCrawlInfoA.Seen)
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetCrawlInfo)).WillReturnRows(rows)
	ci, err := s.store.GetCrawlInfo(1)
	s.NoError(err)
	s.Len(ci, 1)
	s.EqualValues(ci[0], testCrawlInfoA)
}

func (s *PGStoreTestSuite) TestGetCrawlInfo_ErrQuery() {
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlGetCrawlInfo)).WillReturnError(testError)
	ci, err := s.store.GetCrawlInfo(1)
	s.Len(ci, 0)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestCreateCrawlInfo_OK() {
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	s.mdb.ExpectBegin()
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlCreateCrawlInfo)).WithArgs(testCrawlInfoA.SiteDefID, testCrawlInfoA.URL).WillReturnRows(rows)
	s.mdb.ExpectCommit()
	id, err := s.store.CreateCrawlInfo(testCrawlInfoA.SiteDefID, testCrawlInfoA.URL)
	s.NoError(err)
	s.EqualValues(1, id)
}

func (s *PGStoreTestSuite) TestCreateCrawlInfo_ErrBegin() {
	s.mdb.ExpectBegin().WillReturnError(testError)
	id, err := s.store.CreateCrawlInfo(testCrawlInfoA.SiteDefID, testCrawlInfoA.URL)
	s.EqualError(err, "some error")
	s.Zero(id)
}

func (s *PGStoreTestSuite) TestCreateCrawlInfo_ErrExec() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlCreateCrawlInfo)).WithArgs(testCrawlInfoA.SiteDefID, testCrawlInfoA.URL).WillReturnError(testError)
	id, err := s.store.CreateCrawlInfo(testCrawlInfoA.SiteDefID, testCrawlInfoA.URL)
	s.EqualError(err, "some error")
	s.Zero(id)
}

func (s *PGStoreTestSuite) TestCreateCrawlInfo_ErrCommit() {
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	s.mdb.ExpectBegin()
	s.mdb.ExpectQuery(regexp.QuoteMeta(sqlCreateCrawlInfo)).WithArgs(testCrawlInfoA.SiteDefID, testCrawlInfoA.URL).WillReturnRows(rows)
	s.mdb.ExpectCommit().WillReturnError(testError)
	id, err := s.store.CreateCrawlInfo(testCrawlInfoA.SiteDefID, testCrawlInfoA.URL)
	s.EqualError(err, "some error")
	s.EqualValues(0, id)
}

func (s *PGStoreTestSuite) TestStartCrawlInfo_OK() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlStartCrawlInfo)).WithArgs(testCrawlInfoA.ID).WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit()
	err := s.store.StartCrawlInfo(testCrawlInfoA.ID)
	s.NoError(err)
}

func (s *PGStoreTestSuite) TestStartCrawlInfo_ErrBegin() {
	s.mdb.ExpectBegin().WillReturnError(testError)
	err := s.store.StartCrawlInfo(testCrawlInfoA.ID)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestStartCrawlInfo_ErrExec() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlStartCrawlInfo)).WithArgs(testCrawlInfoA.ID).WillReturnError(testError)
	err := s.store.StartCrawlInfo(testCrawlInfoA.ID)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestStartCrawlInfo_ErrCommit() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlStartCrawlInfo)).WithArgs(testCrawlInfoA.ID).WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit().WillReturnError(testError)
	err := s.store.StartCrawlInfo(testCrawlInfoA.ID)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestEndCrawlInfo_OK() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlEndCrawlInfo)).WithArgs(testCrawlInfoA.ID, testError.Error(), 1).WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit()
	err := s.store.EndCrawlInfo(testCrawlInfoA.ID, testError, 1)
	s.NoError(err)
}

func (s *PGStoreTestSuite) TestEndCrawlInfo_ErrBegin() {
	s.mdb.ExpectBegin().WillReturnError(testError)
	err := s.store.EndCrawlInfo(testCrawlInfoA.ID, testError, 1)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestEndCrawlInfo_ErrExec() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlEndCrawlInfo)).WithArgs(testCrawlInfoA.ID, testError.Error(), 1).WillReturnError(testError)
	err := s.store.EndCrawlInfo(testCrawlInfoA.ID, testError, 1)
	s.EqualError(err, "some error")
}

func (s *PGStoreTestSuite) TestEndCrawlInfo_ErrCommit() {
	s.mdb.ExpectBegin()
	s.mdb.ExpectExec(regexp.QuoteMeta(sqlEndCrawlInfo)).WithArgs(testCrawlInfoA.ID, testError.Error(), 1).WillReturnResult(driver.ResultNoRows)
	s.mdb.ExpectCommit().WillReturnError(testError)
	err := s.store.EndCrawlInfo(testCrawlInfoA.ID, testError, 1)
	s.EqualError(err, "some error")
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(PGStoreTestSuite))
}
