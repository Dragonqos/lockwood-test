package app

import (
	"context"
	"errors"
	"testing"

	"github.com/Dragonqos/lockwood-test/internal/resources/chat/domain"
)

type repositoryStub struct {
	joinStatus domain.JoinStatus
	joinErr    error
	current    domain.RoomID
	members    []string
	lastRoom   domain.Room
	lastMember domain.Membership
}

func (s *repositoryStub) CreateAndJoin(_ context.Context, room domain.Room, membership domain.Membership) error {
	s.lastRoom, s.lastMember = room, membership
	return nil
}
func (s *repositoryStub) Join(_ context.Context, membership domain.Membership) (domain.JoinStatus, error) {
	s.lastMember = membership
	return s.joinStatus, s.joinErr
}
func (s *repositoryStub) Leave(_ context.Context, membership domain.Membership) error {
	s.lastMember = membership
	return nil
}
func (s *repositoryStub) CurrentRoom(context.Context, string) (domain.RoomID, error) {
	return s.current, nil
}
func (s *repositoryStub) Members(context.Context, domain.RoomID) ([]string, error) {
	return s.members, nil
}

func TestChatCoreDelegatesRepositoryOperations(t *testing.T) {
	ctx := context.Background()
	stub := &repositoryStub{joinStatus: domain.JoinOK, current: "room-1", members: []string{"alice", "bob"}}
	core := NewChatCore(stub)
	room := domain.Room{ID: "room-1", Capacity: 8}
	membership := domain.Membership{RoomID: room.ID, UserID: "alice"}

	if err := core.CreateAndJoinRoom(ctx, CreateAndJoinCommand{Room: room, Membership: membership}); err != nil {
		t.Fatal(err)
	}
	if stub.lastRoom != room || stub.lastMember != membership {
		t.Fatal("CreateAndJoinRoom did not delegate arguments")
	}
	if status, err := core.JoinRoom(ctx, JoinCommand{Membership: membership}); err != nil || status != domain.JoinOK {
		t.Fatalf("JoinRoom() = %q, %v", status, err)
	}
	if err := core.LeaveRoom(ctx, LeaveCommand{Membership: membership}); err != nil {
		t.Fatal(err)
	}
	if current, err := core.RoomForUser(ctx, "alice"); err != nil || current != room.ID {
		t.Fatalf("RoomForUser() = %q, %v", current, err)
	}
	if members, err := core.ListMembers(ctx, room.ID); err != nil || len(members) != 2 {
		t.Fatalf("ListMembers() = %#v, %v", members, err)
	}
}

func TestChatCorePropagatesJoinError(t *testing.T) {
	wantErr := errors.New("repository unavailable")
	core := NewChatCore(&repositoryStub{joinErr: wantErr})
	_, err := core.JoinRoom(context.Background(), JoinCommand{
		Membership: domain.Membership{RoomID: "room-1", UserID: "alice"},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("JoinRoom() error = %v, want %v", err, wantErr)
	}
}
