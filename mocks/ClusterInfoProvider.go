// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// ClusterInfoProvider is an autogenerated mock type for the ClusterInfoProvider type
type ClusterInfoProvider struct {
	mock.Mock
}

// AmIWriter provides a mock function with given fields: writerNodes
func (_m *ClusterInfoProvider) AmIWriter(writerNodes []uint64) bool {
	ret := _m.Called(writerNodes)

	var r0 bool
	if rf, ok := ret.Get(0).(func([]uint64) bool); ok {
		r0 = rf(writerNodes)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// GetAddressFor provides a mock function with given fields: nodeID
func (_m *ClusterInfoProvider) GetAddressFor(nodeID uint64) string {
	ret := _m.Called(nodeID)

	var r0 string
	if rf, ok := ret.Get(0).(func(uint64) string); ok {
		r0 = rf(nodeID)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// IsLocal provides a mock function with given fields: key
func (_m *ClusterInfoProvider) IsLocal(key []byte) bool {
	ret := _m.Called(key)

	var r0 bool
	if rf, ok := ret.Get(0).(func([]byte) bool); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
