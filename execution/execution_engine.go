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

package execution

import (
	"context"
	"sync"

	"github.com/mhelmich/calvin/pb"
	"github.com/mhelmich/calvin/ulid"
	"github.com/mhelmich/calvin/util"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type DataStore interface {
	Get(key []byte) []byte
	Set(key []byte, value []byte)
}

func NewEngine(scheduledTxnChan <-chan *pb.Transaction, store DataStore, srvr *grpc.Server, connCache util.ConnectionCache, cip util.ClusterInfoProvider, logger *log.Entry) *engine {
	readyToExecChan := make(chan *txnExecEnvironment, 11)

	rrs := newRemoteReadServer(readyToExecChan)
	pb.RegisterRemoteReadServer(srvr, rrs)

	stopChan := make(chan interface{})

	for i := 0; i < 2; i++ {
		w := worker{
			scheduledTxnChan: scheduledTxnChan,
			readyToExecChan:  readyToExecChan,
			stopChan:         stopChan,
			store:            store,
			connCache:        connCache,
			cip:              cip,
			txnsToExecute:    &sync.Map{},
			logger:           logger,
		}
		go w.runWorker()
	}

	return &engine{
		// scheduledTxnChan: scheduledTxnChan,
		// readyToExecChan:  readyToExecChan,
		stopChan: stopChan,
		// store:            store,
		// connCache:        connCache,
	}
}

type engine struct {
	// scheduledTxnChan <-chan *pb.Transaction
	// readyToExecChan  <-chan *txnExecEnvironment
	stopChan chan<- interface{}
	// store            DataStore
	// connCache        util.ConnectionCache
}

func (e *engine) Stop() {
	close(e.stopChan)
}

type worker struct {
	scheduledTxnChan <-chan *pb.Transaction
	readyToExecChan  <-chan *txnExecEnvironment
	doneTxn          chan<- *pb.Transaction
	stopChan         <-chan interface{}
	store            DataStore
	connCache        util.ConnectionCache
	cip              util.ClusterInfoProvider
	txnsToExecute    *sync.Map
	logger           *log.Entry
}

func (w *worker) runWorker() {
	for {
		select {
		// wait for txns to be scheduled
		case txn := <-w.scheduledTxnChan:
			w.processScheduledTxn(txn, w.store)

		// wait for remote reads to be collected
		case execEnv := <-w.readyToExecChan:
			w.runReadyTxn(execEnv)

		case <-w.stopChan:
			return
		}
	}
}

func (w *worker) processScheduledTxn(txn *pb.Transaction, store DataStore) {
	localKeys := make([][]byte, 0)
	localValues := make([][]byte, 0)
	// do local reads
	for idx := range txn.ReadSet {
		key := txn.ReadSet[idx]
		if w.cip.IsLocal(key) {
			value := store.Get(key)
			localKeys = append(localKeys, key)
			localValues = append(localValues, value)
		}
	}
	for idx := range txn.ReadWriteSet {
		key := txn.ReadWriteSet[idx]
		if w.cip.IsLocal(key) {
			value := store.Get(key)
			localKeys = append(localKeys, key)
			localValues = append(localValues, value)
		}
	}

	if w.cip.AmIWriter(txn.WriterNodes) {
		id, err := ulid.ParseIdFromProto(txn.Id)
		if err != nil {
			w.logger.Fatalf("%s\n", err.Error())
		}
		w.txnsToExecute.Store(id.String(), txn)
	}

	// broadcast remote reads to all write peers
	w.broadcastLocalReadsToWriterNodes(txn, localKeys, localValues)
}

func (w *worker) broadcastLocalReadsToWriterNodes(txn *pb.Transaction, keys [][]byte, values [][]byte) {
	for idx := range txn.WriterNodes {
		client, err := w.connCache.GetRemoteReadClient(txn.WriterNodes[idx])
		if err != nil {
			w.logger.Fatalf("%s\n", err.Error())
		}

		resp, err := client.RemoteRead(context.Background(), &pb.RemoteReadRequest{
			TxnId:         txn.Id,
			TotalNumLocks: uint32(len(txn.ReadWriteSet) + len(txn.ReadSet)),
			Keys:          keys,
			Values:        values,
		})
		if err != nil {
			w.logger.Fatalf("%s\n", err.Error())
		} else if resp.Error != "" {
			w.logger.Fatalf("%s\n", resp.Error)
		}
	}
}

func (w *worker) runReadyTxn(execEnv *txnExecEnvironment) {
	t, ok := w.txnsToExecute.Load(execEnv.txnId)
	if !ok {
		w.logger.Fatalf("Can't find txn [%s]\n", execEnv.txnId.String())
	}
	txn := t.(*pb.Transaction)

	w.logger.Infof("ran txn: %s\n", txn.Id.String())

	w.doneTxn <- txn
}
