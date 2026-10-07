package presence

import (
	"context"
	"time"

	authdomain "github.com/Dragonqos/lockwood-test/internal/resources/auth/domain"
	transport "github.com/Dragonqos/lockwood-test/pkg/websocket"
)

const SessionExpiredTopic = "session.expired"

type Observer struct {
	sessions authdomain.SessionRepository
	events   transport.EventBus
	interval time.Duration
	done     chan struct{}
}

func NewObserver(sessions authdomain.SessionRepository, events transport.EventBus, interval time.Duration) *Observer {
	return &Observer{sessions: sessions, events: events, interval: interval, done: make(chan struct{})}
}

func (o *Observer) Start(ctx context.Context) {
	go o.run(ctx)
}

func (o *Observer) Stop() {
	select {
	case <-o.done:
	default:
		close(o.done)
	}
}

func (o *Observer) run(ctx context.Context) {
	ticker := time.NewTicker(o.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			o.publishExpired(ctx)
		case <-ctx.Done():
			return
		case <-o.done:
			return
		}
	}
}

func (o *Observer) publishExpired(ctx context.Context) {
	expired, err := o.sessions.Expired(ctx, time.Now())
	if err != nil {
		return
	}
	for _, session := range expired {
		o.events.Publish(SessionExpiredTopic, session)
	}
}
