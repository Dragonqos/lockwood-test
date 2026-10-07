package infra

import (
	"context"
	"sync"
	"testing"

	"github.com/Dragonqos/lockwood-test/internal/resources/chat/domain"
)

func TestMemoryRepositoryEnforcesCapacityAndSingleRoomMembership(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := NewMemoryRepository()
	roomA := domain.Room{ID: "a", Capacity: 1}
	roomB := domain.Room{ID: "b", Capacity: 2}
	if err := repository.CreateAndJoin(ctx, roomA, domain.Membership{RoomID: roomA.ID, UserID: "alice"}); err != nil {
		t.Fatal(err)
	}
	if err := repository.CreateAndJoin(ctx, roomB, domain.Membership{RoomID: roomB.ID, UserID: "bob"}); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		membership domain.Membership
		wantStatus domain.JoinStatus
	}{
		{
			name:       "full room",
			membership: domain.Membership{RoomID: roomA.ID, UserID: "bob"},
			wantStatus: domain.RoomFull,
		},
		{
			name:       "move user to another room",
			membership: domain.Membership{RoomID: roomB.ID, UserID: "alice"},
			wantStatus: domain.JoinOK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, err := repository.Join(ctx, test.membership)
			if err != nil || status != test.wantStatus {
				t.Fatalf("Join() = %q, %v; want %q, nil", status, err, test.wantStatus)
			}
		})
	}
	current, err := repository.CurrentRoom(ctx, "alice")
	if err != nil || current != roomB.ID {
		t.Fatalf("current room = %q, %v", current, err)
	}
}

func TestMemoryRepositoryRejectsDuplicateRoomWithoutChangingMembership(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := NewMemoryRepository()
	room := domain.Room{ID: "room", Capacity: 2}
	if err := repository.CreateAndJoin(ctx, room, domain.Membership{RoomID: room.ID, UserID: "alice"}); err != nil {
		t.Fatal(err)
	}
	if err := repository.CreateAndJoin(ctx, room, domain.Membership{RoomID: room.ID, UserID: "bob"}); err != ErrRoomExists {
		t.Fatalf("duplicate room error = %v, want %v", err, ErrRoomExists)
	}

	tests := []struct {
		name string
		user string
		want domain.RoomID
	}{
		{name: "original member remains", user: "alice", want: room.ID},
		{name: "second member is not added", user: "bob", want: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			current, err := repository.CurrentRoom(ctx, test.user)
			if err != nil || current != test.want {
				t.Fatalf("CurrentRoom(%q) = %q, %v; want %q, nil", test.user, current, err, test.want)
			}
		})
	}
}

func TestMemoryRepositoryJoinAndLeaveLifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := NewMemoryRepository()
	room := domain.Room{ID: "room", Capacity: 2}

	joinTests := []struct {
		name       string
		membership domain.Membership
		wantStatus domain.JoinStatus
	}{
		{
			name:       "missing room",
			membership: domain.Membership{RoomID: room.ID, UserID: "alice"},
			wantStatus: domain.RoomNotFound,
		},
	}
	for _, test := range joinTests {
		t.Run("join "+test.name, func(t *testing.T) {
			status, err := repository.Join(ctx, test.membership)
			if err != nil || status != test.wantStatus {
				t.Fatalf("Join() = %q, %v; want %q, nil", status, err, test.wantStatus)
			}
		})
	}
	if err := repository.CreateAndJoin(ctx, room, domain.Membership{RoomID: room.ID, UserID: "alice"}); err != nil {
		t.Fatal(err)
	}
	joinTests = []struct {
		name       string
		membership domain.Membership
		wantStatus domain.JoinStatus
	}{
		{
			name:       "same room",
			membership: domain.Membership{RoomID: room.ID, UserID: "alice"},
			wantStatus: domain.JoinOK,
		},
		{
			name:       "new member",
			membership: domain.Membership{RoomID: room.ID, UserID: "bob"},
			wantStatus: domain.JoinOK,
		},
	}
	for _, test := range joinTests {
		t.Run("join "+test.name, func(t *testing.T) {
			status, err := repository.Join(ctx, test.membership)
			if err != nil || status != test.wantStatus {
				t.Fatalf("Join() = %q, %v; want %q, nil", status, err, test.wantStatus)
			}
		})
	}

	leaveTests := []struct {
		name     string
		user     string
		want     []string
		roomGone bool
	}{
		{name: "first member", user: "alice", want: []string{"bob"}},
		{name: "last member", user: "bob", roomGone: true},
	}
	for _, test := range leaveTests {
		t.Run("leave "+test.name, func(t *testing.T) {
			if err := repository.Leave(ctx, domain.Membership{RoomID: room.ID, UserID: test.user}); err != nil {
				t.Fatal(err)
			}
			members, err := repository.Members(ctx, room.ID)
			if err != nil {
				t.Fatal(err)
			}
			if test.roomGone {
				if members != nil {
					t.Fatalf("Members() = %#v, want nil", members)
				}
				return
			}
			if len(members) != len(test.want) || members[0] != test.want[0] {
				t.Fatalf("Members() = %#v, want %#v", members, test.want)
			}
		})
	}
}

func TestMemoryRepositoryJoinCapacityIsAtomic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := NewMemoryRepository()
	room := domain.Room{ID: "room", Capacity: 8}
	if err := repository.CreateAndJoin(ctx, room, domain.Membership{RoomID: room.ID, UserID: "owner"}); err != nil {
		t.Fatal(err)
	}

	const joiners = 32
	statuses := make(chan domain.JoinStatus, joiners)
	var waitGroup sync.WaitGroup
	for i := 0; i < joiners; i++ {
		waitGroup.Add(1)
		go func(i int) {
			defer waitGroup.Done()
			status, err := repository.Join(ctx, domain.Membership{
				RoomID: room.ID,
				UserID: string(rune('a' + i)),
			})
			if err != nil {
				t.Errorf("join %d: %v", i, err)
				return
			}
			statuses <- status
		}(i)
	}
	waitGroup.Wait()
	close(statuses)

	var joined, full int
	for status := range statuses {
		switch status {
		case domain.JoinOK:
			joined++
		case domain.RoomFull:
			full++
		default:
			t.Fatalf("unexpected join status %q", status)
		}
	}
	if joined != room.Capacity-1 || full != joiners-(room.Capacity-1) {
		t.Fatalf("join results: joined=%d full=%d", joined, full)
	}

	members, err := repository.Members(ctx, room.ID)
	if err != nil || len(members) != room.Capacity {
		t.Fatalf("members after concurrent joins = %d, %v", len(members), err)
	}
}
