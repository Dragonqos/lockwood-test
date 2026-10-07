package websocket

import (
	"sync"

	gorilla "github.com/gorilla/websocket"
)

type ConnectionRegistry interface {
	Register(sessionID string, client *Client)
	Unregister(sessionID string)
	Client(sessionID string) *Client
	CloseAll()
}

type LocalConnectionRegistry struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func NewLocalConnectionRegistry() *LocalConnectionRegistry {
	return &LocalConnectionRegistry{clients: make(map[string]*Client)}
}

func (r *LocalConnectionRegistry) Register(sessionID string, client *Client) {
	r.mu.Lock()
	r.clients[sessionID] = client
	r.mu.Unlock()
}

func (r *LocalConnectionRegistry) Unregister(sessionID string) {
	r.mu.Lock()
	delete(r.clients, sessionID)
	r.mu.Unlock()
}

func (r *LocalConnectionRegistry) Client(sessionID string) *Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.clients[sessionID]
}

func (r *LocalConnectionRegistry) CloseAll() {
	r.mu.Lock()
	clients := make([]*Client, 0, len(r.clients))
	for sessionID, client := range r.clients {
		clients = append(clients, client)
		delete(r.clients, sessionID)
	}
	r.mu.Unlock()

	for _, client := range clients {
		client.Close(gorilla.CloseGoingAway, "server shutdown")
	}
}
