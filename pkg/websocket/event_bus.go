package websocket

import "sync"

type Subscriber func(event any)

type EventBus interface {
	Subscribe(topicID, subscriberID string, subscriber Subscriber)
	Unsubscribe(topicID, subscriberID string)
	Publish(topicID string, event any)
}

type LocalEventBus struct {
	mu            sync.RWMutex
	subscribers   map[string]map[string]Subscriber
	subscriptions map[string]string
}

func NewLocalEventBus() *LocalEventBus {
	return &LocalEventBus{
		subscribers:   make(map[string]map[string]Subscriber),
		subscriptions: make(map[string]string),
	}
}

func (b *LocalEventBus) Subscribe(topicID, subscriberID string, subscriber Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.removeSubscriptionLocked(subscriberID)
	if topicID == "" {
		return
	}
	if b.subscribers[topicID] == nil {
		b.subscribers[topicID] = make(map[string]Subscriber)
	}
	b.subscribers[topicID][subscriberID] = subscriber
	b.subscriptions[subscriberID] = topicID
}

func (b *LocalEventBus) Unsubscribe(topicID, subscriberID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subscriptions[subscriberID] == topicID {
		b.removeSubscriptionLocked(subscriberID)
	}
}

func (b *LocalEventBus) Publish(topicID string, event any) {
	b.mu.RLock()
	subscribers := make([]Subscriber, 0, len(b.subscribers[topicID]))
	for _, subscriber := range b.subscribers[topicID] {
		subscribers = append(subscribers, subscriber)
	}
	b.mu.RUnlock()

	for _, subscriber := range subscribers {
		subscriber(event)
	}
}

func (b *LocalEventBus) removeSubscriptionLocked(subscriberID string) {
	topicID := b.subscriptions[subscriberID]
	if topicID == "" {
		return
	}
	delete(b.subscriptions, subscriberID)
	if subscribers := b.subscribers[topicID]; subscribers != nil {
		delete(subscribers, subscriberID)
		if len(subscribers) == 0 {
			delete(b.subscribers, topicID)
		}
	}
}
