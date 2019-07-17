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

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/mhelmich/calvin/interfaces"
	"github.com/mhelmich/calvin/pb"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
)

const (
	sequencerBatchFrequencyMs = 10
)

type SequencerConfig struct {
	newRaftID         uint64
	localRaftStore    localRaftStore
	raftMessageClient interfaces.RaftMessageClient
	logger            *log.Entry
}

type sequencer struct {
	raftId uint64
	// Transaction channel for writers.
	writerChanIn <-chan *pb.Transaction
	// Transaction channel for readers.
	readerChanIn <-chan *pb.Transaction
	// TrancsactionBatch channel for schedulers
	schedulerChanOut chan<- *pb.TransactionBatch
	// The raft node.
	raftNode *raft.RawNode
	// The raft data store.
	localRaftStore localRaftStore
	// The client to all other rafts.
	raftMessageClient interfaces.RaftMessageClient
	// The last index that has been applied. It helps us figuring out which entries to publish.
	appliedIndex uint64
	// // The index of the latest snapshot. Used to compute when to cut the next snapshot.
	// snapshotIndex uint64
	// // The number of log entries after which we cut a snapshot.
	// snapshotFrequency uint64
	// // Defines the number of snapshots that Calvin keeps before deleting old ones.
	// numberOfSnapshotsToKeep int
	// This object describes the topology of the raft group this backend is part of
	confState raftpb.ConfState
	// Keeps track of the latest config change id. The next id is this id + 1.
	// HACK - there's obviously a race condition around this considering that
	// whatever id this raft node has, the leader in the group might have a higher one
	latestConfChangeId uint64
	logger             *log.Entry
}

func NewSequencer(config SequencerConfig) (chan<- *pb.Transaction, chan<- *pb.Transaction, <-chan *pb.TransactionBatch, error) {
	writerChan := make(chan *pb.Transaction)
	readerChan := make(chan *pb.Transaction)
	schedulerChan := make(chan *pb.TransactionBatch)

	if config.logger == nil {
		config.logger = log.WithFields(log.Fields{
			"component": "sequencer",
			"raftIdHex": hex.EncodeToString(uint64ToBytes(config.newRaftID)),
			"raftId":    uint64ToString(config.newRaftID),
		})
	}

	if config.newRaftID == 0 || config.localRaftStore == nil || config.raftMessageClient == nil {
		return nil, nil, nil, fmt.Errorf("One mandatory config item is nil")
	}

	c := &raft.Config{
		ID:              config.newRaftID,
		ElectionTick:    7,
		HeartbeatTick:   5,
		Storage:         config.localRaftStore,
		MaxSizePerMsg:   1024 * 1024 * 1024, // 1 GB (!!!)
		MaxInflightMsgs: 256,
		Logger:          config.logger,
	}

	raftPeers := make([]raft.Peer, 1)
	raftPeers[0] = raft.Peer{
		ID:      config.newRaftID,
		Context: []byte("narf"),
	}

	n, err := raft.NewRawNode(c, raftPeers)
	if err != nil {
		return nil, nil, nil, err
	}

	s := &sequencer{
		raftId:            config.newRaftID,
		writerChanIn:      writerChan,
		readerChanIn:      readerChan,
		schedulerChanOut:  schedulerChan,
		raftNode:          n,
		localRaftStore:    config.localRaftStore,
		raftMessageClient: config.raftMessageClient,
		logger:            config.logger,
	}

	go s.runReader()
	go s.runWriter()

	// TODO: wait until node joined raft group and doesn't drop raft messages anymore
	return writerChan, readerChan, schedulerChan, nil
}

////////////////////////////////////////////////
////////////////////////////////////////////////
/////////////// RAFT CODE

func (s *sequencer) processReady(rd raft.Ready) {
	s.logger.Debugf("ID: %d %x Hardstate: %v Entries: %v Snapshot: %v Messages: %v Committed: %v", s.raftId, s.raftId, rd.HardState, rd.Entries, rd.Snapshot, rd.Messages, rd.CommittedEntries)
	s.localRaftStore.saveEntriesAndState(rd.Entries, rd.HardState)

	if !raft.IsEmptySnap(rd.Snapshot) {
		if err := s.localRaftStore.saveSnap(rd.Snapshot); err != nil {
			s.logger.Errorf("Couldn't save snapshot: %s", err.Error())
			return
		}
	}

	sendingErrors := s.raftMessageClient.SendMessages(rd.Messages)
	if sendingErrors != nil {
		for _, failedMsg := range sendingErrors.FailedMessages {
			// TODO - think this through
			// rb.logger.Errorf("Reporting raft [%d %x] unreachable", failedMsg.To, failedMsg.To)
			// rb.raftNode.ReportUnreachable(failedMsg.To)
			if isMsgSnap(failedMsg) {
				s.logger.Errorf("Reporting snapshot failure for raft [%d %x]", failedMsg.To, failedMsg.To)
				s.raftNode.ReportSnapshot(failedMsg.To, raft.SnapshotFailure)
			}
		}

		for _, snapMsg := range sendingErrors.SucceededSnapshotMessages {
			s.raftNode.ReportSnapshot(snapMsg.To, raft.SnapshotFinish)
		}
	}

	s.publishEntries(s.entriesToApply(rd.CommittedEntries))
	s.maybeTriggerSnapshot()
	s.raftNode.Advance(rd)
}

func (s *sequencer) publishEntries(ents []raftpb.Entry) {
	for idx := range ents {
		switch ents[idx].Type {
		case raftpb.EntryNormal:
			s.publishTransactionBatch(ents[idx])

		case raftpb.EntryConfChange:
			s.publishConfigChange(ents[idx])
		}
		s.appliedIndex = ents[idx].Index
	}
}

func (s *sequencer) publishConfigChange(entry raftpb.Entry) {
	var cc raftpb.ConfChange
	cc.Unmarshal(entry.Data)
	s.logger.Infof("Publishing config change: [%s]", cc.String())
	s.confState = *s.raftNode.ApplyConfChange(cc)
	s.localRaftStore.saveConfigState(s.confState)
	s.latestConfChangeId = cc.ID
}

func (s *sequencer) entriesToApply(ents []raftpb.Entry) []raftpb.Entry {
	if len(ents) == 0 {
		return make([]raftpb.Entry, 0)
	}

	firstIdx := ents[0].Index
	if firstIdx > s.appliedIndex+1 {
		// if I'm getting invalid data, I'm shutting down
		s.logger.Panicf("First index of committed entry [%d] should <= progress.appliedIndex[%d] !", firstIdx, s.appliedIndex)
		return make([]raftpb.Entry, 0)
	}

	return ents
}

func (s *sequencer) maybeTriggerSnapshot() {
	// triggering a snapshot means consistently capturing the
	// log and the data file and bundelling all of that into a snapshot
}

////////////////////////////////////////////////
////////////////////////////////////////////////
/////////////// CALVIN CODE

// low-isolation reads and single partition snapshot reads go here
func (s *sequencer) runReader() {
	for {
		select {
		case txn := <-s.readerChanIn:
			if txn == nil {
				s.logger.Warningf("Ending reader loop")
				return
			}
		}
	}
}

// transactions and distributed snapshot reads go here
func (s *sequencer) runWriter() {
	batch := &pb.TransactionBatch{}
	raftTicker := time.NewTicker(sequencerBatchFrequencyMs * time.Millisecond)
	batchTicker := time.NewTicker(sequencerBatchFrequencyMs * time.Millisecond)
	defer raftTicker.Stop()
	defer batchTicker.Stop()

	for {
		select {
		case txn := <-s.writerChanIn:
			if txn == nil {
				s.logger.Warningf("Ending writer loop")
				s.shutdown()
				return
			}

			batch.Transactions = append(batch.Transactions, txn)

		case <-raftTicker.C:
			s.raftNode.Tick()

		case <-batchTicker.C:
			if len(batch.Transactions) > 0 {
				bites, err := batch.Marshal()
				if err != nil {
					s.logger.Errorf("%s", err)
				}

				err = s.raftNode.Propose(bites)
				if err != nil {
					s.logger.Errorf("%s", err)
				}
			}

			batch = &pb.TransactionBatch{}

		default:
			if s.raftNode.HasReady() {
				rd := s.raftNode.Ready()
				s.processReady(rd)
			} else {
				time.Sleep((sequencerBatchFrequencyMs / 10) * time.Millisecond)
			}
		}
	}
}

func (s *sequencer) publishTransactionBatch(entry raftpb.Entry) {
	if len(entry.Data) <= 0 {
		return
	}

	batch := &pb.TransactionBatch{}
	err := batch.Unmarshal(entry.Data)
	if err != nil {
		s.logger.Panicf(err.Error())
	}

	batch.Term = entry.Term
	batch.Index = entry.Index
	batch.NodeId = s.raftId
	s.schedulerChanOut <- batch
}

func (s *sequencer) shutdown() {
	close(s.schedulerChanOut)
}
