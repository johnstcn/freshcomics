package dao

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/erikstmartin/go-testdb"
	"github.com/stretchr/testify/assert"

	"github.com/johnstcn/freshcomics/backendsvc/models"
	"time"
)

var siteUpdateTestOpener = func() (*sql.DB) {
	db, _ := sql.Open("testdb", "")
	testdb.StubExec(SiteUpdateSchema, &testdb.Result{})
	return db
}


func TestNewSiteUpdateStore_OK(t *testing.T) {
	defer testdb.Reset()
	store, err := NewSiteUpdateStore(siteUpdateTestOpener)
	assert.NotNil(t, store)
	assert.Nil(t, err)
}

func TestNewSiteUpdateStore_Err_Open(t *testing.T) {
	defer testdb.Reset()
	testErr := errors.New("test error")
	testdb.SetOpenFunc(func(dsn string) (driver.Conn, error) {
		return nil, testErr
	})
	store, err := NewSiteUpdateStore(siteUpdateTestOpener)
	assert.Nil(t, store)
	assert.EqualValues(t, testErr, err)
}

func TestNewSiteUpdateStore_Err_Exec(t *testing.T) {
	defer testdb.Reset()
	testErr := errors.New("test error")
	errOpener := func() (*sql.DB) {
		db, _ := sql.Open("testdb", "")
		testdb.StubExecError(SiteUpdateSchema, testErr)
		return db
	}
	store, err := NewSiteUpdateStore(errOpener)
	assert.Nil(t, store)
	assert.EqualValues(t, testErr, err)
}

func TestSiteUpdateStore_Create_OK(t *testing.T) {
	defer testdb.Reset()
	createCols := []string{"id"}
	createRows := `
	1
	`
	createResult := testdb.RowsFromCSVString(createCols, createRows)
	testdb.StubQuery(SiteUpdateCreateStmt, createResult)

	s, _ := NewSiteUpdateStore(siteUpdateTestOpener)
	id, err := s.Create(1, "test", "http://test.com", "Test", time.Now())
	assert.EqualValues(t, id, 1)
	assert.Nil(t, err)
}

func TestSiteUpdateStore_Create_Err_Begin(t *testing.T) {
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
	testdb.StubQuery(SiteUpdateCreateStmt, createResult)

	s, _ := NewSiteUpdateStore(siteUpdateTestOpener)
	id, err := s.Create(1, "test", "http://test.com", "Test", time.Now())
	assert.EqualValues(t, -1, id)
	assert.EqualValues(t, testErr, err)
}

func TestSiteUpdateStore_Create_Err_Scan(t *testing.T) {
	defer testdb.Reset()
	testdb.StubQuery(SiteUpdateCreateStmt, testdb.RowsFromCSVString([]string{}, ``))
	s, _ := NewSiteUpdateStore(siteUpdateTestOpener)
	id, err := s.Create(1, "test", "http://test.com", "Test", time.Now())
	assert.EqualValues(t, -1, id)
	assert.NotNil(t, err)
}

func TestSiteUpdateStore_Create_Err_Commit(t *testing.T) {
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
	testdb.StubQuery(SiteUpdateCreateStmt, createResult)

	s, _ := NewSiteUpdateStore(siteUpdateTestOpener)
	id, err := s.Create(1, "test", "http://test.com", "Test", time.Now())
	assert.False(t, committed)
	assert.EqualValues(t, -1, id)
	assert.NotNil(t, err)
}

func TestSiteUpdateStore_Get_OK(t *testing.T) {
	defer testdb.Reset()
	getCols := []string{"id", "site_def_id", "ref", "url", "title", "seen_at"}
	getRows := `
	1,1,test ref,http://example.com/test%20ref,test title,2006-01-02T15:04:05.999999999Z07:00
	`
	getResult := testdb.RowsFromCSVString(getCols, getRows)
	testdb.StubQuery(SiteUpdateGetStmt, getResult)

	s, err := NewSiteUpdateStore(siteUpdateTestOpener)
	assert.NotNil(t, s)
	assert.Nil(t, err)
	def, err := s.Get(1)
	assert.NotNil(t, def)
	assert.Nil(t, err)
}

func TestSiteUpdateStore_Get_Err_Scan(t *testing.T) {
	defer testdb.Reset()
	testErr := errors.New("test error")
	testdb.StubQueryError(SiteUpdateGetStmt, testErr)
	s, _ := NewSiteUpdateStore(siteUpdateTestOpener)
	def, err := s.Get(1)
	assert.Equal(t, (*models.SiteUpdate)(nil), def)
	assert.NotNil(t, err)
}

func TestSiteUpdateStore_GetAllBySiteDefID_OK(t *testing.T) {
	defer testdb.Reset()
	getCols := []string{"id", "site_def_id", "ref", "url", "title", "seen_at"}
	getRows := `
	1,1,test ref,http://example.com/test%20ref,test title,2006-01-02T15:04:05.999999999Z07:00
	`
	getResult := testdb.RowsFromCSVString(getCols, getRows)
	testdb.StubQuery(SiteUpdateGetAllBySiteDefIDStmt, getResult)

	s, err := NewSiteUpdateStore(siteUpdateTestOpener)
	assert.NotNil(t, s)
	assert.Nil(t, err)
	defs, err := s.GetAllBySiteDefID(1)
	assert.NotNil(t, defs)
	assert.Nil(t, err)
}

func TestSiteUpdateStore_GetAll_Err_Query(t *testing.T) {
	defer testdb.Reset()
	testErr := errors.New("test error")
	testdb.StubQueryError(SiteUpdateGetAllBySiteDefIDStmt, testErr)
	s, _ := NewSiteUpdateStore(siteUpdateTestOpener)
	defs, err := s.GetAllBySiteDefID(1)
	assert.Equal(t, (*[]models.SiteUpdate)(nil), defs)
	assert.NotNil(t, err)
}

func TestSiteUpdateStore_GetAll_Err_Scan(t *testing.T) {
	defer testdb.Reset()
	testdb.StubQuery(SiteUpdateGetAllBySiteDefIDStmt, testdb.RowsFromCSVString([]string{"fdsa"}, `\nasdf\n`))
	s, _ := NewSiteUpdateStore(siteUpdateTestOpener)
	defs, err := s.GetAllBySiteDefID(1)
	assert.EqualValues(t, (*[]models.SiteUpdate)(nil), defs)
	assert.NotNil(t, err)
}

func TestSiteUpdateStore_Update_OK(t *testing.T) {
	defer testdb.Reset()
	testdb.StubExec(SiteUpdateUpdateStmt, &testdb.Result{})
	s, _ := NewSiteUpdateStore(siteUpdateTestOpener)
	err := s.Update(&models.SiteUpdate{})
	assert.Nil(t, err)
}

func TestSiteUpdateStore_Update_Err_Begin(t *testing.T) {
	defer testdb.Reset()
	beginErr := errors.New("test begin error")
	testdb.SetBeginFunc(func() (tx driver.Tx, err error) {
		return nil, beginErr
	})
	s, _ := NewSiteUpdateStore(siteUpdateTestOpener)
	err := s.Update(&models.SiteUpdate{})
	assert.EqualValues(t, beginErr, err)
}

func TestSiteUpdateStore_Update_Err_Exec(t *testing.T) {
	defer testdb.Reset()
	execErr := errors.New("test exec error")
	testdb.StubExecError(SiteUpdateUpdateStmt, execErr)
	s, _ := NewSiteUpdateStore(siteUpdateTestOpener)
	err := s.Update(&models.SiteUpdate{})
	assert.Equal(t, execErr, err)
}

func TestSiteUpdateStore_Update_Err_Commit(t *testing.T) {
	defer testdb.Reset()
	var committed bool
	testdb.StubExec(SiteUpdateUpdateStmt, &testdb.Result{})
	commitErr := errors.New("test commit error")
	testdb.SetBeginFunc(func() (driver.Tx, error) {
		tx := &testdb.Tx{}
		tx.SetCommitFunc(func() error {
			return commitErr
		})
		return tx, nil
	})
	s, _ := NewSiteUpdateStore(siteUpdateTestOpener)
	err := s.Update(&models.SiteUpdate{})
	assert.Equal(t, commitErr, err)
	assert.False(t, committed)
}