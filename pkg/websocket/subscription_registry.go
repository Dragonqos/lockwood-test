package websocket

type SubscriptionRegistry interface {
	Subscribe(topicID, sessionID string)
	Unsubscribe(topicID, sessionID string)
}

type LocalSubscriptionRegistry struct {
	events      EventBus
	connections ConnectionRegistry
}

func NewLocalSubscriptionRegistry(events EventBus, connections ConnectionRegistry) *LocalSubscriptionRegistry {
	return &LocalSubscriptionRegistry{events: events, connections: connections}
}

func (r *LocalSubscriptionRegistry) Subscribe(topicID, sessionID string) {
	r.events.Subscribe(topicID, sessionID, func(update any) {
		if client := r.connections.Client(sessionID); client != nil {
			client.SendMessage(update)
		}
	})
}

func (r *LocalSubscriptionRegistry) Unsubscribe(topicID, sessionID string) {
	r.events.Unsubscribe(topicID, sessionID)
}
