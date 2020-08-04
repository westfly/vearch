// Copyright 2019 The Vearch Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package entity

import (
	"github.com/spf13/cast"
	"github.com/vearch/vearch/util"
)

type BuildVersion struct {
	BuildVersion string `json:"build_version"`
	BuildTime    string `json:"build_time"`
	CommitID     string `json:"commit_id"`
}

//server/id:[body] ttl 3m 3s
type Server struct {
	ID                NodeID        `json:"name,omitempty"` //unique name for raft
	RpcPort           uint16        `json:"rpc_port"`
	RaftHeartbeatPort uint16        `json:"raft_heartbeat_port"`
	RaftReplicatePort uint16        `json:"raft_replicate_port"`
	Ip                string        `json:"ip,omitempty"`
	PartitionIds      []PartitionID `json:"p_ids,omitempty"`
	Size              uint64        `json:"size,omitempty"`
	Private           bool          `json:"private"`
	Version           *BuildVersion `json:"version"`
}

func (s *Server) RpcAddr() string {
	return s.Ip + ":" + cast.ToString(s.RpcPort)
}

func (s *Server) Replica() *Replica {
	return &Replica{
		NodeID:        s.ID,
		HeartbeatAddr: util.BuildAddr(s.Ip, s.RaftHeartbeatPort),
		ReplicateAddr: util.BuildAddr(s.Ip, s.RaftReplicatePort),
		RpcAddr:       util.BuildAddr(s.Ip, s.RpcPort),
	}
}
