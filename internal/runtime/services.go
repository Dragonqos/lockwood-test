package runtime

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/Dragonqos/lockwood-test/internal/app/presence"
	httpapp "github.com/Dragonqos/lockwood-test/internal/http"
	authpublic "github.com/Dragonqos/lockwood-test/internal/resources/auth/app/public"
	authdomain "github.com/Dragonqos/lockwood-test/internal/resources/auth/domain"
	authinfra "github.com/Dragonqos/lockwood-test/internal/resources/auth/infra"
	chatapp "github.com/Dragonqos/lockwood-test/internal/resources/chat/app"
	chatpublic "github.com/Dragonqos/lockwood-test/internal/resources/chat/app/public"
	chatdomain "github.com/Dragonqos/lockwood-test/internal/resources/chat/domain"
	chatinfra "github.com/Dragonqos/lockwood-test/internal/resources/chat/infra"
	transport "github.com/Dragonqos/lockwood-test/pkg/websocket"
)

type Config struct {
	HTTPAddr string
}

type Services struct {
	Config      Config
	Chat        chatapp.ChatCoreAPI
	Validator   *validator.Validate
	Rooms       chatdomain.Repository
	Sessions    authdomain.SessionRepository
	EventBus    transport.EventBus
	Connections transport.ConnectionRegistry
	Observer    *presence.Observer
	Handlers    http.Handler
}

func NewServices(config Config) (*Services, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Repo
	var (
		roomsRepo    = chatinfra.NewMemoryRepository()
		sessionsRepo = authinfra.NewMemorySessionRepository()
	)

	// Services
	var (
		v             = validator.New()
		chatService   = chatapp.NewChatCore(roomsRepo)
		eventBus      = transport.NewLocalEventBus()
		connections   = transport.NewLocalConnectionRegistry()
		subscriptions = transport.NewLocalSubscriptionRegistry(eventBus, connections)
		clientFactory = transport.NewLocalClientFactory(connections, 32, func(ctx context.Context, sessionID string) {
			_ = sessionsRepo.Touch(ctx, sessionID)
		})
		authPublic = authpublic.NewHandler(sessionsRepo)
		chatPublic = chatpublic.NewHandler(chatService)
		handlers   = httpapp.NewHandlers(authPublic, chatPublic, sessionsRepo, v, eventBus, connections, subscriptions, clientFactory)
	)

	// Observers
	var (
		observer = presence.NewObserver(sessionsRepo, eventBus, time.Second)
		cleanup  = presence.NewCleanup(sessionsRepo, chatService, eventBus)
	)

	eventBus.Subscribe(presence.SessionExpiredTopic, "session-cleanup-expired", cleanup.Handle)
	eventBus.Subscribe(authdomain.SessionReplacedTopic, "session-cleanup-replaced", cleanup.Handle)
	observer.Start(context.Background())

	return &Services{
		Config:      config,
		Chat:        chatService,
		Validator:   v,
		Rooms:       roomsRepo,
		Sessions:    sessionsRepo,
		EventBus:    eventBus,
		Connections: connections,
		Observer:    observer,
		Handlers:    httpapp.NewRouter(httpapp.RouterDependencies{Handlers: handlers}),
	}, nil
}

func (s *Services) Close() error {
	if s == nil {
		return nil
	}
	if s.Observer != nil {
		s.Observer.Stop()
	}
	if s.Connections != nil {
		s.Connections.CloseAll()
	}
	return nil
}

func (config Config) Validate() error {
	if config.HTTPAddr == "" {
		return fmt.Errorf("HTTP_ADDR is required")
	}
	return nil
}
