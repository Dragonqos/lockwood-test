package presence

import (
	"context"

	authdomain "github.com/Dragonqos/lockwood-test/internal/resources/auth/domain"
	chatapp "github.com/Dragonqos/lockwood-test/internal/resources/chat/app"
	chatdomain "github.com/Dragonqos/lockwood-test/internal/resources/chat/domain"
	transport "github.com/Dragonqos/lockwood-test/pkg/websocket"
)

type Cleanup struct {
	sessions authdomain.SessionRepository
	chat     chatapp.ChatCoreAPI
	events   transport.EventBus
}

func NewCleanup(sessions authdomain.SessionRepository, chat chatapp.ChatCoreAPI, events transport.EventBus) *Cleanup {
	return &Cleanup{sessions: sessions, chat: chat, events: events}
}

func (c *Cleanup) Handle(raw any) {
	sessionID, userID, release := c.eventDetails(raw)
	if sessionID == "" || userID == "" {
		return
	}

	ctx := context.Background()
	roomID, err := c.chat.RoomForUser(ctx, userID)
	if err != nil || roomID == "" {
		if release {
			_ = c.sessions.Release(ctx, userID, sessionID)
		}
		return
	}
	if !release {
		// Replacing a socket is not the same as leaving the room. Keep the
		// membership during the reconnect grace period and only remove the
		// stale subscription.
		c.events.Unsubscribe(string(roomID), sessionID)
		return
	}

	if err := c.chat.LeaveRoom(ctx, chatapp.LeaveCommand{
		Membership: chatdomain.Membership{RoomID: roomID, UserID: userID},
	}); err != nil {
		return
	}

	c.events.Unsubscribe(string(roomID), sessionID)
	c.events.Publish(string(roomID), struct {
		Cmd  string `json:"cmd"`
		User string `json:"user"`
	}{Cmd: "user_left", User: userID})

	_ = c.sessions.Release(ctx, userID, sessionID)
}

func (c *Cleanup) eventDetails(raw any) (sessionID, userID string, release bool) {
	switch event := raw.(type) {
	case authdomain.Session:
		return event.ID, event.UserID, true
	case authdomain.SessionReplaced:
		return event.PreviousSession, event.UserID, false
	default:
		return "", "", false
	}
}
