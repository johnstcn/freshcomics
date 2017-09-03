package dao

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/erikstmartin/go-testdb"
	"github.com/stretchr/testify/assert"

	"github.com/johnstcn/freshcomics/backendsvc/models"
)

var siteDefTestOpener = func() (*sql.DB) {
	db, _ := sql.Open("testdb", "")
	testdb.StubExec(SiteDefSchema, &testdb.Result{})
	return db
}


func TestNewSiteDefStore_OK(t *testing.T) {
	defer testdb.Reset()
	store, err := NewSiteDefStore(siteDefTestOpener)
	assert.NotNil(t, store)
	assert.Nil(t, err)
}

func TestNewSiteDefStore_Err_Open(t *testing.T) {
	defer testdb.Reset()
	testErr := errors.New("test error")
	testdb.SetOpenFunc(func(dsn string) (driver.Conn, error) {
		return nil, testErr
	})
	store, err := NewSiteDefStore(siteDefTestOpener)
	assert.Nil(t, store)
	assert.EqualValues(t, testErr, err)
}

func TestNewSiteDefStore_Err_Exec(t *testing.T) {
	defer testdb.Reset()
	testErr := errors.New("test error")
	errOpener := func() (*sql.DB) {
		db, _ := sql.Open("testdb", "")
		testdb.StubExecError(SiteDefSchema, testErr)
		return db
	}
	store, err := NewSiteDefStore(errOpener)
	assert.Nil(t, store)
	assert.EqualValues(t, testErr, err)
}

func TestSiteDefStore_Create_OK(t *testing.T) {
	defer testdb.Reset()
	createCols := []string{"id"}
	createRows := `
	1
	`
	createResult := testdb.RowsFromCSVString(createCols, createRows)
	testdb.StubQuery(SiteDefCreateStmt, createResult)

	s, _ := NewSiteDefStore(siteDefTestOpener)
	id, err := s.Create()
	assert.EqualValues(t, id, 1)
	assert.Nil(t, err)
}

func TestSiteDefStore_Create_Err_Begin(t *testing.T) {
	defer testdb.Reset()
	testErr := errors.New("test error")

	testdb.SetBeginFunc(func() (tx driver.Tx, err error) {
		return nil, testErr
	})

	createCols := []string{"id"}
	createRows := `
	1
	`
	createResult := testdb.RowsFromCSVString(createCols, createRows)
	testdb.StubQuery(SiteDefCreateStmt, createResult)

	s, _ := NewSiteDefStore(siteDefTestOpener)
	id, err := s.Create()
	assert.EqualValues(t, -1, id)
	assert.EqualValues(t, testErr, err)
}

func TestSiteDefStore_Create_Err_Scan(t *testing.T) {
	defer testdb.Reset()
	testdb.StubQuery(SiteDefCreateStmt, testdb.RowsFromCSVString([]string{}, ``))
	s, _ := NewSiteDefStore(siteDefTestOpener)
	id, err := s.Create()
	assert.EqualValues(t, -1, id)
	assert.NotNil(t, err)
}

func TestSiteDefStore_Create_Err_Commit(t *testing.T) {
	defer testdb.Reset()
	var committed bool
	testErr := errors.New("test error")

	testdb.SetBeginFunc(func() (driver.Tx, error) {
		tx := &testdb.Tx{}
		tx.SetCommitFunc(func() error {
			return testErr
		})
		return tx, nil
	})

	createCols := []string{"id"}
	createRows := `
	1
	`
	createResult := testdb.RowsFromCSVString(createCols, createRows)
	testdb.StubQuery(SiteDefCreateStmt, createResult)

	s, _ := NewSiteDefStore(siteDefTestOpener)
	id, err := s.Create()
	assert.False(t, committed)
	assert.EqualValues(t, -1, id)
	assert.NotNil(t, err)
}

func TestSiteDefStore_Get_OK(t *testing.T) {
	defer testdb.Reset()
	getCols := []string{"id", "name", "active", "nsfw", "start_url", "last_checked_at", "url_template", "ref_xpath", "ref_regexp", "title_xpath", "title_regexp"}
	getRows := `
	1,test_name,false,false,http://example.com,2006-01-02T15:04:05.999999999Z07:00,http://example.com/%s,//a,(.*),//a,(.*)
	`
	getResult := testdb.RowsFromCSVString(getCols, getRows)
	testdb.StubQuery(SiteDefGetStmt, getResult)

	s, err := NewSiteDefStore(siteDefTestOpener)
	assert.NotNil(t, s)
	assert.Nil(t, err)
	def, err := s.Get(1)
	assert.NotNil(t, def)
	assert.Nil(t, err)
}

func TestSiteDefStore_Get_Err_Scan(t *testing.T) {
	defer testdb.Reset()
	testErr := errors.New("test error")
	testdb.StubQueryError(SiteDefGetStmt, testErr)
	s, _ := NewSiteDefStore(siteDefTestOpener)
	def, err := s.Get(1)
	assert.Equal(t, (*models.SiteDef)(nil), def)
	assert.NotNil(t, err)
}

func TestSiteDefStore_GetAll_OK(t *testing.T) {
	defer testdb.Reset()
	getCols := []string{"id", "name", "active", "nsfw", "start_url", "last_checked_at", "url_template", "ref_xpath", "ref_regexp", "title_xpath", "title_regexp"}
	getRows := `
	1,test_name,false,false,http://example.com,2006-01-02T15:04:05.999999999Z07:00,http://example.com/%s,//a,(.*),//a,(.*)
	`
	getResult := testdb.RowsFromCSVString(getCols, getRows)
	testdb.StubQuery(SiteDefGetAllStmt, getResult)

	s, err := NewSiteDefStore(siteDefTestOpener)
	assert.NotNil(t, s)
	assert.Nil(t, err)
	defs, err := s.GetAll()
	assert.NotNil(t, defs)
	assert.Nil(t, err)
}

func TestSiteDefStore_GetAll_Err_Query(t *testing.T) {
	defer testdb.Reset()
	testErr := errors.New("test error")
	testdb.StubQueryError(SiteDefGetAllStmt, testErr)
	s, _ := NewSiteDefStore(siteDefTestOpener)
	defs, err := s.GetAll()
	assert.Equal(t, (*[]models.SiteDef)(nil), defs)
	assert.NotNil(t, err)
}

func TestSiteDefStore_GetAll_Err_Scan(t *testing.T) {
	defer testdb.Reset()
	testdb.StubQuery(SiteDefGetAllStmt, testdb.RowsFromCSVString([]string{"fdsa"}, `\nasdf\n`))
	s, _ := NewSiteDefStore(siteDefTestOpener)
	defs, err := s.GetAll()
	assert.EqualValues(t, (*[]models.SiteDef)(nil), defs)
	assert.NotNil(t, err)
}

func TestSiteDefStore_Update_OK(t *testing.T) {
	defer testdb.Reset()
	testdb.StubExec(SiteDefUpdateStmt, &testdb.Result{})
	s, _ := NewSiteDefStore(siteDefTestOpener)
	err := s.Update(&models.SiteDef{})
	assert.Nil(t, err)
}

func TestSiteDefStore_Update_Err_Begin(t *testing.T) {
	defer testdb.Reset()
	beginErr := errors.New("test begin error")
	testdb.SetBeginFunc(func() (tx driver.Tx, err error) {
		return nil, beginErr
	})
	s, _ := NewSiteDefStore(siteDefTestOpener)
	err := s.Update(&models.SiteDef{})
	assert.EqualValues(t, beginErr, err)
}

func TestSiteDefStore_Update_Err_Exec(t *testing.T) {
	defer testdb.Reset()
	execErr := errors.New("test exec error")
	testdb.StubExecError(SiteDefUpdateStmt, execErr)
	s, _ := NewSiteDefStore(siteDefTestOpener)
	err := s.Update(&models.SiteDef{})
	assert.Equal(t, execErr, err)
}

func TestSiteDefStore_Update_Err_Commit(t *testing.T) {
	defer testdb.Reset()
	var committed bool
	testdb.StubExec(SiteDefUpdateStmt, &testdb.Result{})
	commitErr := errors.New("test commit error")
	testdb.SetBeginFunc(func() (driver.Tx, error) {
		tx := &testdb.Tx{}
		tx.SetCommitFunc(func() error {
			return commitErr
		})
		return tx, nil
	})
	s, _ := NewSiteDefStore(siteDefTestOpener)
	err := s.Update(&models.SiteDef{})
	assert.Equal(t, commitErr, err)
	assert.False(t, committed)
}