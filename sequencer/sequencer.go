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
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/mhelmich/calvin/pb"
	"github.com/mhelmich/calvin/ulid"
	"github.com/mhelmich/calvin/util"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
	"google.golang.org/grpc"
)

const (
	sequencerBatchFrequencyMs = 100
)

func NewSequencer(raftID uint64, txnBatchChan chan<- *pb.TransactionBatch, peers []raft.Peer, storeDir string, connCache util.ConnectionCache, cip util.ClusterInfoProvider, srvr *grpc.Server, snapshotHandler SnapshotHandler, logger *log.Entry) *Sequencer {
	proposeChan := make(chan []byte)
	proposeConfChangeChan := make(chan raftpb.ConfChange)
	writerChan := make(chan *pb.Transaction)
	s := &Sequencer{
		proposeChan:           proposeChan,
		proposeConfChangeChan: proposeConfChangeChan,
		writerChan:            writerChan,
		cip:                   cip,
		rb:                    newRaftBackend(raftID, proposeChan, proposeConfChangeChan, txnBatchChan, peers, storeDir, connCache, snapshotHandler, logger),
		logger:                logger,
	}

	pb.RegisterRaftTransportServer(srvr, s.rb)
	go s.serveTxnBatches()
	return s
}

type Sequencer struct {
	rb                    *raftBackend
	proposeChan           chan<- []byte
	proposeConfChangeChan chan<- raftpb.ConfChange
	writerChan            chan *pb.Transaction
	cip                   util.ClusterInfoProvider
	logger                *log.Entry
}

// transactions and distributed snapshot reads go here
func (s *Sequencer) serveTxnBatches() {
	batch := &pb.TransactionBatch{}
	batchTicker := time.NewTicker(sequencerBatchFrequencyMs * time.Millisecond)
	defer batchTicker.Stop()

	for {
		select {
		case txn, ok := <-s.writerChan:
			if !ok {
				s.logger.Warningf("Stop serving txn batches")
				close(s.proposeChan)
				close(s.proposeConfChangeChan)
				return
			}

			if txn == nil {
				s.logger.Warningf("Sent nil transaction")
				return
			}

			s.findParticipants(txn)
			batch.Transactions = append(batch.Transactions, txn)
			if log.GetLevel() == log.DebugLevel {
				id, _ := ulid.ParseIdFromProto(txn.Id)
				s.logger.Debugf("Appended txn [%s]", id.String())
				a := make([]string, 0)
				for idx := range txn.WriterNodes {
					a = append(a, strconv.FormatUint(txn.WriterNodes[idx], 10))
				}
				s.logger.Debugf("[%s] WriterNodes: %s", id.String(), strings.Join(a, ", "))

				a = make([]string, 0)
				for idx := range txn.ReaderNodes {
					a = append(a, strconv.FormatUint(txn.ReaderNodes[idx], 10))
				}
				s.logger.Debugf("[%s] ReaderNodes: %s", id.String(), strings.Join(a, ", "))
			}

		case <-batchTicker.C:
			if len(batch.Transactions) > 0 {
				bites, err := batch.Marshal()
				if err != nil {
					s.logger.Panicf("%s", err)
				}

				s.proposeChan <- bites
				batch = &pb.TransactionBatch{}
			}

		}
	}
}

func (s *Sequencer) findParticipants(txn *pb.Transaction) {
	readerMap := make(map[uint64]bool)
	writerMap := make(map[uint64]bool)

	for idx := range txn.ReadWriteSet {
		ownerID := s.cip.FindOwnerForKey(txn.ReadWriteSet[idx])
		readerMap[ownerID] = true
		writerMap[ownerID] = true
	}

	for idx := range txn.ReadSet {
		ownerID := s.cip.FindOwnerForKey(txn.ReadSet[idx])
		readerMap[ownerID] = true
	}

	writers := make([]uint64, len(writerMap))
	readers := make([]uint64, len(readerMap))

	i := 0
	for key := range writerMap {
		writers[i] = key
		i++
	}

	i = 0
	for key := range readerMap {
		readers[i] = key
		i++
	}

	txn.WriterNodes = writers
	txn.ReaderNodes = readers
}

func (s *Sequencer) SubmitTransaction(txn *pb.Transaction) {
	s.writerChan <- txn
}

func (s *Sequencer) Stop() {
	close(s.writerChan)
}

func (s *Sequencer) LogToJSON(out io.Writer, n int) error {
	return s.rb.logToJSON(out, n)
}
