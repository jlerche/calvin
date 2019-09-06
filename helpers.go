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

package calvin

import (
	"github.com/mhelmich/calvin/pb"
	"github.com/mhelmich/calvin/sequencer"
	"github.com/mhelmich/calvin/ulid"
	log "github.com/sirupsen/logrus"
)

func NewTransaction() *pb.Transaction {
	id, err := ulid.NewId()
	if err != nil {
		log.Panicf("Can't generate new ulid")
	}

	return &pb.Transaction{
		Id: id.ToProto(),
	}
}

func DefaultOptions(configPath string, clusterInfoPath string) *Options {
	return &Options{
		configPath:      configPath,
		clusterInfoPath: clusterInfoPath,
	}
}

type Options struct {
	configPath      string
	clusterInfoPath string
	snapshotHandler sequencer.SnapshotHandler
}

func (o *Options) WithSnapshotHandler(snapshotHandler sequencer.SnapshotHandler) *Options {
	o.snapshotHandler = snapshotHandler
	return o
}
