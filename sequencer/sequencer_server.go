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

// func newSequencerServer() *sequencerServer {
// 	return &sequencerServer{
// 		idsToRaftNodes: make(map[int64]*sequencer),
// 	}
// }
//
// type sequencerServer struct {
// 	idsToRaftNodes map[int64]*sequencer
// }

// func (ss *sequencerServer) Step(stream pb.RaftTransportService_StepServer) error {
// 	for { //ever...
// 		request, err := stream.Recv()
// 		if err == io.EOF {
// 			return nil
// 		} else if err != nil {
// 			return err
// 		}
//
// 		s := ss.idsToRaftNodes[request.RaftNodeId]
// 		err = s.raftNode.Step(*request.Message)
// 		if err != nil {
// 			return err
// 		}
//
// 		stream.Send(&pb.StepResp{})
// 	}
// }
