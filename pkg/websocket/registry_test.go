package websocket

import (
	"context"
	"testing"
)

func TestLocalConnectionRegistryLifecycle(t *testing.T) {
	t.Parallel()

	registry := NewLocalConnectionRegistry()
	first := NewClient(nil, "session-1", 1)
	second := NewClient(nil, "session-2", 1)

	registry.Register(first.SessionID(), first)
	registry.Register(second.SessionID(), second)

	tests := []struct {
		name string
		id   string
		want *Client
	}{
		{name: "first client", id: first.SessionID(), want: first},
		{name: "second client", id: second.SessionID(), want: second},
		{name: "unknown client", id: "missing", want: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := registry.Client(test.id); got != test.want {
				t.Fatalf("Client(%q) = %p, want %p", test.id, got, test.want)
			}
		})
	}

	registry.Unregister(first.SessionID())
	if got := registry.Client(first.SessionID()); got != nil {
		t.Fatalf("unregistered client = %p, want nil", got)
	}
}

func TestLocalClientFactoryBuildsAndRegistersClient(t *testing.T) {
	t.Parallel()

	connections := NewLocalConnectionRegistry()
	factory := NewLocalClientFactory(connections, 4, nil)
	client, sessionID := factory.Build(context.Background(), nil)

	if client == nil {
		t.Fatal("Build() returned nil client")
	}
	if sessionID == "" || client.SessionID() != sessionID {
		t.Fatalf("session ID = %q, client ID = %q", sessionID, client.SessionID())
	}
	if got := connections.Client(sessionID); got != client {
		t.Fatalf("registered client = %p, want %p", got, client)
	}
}

func TestLocalSubscriptionRegistryForwardsEventsToClient(t *testing.T) {
	t.Parallel()

	events := NewLocalEventBus()
	connections := NewLocalConnectionRegistry()
	subscriptions := NewLocalSubscriptionRegistry(events, connections)
	client := NewClient(nil, "session-1", 1)
	connections.Register(client.SessionID(), client)

	subscriptions.Subscribe("room-1", client.SessionID())
	want := map[string]string{"cmd": "user_joined"}
	events.Publish("room-1", want)

	select {
	case got := <-client.send:
		if got.(map[string]string)["cmd"] != want["cmd"] {
			t.Fatalf("event = %#v, want %#v", got, want)
		}
	default:
		t.Fatal("client did not receive event")
	}

	subscriptions.Unsubscribe("room-1", client.SessionID())
	events.Publish("room-1", want)
	select {
	case got := <-client.send:
		t.Fatalf("received event after unsubscribe: %#v", got)
	default:
	}
}
