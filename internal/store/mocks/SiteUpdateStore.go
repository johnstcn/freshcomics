// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	store "github.com/johnstcn/freshcomics/internal/store"
	mock "github.com/stretchr/testify/mock"
)

// SiteUpdateStore is an autogenerated mock type for the SiteUpdateStore type
type SiteUpdateStore struct {
	mock.Mock
}

// CreateSiteUpdate provides a mock function with given fields: su
func (_m *SiteUpdateStore) CreateSiteUpdate(su store.SiteUpdate) (store.SiteUpdateID, error) {
	ret := _m.Called(su)

	var r0 store.SiteUpdateID
	if rf, ok := ret.Get(0).(func(store.SiteUpdate) store.SiteUpdateID); ok {
		r0 = rf(su)
	} else {
		r0 = ret.Get(0).(store.SiteUpdateID)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(store.SiteUpdate) error); ok {
		r1 = rf(su)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSiteUpdate provides a mock function with given fields: id, ref
func (_m *SiteUpdateStore) GetSiteUpdate(id store.SiteDefID, ref string) (store.SiteUpdate, bool, error) {
	ret := _m.Called(id, ref)

	var r0 store.SiteUpdate
	if rf, ok := ret.Get(0).(func(store.SiteDefID, string) store.SiteUpdate); ok {
		r0 = rf(id, ref)
	} else {
		r0 = ret.Get(0).(store.SiteUpdate)
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(store.SiteDefID, string) bool); ok {
		r1 = rf(id, ref)
	} else {
		r1 = ret.Get(1).(bool)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(store.SiteDefID, string) error); ok {
		r2 = rf(id, ref)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetSiteUpdates provides a mock function with given fields: id
func (_m *SiteUpdateStore) GetSiteUpdates(id store.SiteDefID) ([]store.SiteUpdate, error) {
	ret := _m.Called(id)

	var r0 []store.SiteUpdate
	if rf, ok := ret.Get(0).(func(store.SiteDefID) []store.SiteUpdate); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]store.SiteUpdate)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(store.SiteDefID) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}