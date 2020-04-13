// Copyright 2019-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package roles

import (
	"context"

	raft "github.com/atomix/raft-storage/pkg/atomix/raft/protocol"
	"github.com/atomix/raft-storage/pkg/atomix/raft/state"
	"github.com/atomix/raft-storage/pkg/atomix/raft/store"
	"github.com/atomix/raft-storage/pkg/atomix/raft/util"
)

// GetRoles returns a mapping of role types to role factories
func GetRoles(state state.Manager, store store.Store) map[raft.RoleType]func(raft.Raft) raft.Role {
	return map[raft.RoleType]func(raft.Raft) raft.Role{
		raft.RoleFollower: func(raft raft.Raft) raft.Role {
			return newFollowerRole(raft, state, store)
		},
		raft.RoleCandidate: func(raft raft.Raft) raft.Role {
			return newCandidateRole(raft, state, store)
		},
		raft.RoleLeader: func(raft raft.Raft) raft.Role {
			return newLeaderRole(raft, state, store)
		},
	}
}

func newRaftRole(raft raft.Raft, state state.Manager, store store.Store, log util.Logger) *raftRole {
	return &raftRole{
		raft:   raft,
		state:  state,
		store:  store,
		log:    log,
		active: true,
	}
}

// raftRole is the base role for all Raft Role implementations
type raftRole struct {
	raft   raft.Raft
	state  state.Manager
	store  store.Store
	log    util.Logger
	active bool
}

// Start starts the role
func (r *raftRole) Start() error {
	return nil
}

// Stop stops the role
func (r *raftRole) Stop() error {
	r.active = false
	return nil
}

// Join handles a join request
func (r *raftRole) Join(ctx context.Context, request *raft.JoinRequest) (*raft.JoinResponse, error) {
	r.log.Request("JoinRequest", request)
	response := &raft.JoinResponse{
		Status: raft.ResponseStatus_ERROR,
		Error:  raft.ResponseError_ILLEGAL_MEMBER_STATE,
	}
	_ = r.log.Response("JoinResponse", response, nil)
	return response, nil
}

// Leave handles a leave request
func (r *raftRole) Leave(ctx context.Context, request *raft.LeaveRequest) (*raft.LeaveResponse, error) {
	r.log.Request("LeaveRequest", request)
	response := &raft.LeaveResponse{
		Status: raft.ResponseStatus_ERROR,
		Error:  raft.ResponseError_ILLEGAL_MEMBER_STATE,
	}
	_ = r.log.Response("LeaveResponse", response, nil)
	return response, nil
}

// Configure handles a configure request
func (r *raftRole) Configure(ctx context.Context, request *raft.ConfigureRequest) (*raft.ConfigureResponse, error) {
	r.log.Request("ConfigureRequest", request)
	response := &raft.ConfigureResponse{
		Status: raft.ResponseStatus_ERROR,
		Error:  raft.ResponseError_ILLEGAL_MEMBER_STATE,
	}
	_ = r.log.Response("", response, nil)
	return response, nil
}

// Reconfigure handles a reconfigure request
func (r *raftRole) Reconfigure(ctx context.Context, request *raft.ReconfigureRequest) (*raft.ReconfigureResponse, error) {
	r.log.Request("ReconfigureRequest", request)
	response := &raft.ReconfigureResponse{
		Status: raft.ResponseStatus_ERROR,
		Error:  raft.ResponseError_ILLEGAL_MEMBER_STATE,
	}
	_ = r.log.Response("ReconfigureResponse", response, nil)
	return response, nil
}

// Poll handles a poll request
func (r *raftRole) Poll(ctx context.Context, request *raft.PollRequest) (*raft.PollResponse, error) {
	r.log.Request("PollRequest", request)
	response := &raft.PollResponse{
		Status: raft.ResponseStatus_ERROR,
		Error:  raft.ResponseError_ILLEGAL_MEMBER_STATE,
	}
	_ = r.log.Response("PollResponse", response, nil)
	return response, nil
}

// Vote handles a vote request
func (r *raftRole) Vote(ctx context.Context, request *raft.VoteRequest) (*raft.VoteResponse, error) {
	r.log.Request("VoteRequest", request)
	response := &raft.VoteResponse{
		Status: raft.ResponseStatus_ERROR,
		Error:  raft.ResponseError_ILLEGAL_MEMBER_STATE,
	}
	_ = r.log.Response("VoteResponse", response, nil)
	return response, nil
}

// Transfer handles a transfer request
func (r *raftRole) Transfer(ctx context.Context, request *raft.TransferRequest) (*raft.TransferResponse, error) {
	r.log.Request("TransferRequest", request)
	response := &raft.TransferResponse{
		Status: raft.ResponseStatus_ERROR,
		Error:  raft.ResponseError_ILLEGAL_MEMBER_STATE,
	}
	_ = r.log.Response("TransferResponse", response, nil)
	return response, nil
}

// Append handles a append request
func (r *raftRole) Append(ctx context.Context, request *raft.AppendRequest) (*raft.AppendResponse, error) {
	r.log.Request("AppendRequest", request)
	response := &raft.AppendResponse{
		Status: raft.ResponseStatus_ERROR,
		Error:  raft.ResponseError_ILLEGAL_MEMBER_STATE,
	}
	_ = r.log.Response("AppendResponse", response, nil)
	return response, nil
}

// Install handles an install request
func (r *raftRole) Install(ch <-chan *raft.InstallStreamRequest) (*raft.InstallResponse, error) {
	response := &raft.InstallResponse{
		Status: raft.ResponseStatus_ERROR,
		Error:  raft.ResponseError_ILLEGAL_MEMBER_STATE,
	}
	_ = r.log.Response("InstallResponse", response, nil)
	return response, nil
}

// Command handles a command request
func (r *raftRole) Command(request *raft.CommandRequest, ch chan<- *raft.CommandStreamResponse) error {
	defer close(ch)
	r.log.Request("CommandRequest", request)
	response := &raft.CommandResponse{
		Status: raft.ResponseStatus_ERROR,
		Error:  raft.ResponseError_ILLEGAL_MEMBER_STATE,
	}
	_ = r.log.Response("CommandResponse", response, nil)
	ch <- raft.NewCommandStreamResponse(response, nil)
	return nil
}

// Query handles a query request
func (r *raftRole) Query(request *raft.QueryRequest, ch chan<- *raft.QueryStreamResponse) error {
	defer close(ch)
	r.log.Request("QueryRequest", request)
	response := &raft.QueryResponse{
		Status: raft.ResponseStatus_ERROR,
		Error:  raft.ResponseError_ILLEGAL_MEMBER_STATE,
	}
	_ = r.log.Response("QueryResponse", response, nil)
	ch <- raft.NewQueryStreamResponse(response, nil)
	return nil
}
