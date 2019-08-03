// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import pb "github.com/mhelmich/calvin/pb"

// ConnectionCache is an autogenerated mock type for the ConnectionCache type
type ConnectionCache struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *ConnectionCache) Close() {
	_m.Called()
}

// GetRaftTransportClient provides a mock function with given fields: nodeID
func (_m *ConnectionCache) GetRaftTransportClient(nodeID uint64) (pb.RaftTransportClient, error) {
	ret := _m.Called(nodeID)

	var r0 pb.RaftTransportClient
	if rf, ok := ret.Get(0).(func(uint64) pb.RaftTransportClient); ok {
		r0 = rf(nodeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pb.RaftTransportClient)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(nodeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRemoteReadClient provides a mock function with given fields: nodeID
func (_m *ConnectionCache) GetRemoteReadClient(nodeID uint64) (pb.RemoteReadClient, error) {
	ret := _m.Called(nodeID)

	var r0 pb.RemoteReadClient
	if rf, ok := ret.Get(0).(func(uint64) pb.RemoteReadClient); ok {
		r0 = rf(nodeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pb.RemoteReadClient)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(nodeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
