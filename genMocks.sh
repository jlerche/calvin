#! /bin/bash

# Copyright 2019 Marco Helmich
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# go get -u github.com/vektra/mockery

rm mocks/*.go

mockery -dir util -name ClusterInfoProvider -output "$(dirname "$0")/mocks"
mockery -dir util -name ConnectionCache -output "$(dirname "$0")/mocks"
mockery -dir util -name PartitionedDataStore -output "$(dirname "$0")/mocks"
mockery -dir util -name DataStoreTxnProvider -output "$(dirname "$0")/mocks"
mockery -dir util -name DataStoreTxn -output "$(dirname "$0")/mocks"
mockery -dir pb -name RemoteReadClient -output "$(dirname "$0")/mocks"
mockery -dir sequencer -name SnapshotHandler -output "$(dirname "$0")/mocks"
mockery -dir sequencer -name PartialSnapshotHandler -output "$(dirname "$0")/mocks"

go mod tidy
