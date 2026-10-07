package presence

import (
	"context"
	"testing"

	authdomain "github.com/Dragonqos/lockwood-test/internal/resources/auth/domain"
	authinfra "github.com/Dragonqos/lockwood-test/internal/resources/auth/infra"
	chatapp "github.com/Dragonqos/lockwood-test/internal/resources/chat/app"
	chatdomain "github.com/Dragonqos/lockwood-test/internal/resources/chat/domain"
	chatinfra "github.com/Dragonqos/lockwood-test/internal/resources/chat/infra"
	transport "github.com/Dragonqos/lockwood-test/pkg/websocket"
)

func TestCleanupKeepsMembershipWhenSessionIsReplaced(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	sessions := authinfra.NewMemorySessionRepository()
	rooms := chatinfra.NewMemoryRepository()
	chat := chatapp.NewChatCore(rooms)
	events := transport.NewLocalEventBus()
	cleanup := NewCleanup(sessions, chat, events)

	room := chatdomain.Room{ID: "room-1", Capacity: 8}
	if err := rooms.CreateAndJoin(ctx, room, chatdomain.Membership{RoomID: room.ID, UserID: "alice"}); err != nil {
		t.Fatal(err)
	}
	if err := sessions.Claim(ctx, "alice", "old-session"); err != nil {
		t.Fatal(err)
	}

	var left int
	events.Subscribe(string(room.ID), "observer", func(any) { left++ })
	cleanup.Handle(authdomain.SessionReplaced{UserID: "alice", PreviousSession: "old-session"})

	current, err := chat.RoomForUser(ctx, "alice")
	if err != nil || current != room.ID {
		t.Fatalf("current room = %q, %v; want %q, nil", current, err, room.ID)
	}
	if left != 0 {
		t.Fatalf("user_left events = %d, want 0", left)
	}

	var oldSessionEvents int
	events.Subscribe(string(room.ID), "old-session", func(any) { oldSessionEvents++ })
	cleanup.Handle(authdomain.SessionReplaced{UserID: "alice", PreviousSession: "old-session"})
	events.Publish(string(room.ID), "event")
	if oldSessionEvents != 0 {
		t.Fatalf("old session received %d events after replacement", oldSessionEvents)
	}
}
