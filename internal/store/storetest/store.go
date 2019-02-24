// Code autogenerated by mockery v2.0.0
//
// Do not manually edit the content of this file.

// Package storetest contains autogenerated mocks.
package storetest

import "github.com/stretchr/testify/mock"
import "net"
import "github.com/johnstcn/freshcomics/internal/store"
import "time"

// Store is an autogenerated mock type for the Store type
type Store struct {
	mock.Mock
}

// CreateCrawlInfo provides a mock function with given fields: ci
func (mockerySelf *Store) CreateCrawlInfo(ci store.CrawlInfo) error {
	ret := mockerySelf.Called(ci)

	var r0 error
	if rf, ok := ret.Get(0).(func(store.CrawlInfo) error); ok {
		r0 = rf(ci)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateSiteDef provides a mock function with given fields: mockeryArg0
func (mockerySelf *Store) CreateSiteDef(mockeryArg0 store.SiteDef) (int, error) {
	ret := mockerySelf.Called(mockeryArg0)

	var r0 int
	if rf, ok := ret.Get(0).(func(store.SiteDef) int); ok {
		r0 = rf(mockeryArg0)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(store.SiteDef) error); ok {
		r1 = rf(mockeryArg0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateSiteUpdate provides a mock function with given fields: su
func (mockerySelf *Store) CreateSiteUpdate(su store.SiteUpdate) error {
	ret := mockerySelf.Called(su)

	var r0 error
	if rf, ok := ret.Get(0).(func(store.SiteUpdate) error); ok {
		r0 = rf(su)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetSiteDefs provides a mock function with given fields: includeInactive
func (mockerySelf *Store) GetAllSiteDefs(includeInactive bool) ([]store.SiteDef, error) {
	ret := mockerySelf.Called(includeInactive)

	var r0 []store.SiteDef
	if rf, ok := ret.Get(0).(func(bool) []store.SiteDef); ok {
		r0 = rf(includeInactive)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]store.SiteDef)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(bool) error); ok {
		r1 = rf(includeInactive)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetComics provides a mock function with given fields:
func (mockerySelf *Store) GetComics() ([]store.Comic, error) {
	ret := mockerySelf.Called()

	var r0 []store.Comic
	if rf, ok := ret.Get(0).(func() []store.Comic); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]store.Comic)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCrawlInfo provides a mock function with given fields:
func (mockerySelf *Store) GetCrawlInfo() ([]store.CrawlInfo, error) {
	ret := mockerySelf.Called()

	var r0 []store.CrawlInfo
	if rf, ok := ret.Get(0).(func() []store.CrawlInfo); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]store.CrawlInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCrawlInfo provides a mock function with given fields: sdid
func (mockerySelf *Store) GetCrawlInfo(sdid int64) ([]store.CrawlInfo, error) {
	ret := mockerySelf.Called(sdid)

	var r0 []store.CrawlInfo
	if rf, ok := ret.Get(0).(func(int64) []store.CrawlInfo); ok {
		r0 = rf(sdid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]store.CrawlInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(sdid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Redirect provides a mock function with given fields: suID
func (mockerySelf *Store) GetRedirectURL(suID string) (string, error) {
	ret := mockerySelf.Called(suID)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(suID)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(suID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSiteDef provides a mock function with given fields: id
func (mockerySelf *Store) GetSiteDefByID(id int64) (store.SiteDef, error) {
	ret := mockerySelf.Called(id)

	var r0 store.SiteDef
	if rf, ok := ret.Get(0).(func(int64) store.SiteDef); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(store.SiteDef)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLastChecked provides a mock function with given fields:
func (mockerySelf *Store) GetSiteDefLastChecked() (store.SiteDef, error) {
	ret := mockerySelf.Called()

	var r0 store.SiteDef
	if rf, ok := ret.Get(0).(func() store.SiteDef); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(store.SiteDef)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSiteUpdate provides a mock function with given fields: sdid, ref
func (mockerySelf *Store) GetSiteUpdate(sdid int64, ref string) (store.SiteUpdate, error) {
	ret := mockerySelf.Called(sdid, ref)

	var r0 store.SiteUpdate
	if rf, ok := ret.Get(0).(func(int64, string) store.SiteUpdate); ok {
		r0 = rf(sdid, ref)
	} else {
		r0 = ret.Get(0).(store.SiteUpdate)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64, string) error); ok {
		r1 = rf(sdid, ref)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSiteUpdates provides a mock function with given fields: sdid
func (mockerySelf *Store) GetSiteUpdates(sdid int64) ([]store.SiteUpdate, error) {
	ret := mockerySelf.Called(sdid)

	var r0 []store.SiteUpdate
	if rf, ok := ret.Get(0).(func(int64) []store.SiteUpdate); ok {
		r0 = rf(sdid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]store.SiteUpdate)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(sdid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetStartURLForCrawl provides a mock function with given fields: sd
func (mockerySelf *Store) GetStartURLForCrawl(sd store.SiteDef) (string, error) {
	ret := mockerySelf.Called(sd)

	var r0 string
	if rf, ok := ret.Get(0).(func(store.SiteDef) string); ok {
		r0 = rf(sd)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(store.SiteDef) error); ok {
		r1 = rf(sd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateClickLog provides a mock function with given fields: updateID, addr
func (mockerySelf *Store) RecordClick(updateID int, addr net.IP) error {
	ret := mockerySelf.Called(updateID, addr)

	var r0 error
	if rf, ok := ret.Get(0).(func(int, net.IP) error); ok {
		r0 = rf(updateID, addr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateSiteDef provides a mock function with given fields: sd
func (mockerySelf *Store) SaveSiteDef(sd store.SiteDef) error {
	ret := mockerySelf.Called(sd)

	var r0 error
	if rf, ok := ret.Get(0).(func(store.SiteDef) error); ok {
		r0 = rf(sd)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetLastChecked provides a mock function with given fields: sd, when
func (mockerySelf *Store) SetSiteDefLastChecked(sd store.SiteDef, when time.Time) error {
	ret := mockerySelf.Called(sd, when)

	var r0 error
	if rf, ok := ret.Get(0).(func(store.SiteDef, time.Time) error); ok {
		r0 = rf(sd, when)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
