package websocket

import (
	"sync/atomic"
	"testing"
)

func TestLocalEventBusPublishesToSubscribers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		topic string
		id    string
	}{
		{name: "room topic", topic: "room-1", id: "session-1"},
		{name: "another room topic", topic: "room-2", id: "session-2"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bus := NewLocalEventBus()
			events := make(chan any, 1)
			bus.Subscribe(test.topic, test.id, func(event any) { events <- event })

			want := "user_joined"
			bus.Publish(test.topic, want)

			select {
			case got := <-events:
				if got != want {
					t.Fatalf("event = %#v, want %#v", got, want)
				}
			default:
				t.Fatal("subscriber did not receive event")
			}
		})
	}
}

func TestLocalEventBusMovesSubscriberBetweenTopics(t *testing.T) {
	t.Parallel()

	bus := NewLocalEventBus()
	var received atomic.Int64
	bus.Subscribe("room-a", "session-1", func(any) { received.Add(1) })
	bus.Subscribe("room-b", "session-1", func(any) { received.Add(10) })

	bus.Publish("room-a", "old")
	bus.Publish("room-b", "current")

	if got := received.Load(); got != 10 {
		t.Fatalf("received score = %d, want 10", got)
	}
}

func TestLocalEventBusUnsubscribeDoesNotRemoveNewTopic(t *testing.T) {
	t.Parallel()

	bus := NewLocalEventBus()
	var received atomic.Int64
	bus.Subscribe("room-a", "session-1", func(any) { received.Add(1) })
	bus.Subscribe("room-b", "session-1", func(any) { received.Add(10) })
	bus.Unsubscribe("room-a", "session-1")

	bus.Publish("room-b", "current")
	if got := received.Load(); got != 10 {
		t.Fatalf("received score = %d, want 10", got)
	}
}
