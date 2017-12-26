package store

import (
	"net"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"time"
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
	args := m.Called()
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

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
