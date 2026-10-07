package websocket

import (
	"context"

	"github.com/google/uuid"
	gorilla "github.com/gorilla/websocket"
)

type ClientFactory interface {
	Build(ctx context.Context, conn *gorilla.Conn) (*Client, string)
}

type LocalClientFactory struct {
	connections ConnectionRegistry
	queueSize   int
	onActivity  func(ctx context.Context, sessionID string)
}

func NewLocalClientFactory(connections ConnectionRegistry, queueSize int, onActivity func(ctx context.Context, sessionID string)) *LocalClientFactory {
	return &LocalClientFactory{
		connections: connections,
		queueSize:   queueSize,
		onActivity:  onActivity,
	}
}

func (f *LocalClientFactory) Build(ctx context.Context, conn *gorilla.Conn) (*Client, string) {
	sessionID := uuid.NewString()
	client := NewClient(conn, sessionID, f.queueSize)

	f.connections.Register(sessionID, client)
	if f.onActivity != nil {
		client.SetActivityHandler(func() {
			f.onActivity(ctx, sessionID)
		})
	}

	return client, sessionID
}
