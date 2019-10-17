// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import raftpb "go.etcd.io/etcd/raft/raftpb"

// PartialSnapshotHandler is an autogenerated mock type for the PartialSnapshotHandler type
type PartialSnapshotHandler struct {
	mock.Mock
}

// Consume provides a mock function with given fields: partitionID, snapshotData
func (_m *PartialSnapshotHandler) Consume(partitionID int, snapshotData []byte) error {
	ret := _m.Called(partitionID, snapshotData)

	var r0 error
	if rf, ok := ret.Get(0).(func(int, []byte) error); ok {
		r0 = rf(partitionID, snapshotData)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Provide provides a mock function with given fields: partitionID, lastSnapshot, entriesAppliedSinceLastSnapshot
func (_m *PartialSnapshotHandler) Provide(partitionID int, lastSnapshot raftpb.Snapshot, entriesAppliedSinceLastSnapshot []raftpb.Entry) ([]byte, error) {
	ret := _m.Called(partitionID, lastSnapshot, entriesAppliedSinceLastSnapshot)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(int, raftpb.Snapshot, []raftpb.Entry) []byte); ok {
		r0 = rf(partitionID, lastSnapshot, entriesAppliedSinceLastSnapshot)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int, raftpb.Snapshot, []raftpb.Entry) error); ok {
		r1 = rf(partitionID, lastSnapshot, entriesAppliedSinceLastSnapshot)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}