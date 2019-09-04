// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import io "io"
import mock "github.com/stretchr/testify/mock"
import util "github.com/mhelmich/calvin/util"

// DataStoreTxnProvider is an autogenerated mock type for the DataStoreTxnProvider type
type DataStoreTxnProvider struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *DataStoreTxnProvider) Close() {
	_m.Called()
}

// Snapshot provides a mock function with given fields: w
func (_m *DataStoreTxnProvider) Snapshot(w io.Writer) error {
	ret := _m.Called(w)

	var r0 error
	if rf, ok := ret.Get(0).(func(io.Writer) error); ok {
		r0 = rf(w)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StartTxn provides a mock function with given fields: writable
func (_m *DataStoreTxnProvider) StartTxn(writable bool) (util.DataStoreTxn, error) {
	ret := _m.Called(writable)

	var r0 util.DataStoreTxn
	if rf, ok := ret.Get(0).(func(bool) util.DataStoreTxn); ok {
		r0 = rf(writable)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(util.DataStoreTxn)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(bool) error); ok {
		r1 = rf(writable)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
