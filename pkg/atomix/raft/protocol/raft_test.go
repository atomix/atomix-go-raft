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

package protocol

import (
	"context"
	"testing"
	"time"

	atomix "github.com/atomix/go-framework/pkg/atomix/cluster"
	"github.com/atomix/raft-storage/pkg/atomix/raft/config"
	"github.com/stretchr/testify/assert"
)

func TestRaftProtocol(t *testing.T) {
	foo := MemberID("foo")
	bar := MemberID("bar")

	cluster := atomix.Cluster{
		MemberID: "foo",
		Members: map[string]atomix.Member{
			"foo": {
				ID:   "foo",
				Host: "foo",
				Port: 5678,
			},
			"bar": {
				ID:   "bar",
				Host: "bar",
				Port: 5679,
			},
			"baz": {
				ID:   "baz",
				Port: 5680,
			},
		},
	}

	store := newMemoryMetadataStore()
	electionTimeout := 10 * time.Second
	roles := make(map[RoleType]func(Raft) Role)
	raft := newRaft(NewCluster(cluster), &config.ProtocolConfig{ElectionTimeout: &electionTimeout}, &unimplementedClient{}, roles, store)
	assert.Equal(t, StatusStopped, raft.Status())
	statusCh := make(chan Status, 1)
	raft.Watch(func(event Event) {
		if event.Type == EventTypeStatus {
			statusCh <- event.Status
		}
	})
	raft.WriteLock()
	raft.Init()
	assert.Equal(t, StatusRunning, raft.Status())
	assert.Equal(t, StatusRunning, <-statusCh)
	raft.WriteUnlock()

	// Verify the cluster configuration
	assert.Equal(t, electionTimeout, raft.Config().GetElectionTimeoutOrDefault())
	assert.Equal(t, MemberID("foo"), raft.Member())
	assert.Len(t, raft.Members(), 3)
	assert.Equal(t, MemberID("foo"), raft.GetMember(MemberID("foo")).MemberID)
	assert.Equal(t, MemberID("bar"), raft.GetMember(MemberID("bar")).MemberID)
	assert.Equal(t, MemberID("baz"), raft.GetMember(MemberID("baz")).MemberID)

	// Verify the initial values
	assert.Equal(t, foo, raft.Member())
	assert.Len(t, raft.Members(), 3)
	assert.Equal(t, Term(0), raft.Term())
	assert.Nil(t, raft.Leader())
	assert.Nil(t, raft.LastVotedFor())
	assert.Equal(t, Index(0), raft.CommitIndex())

	// Verify that the term cannot be decreased
	assert.NoError(t, raft.SetTerm(Term(1)))
	assert.Equal(t, Term(1), raft.Term())
	assert.NoError(t, raft.SetTerm(Term(3)))
	assert.Equal(t, Term(3), raft.Term())
	assert.Error(t, raft.SetTerm(Term(2)))

	// Verify that only one vote can be cast per term
	assert.NoError(t, raft.SetLastVotedFor(foo))
	assert.Equal(t, &foo, raft.LastVotedFor())
	assert.Error(t, raft.SetLastVotedFor(bar))
	assert.Equal(t, &foo, raft.LastVotedFor())
	assert.NoError(t, raft.SetTerm(Term(4)))
	assert.Nil(t, raft.Leader())
	assert.Nil(t, raft.LastVotedFor())
	assert.Error(t, raft.SetLastVotedFor("none"))
	assert.Nil(t, raft.LastVotedFor())
	assert.NoError(t, raft.SetLastVotedFor(bar))
	assert.Equal(t, &bar, raft.LastVotedFor())
	assert.Error(t, raft.SetLastVotedFor(""))
	assert.Equal(t, &bar, raft.LastVotedFor())

	// Verify that the leader cannot be changed
	assert.Nil(t, raft.Leader())
	assert.NoError(t, raft.SetLeader(&bar))
	assert.Equal(t, &bar, raft.Leader())
	assert.Error(t, raft.SetLeader(&foo))
	assert.Equal(t, &bar, raft.Leader())
	assert.Equal(t, Term(4), raft.Term())
	assert.NoError(t, raft.SetLeader(nil))
	assert.Equal(t, Term(4), raft.Term())
	assert.Nil(t, raft.Leader())
	assert.Equal(t, &bar, raft.LastVotedFor())

	// Verify that the lastVotedFor and leader are reset when term changes
	assert.NoError(t, raft.SetTerm(Term(5)))
	assert.Nil(t, raft.LastVotedFor())
	assert.Nil(t, raft.Leader())

	// Verify that the state changes once the initial commit index is reached
	assert.Equal(t, Index(0), raft.CommitIndex())
	assert.Equal(t, StatusRunning, raft.Status())
	assert.Equal(t, Index(0), raft.Commit(Index(5))) // Commit a change before setting the first commit index
	assert.Equal(t, Index(5), raft.CommitIndex())
	assert.Equal(t, StatusRunning, raft.Status())
	raft.SetCommitIndex(Index(10)) // Set the first commit index to 10
	assert.Equal(t, StatusRunning, raft.Status())
	raft.SetCommitIndex(Index(50))                   // Ensure the first commit index is idempotent
	assert.Equal(t, Index(5), raft.Commit(Index(9))) // Ensure a commit lower than the first index does not change the node's state
	assert.Equal(t, StatusRunning, raft.Status())
	assert.Equal(t, Index(9), raft.Commit(Index(10))) // Commit the first commit index
	assert.Equal(t, StatusReady, raft.Status())
	assert.Equal(t, StatusReady, <-statusCh)
	assert.Equal(t, Index(10), raft.Commit(Index(3))) // Ensure the commit index cannot be decreased
	assert.Equal(t, Index(10), raft.CommitIndex())

	// Increment the term and vote for later tests
	assert.NoError(t, raft.SetTerm(Term(10)))
	assert.NoError(t, raft.SetLastVotedFor(bar))

	// Verify that the status is changed on close
	assert.NoError(t, raft.Close())
	assert.Equal(t, StatusStopped, raft.Status())
	assert.Equal(t, StatusStopped, <-statusCh)

	// Verify that the cluster state is reloaded from the metadata store when restarted
	follower := &followerRole{&testRole{}}
	leader := &leaderRole{&testRole{}}
	roles = map[RoleType]func(Raft) Role{
		RoleFollower: func(r Raft) Role {
			return follower
		},
		RoleLeader: func(r Raft) Role {
			return leader
		},
	}
	raft = newRaft(NewCluster(cluster), &config.ProtocolConfig{}, &unimplementedClient{}, roles, store)
	assert.Equal(t, StatusStopped, raft.Status())
	raft.WriteLock()
	raft.Init()
	assert.Equal(t, StatusRunning, raft.Status())
	raft.WriteUnlock()
	assert.Equal(t, foo, raft.Member())
	assert.Len(t, raft.Members(), 3)
	assert.Equal(t, Term(10), raft.Term())
	assert.Nil(t, raft.Leader())
	assert.Equal(t, &bar, raft.LastVotedFor())
	assert.Equal(t, Index(0), raft.CommitIndex())

	// Test a role change
	roleCh := make(chan RoleType, 1)
	raft.Watch(func(event Event) {
		if event.Type == EventTypeRole {
			roleCh <- event.Role
		}
	})
	raft.WriteLock()
	raft.SetRole(RoleFollower)
	assert.Equal(t, RoleFollower, raft.Role())
	raft.WriteUnlock()
	assert.False(t, follower.appended)
	_, _ = raft.Append(context.TODO(), &AppendRequest{})
	assert.True(t, follower.appended)

	raft.WriteLock()
	raft.SetRole(RoleLeader)
	assert.Equal(t, RoleLeader, raft.Role())
	assert.Equal(t, RoleLeader, <-roleCh)
	raft.WriteUnlock()
}

type testRole struct {
	Role
	appended bool
}

func (r *testRole) Append(context.Context, *AppendRequest) (*AppendResponse, error) {
	r.appended = true
	return &AppendResponse{}, nil
}

func (r *testRole) Start() error {
	return nil
}

func (r *testRole) Stop() error {
	return nil
}

type followerRole struct {
	*testRole
}

func (r *followerRole) Type() RoleType {
	return RoleFollower
}

type leaderRole struct {
	*testRole
}

func (r *leaderRole) Type() RoleType {
	return RoleLeader
}
