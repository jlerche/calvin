/*
 * Copyright 2019 Marco Helmich
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sequencer

import "go.etcd.io/etcd/raft/raftpb"

type SnapshotHandler interface {
	Consume(snapshotData []byte) error
	Provide(lastSnapshot raftpb.Snapshot, entriesAppliedSinceLastSnapshot []raftpb.Entry) ([]byte, error)
}

type PartialSnapshotHandler interface {
	Consume(partitionID int, snapshotData []byte) error
	Provide(partitionID int, lastSnapshot raftpb.Snapshot, entriesAppliedSinceLastSnapshot []raftpb.Entry) ([]byte, error)
}
