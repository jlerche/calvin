/*
 * Copyright 2018 Marco Helmich
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

package util

import (
	"fmt"
	"hash/fnv"
	"log"
	"os"

	"github.com/naoina/toml"
)

type clusterInfo struct {
	NumberPrimaries  int
	NumberPartitions int
	Nodes            map[uint64]Node
}

type Node struct {
	NodeID     uint64
	Hostname   string
	Port       int
	Partitions []int
}

var staticClusterInfo clusterInfo

func readClusterInfo(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("%s\n", err.Error())
	}
	defer f.Close()

	if err := toml.NewDecoder(f).Decode(&staticClusterInfo); err != nil {
		log.Fatalf("%s\n", err.Error())
	}
}

func IsLocal(key []byte, nodeID uint64) bool {
	hasher := fnv.New64()
	hasher.Write(key)
	partition := int(hasher.Sum64() % uint64(staticClusterInfo.NumberPartitions))
	node := staticClusterInfo.Nodes[nodeID]
	for idx := range node.Partitions {
		if node.Partitions[idx] == partition {
			return true
		}
	}
	return false
}

func getAddressForNodeID(nodeID uint64) string {
	info := staticClusterInfo.Nodes[nodeID]
	return fmt.Sprintf("%s:%d", info.Hostname, info.Port)
}
