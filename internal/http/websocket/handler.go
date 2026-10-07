package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/go-playground/validator/v10"
	gorilla "github.com/gorilla/websocket"

	authpublic "github.com/Dragonqos/lockwood-test/internal/resources/auth/app/public"
	authdomain "github.com/Dragonqos/lockwood-test/internal/resources/auth/domain"
	chatpublic "github.com/Dragonqos/lockwood-test/internal/resources/chat/app/public"
	transport "github.com/Dragonqos/lockwood-test/pkg/websocket"
)

const (
	maxMessageSize = 4096
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = pongWait * 9 / 10
)

type Handler struct {
	upgrader      gorilla.Upgrader
	authPublic    *authpublic.Handler
	chatPublic    *chatpublic.Handler
	sessions      authdomain.SessionRepository
	validator     *validator.Validate
	eventBus      transport.EventBus
	connections   transport.ConnectionRegistry
	subscriptions transport.SubscriptionRegistry
	clientFactory transport.ClientFactory
}

func NewHandler(
	authPublic *authpublic.Handler,
	chatPublic *chatpublic.Handler,
	sessions authdomain.SessionRepository,
	requestValidator *validator.Validate,
	eventBus transport.EventBus,
	connections transport.ConnectionRegistry,
	subscriptions transport.SubscriptionRegistry,
	clientFactory transport.ClientFactory,
) http.Handler {
	if requestValidator == nil {
		panic("validator must not be nil")
	}

	return &Handler{
		upgrader: gorilla.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     sameOrigin,
		},
		authPublic:    authPublic,
		chatPublic:    chatPublic,
		sessions:      sessions,
		validator:     requestValidator,
		eventBus:      eventBus,
		connections:   connections,
		subscriptions: subscriptions,
		clientFactory: clientFactory,
	}
}

func sameOrigin(request *http.Request) bool {
	origin := request.Header.Get("Origin")
	if origin == "" {
		return true
	}

	originURL, err := url.Parse(origin)
	return err == nil && originURL.Host == request.Host
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	connectionCtx, cancel := context.WithCancel(r.Context())
	defer cancel()

	client, sessionID := h.clientFactory.Build(connectionCtx, conn)
	go client.WriteLoop(writeWait, pingPeriod)
	go client.ReadLoop(maxMessageSize, pongWait)

	for payload := range client.Requests() {
		var request Request
		if err := json.Unmarshal(payload, &request); err != nil {
			client.SendResponse(nil, "", "bad_request")
			continue
		}
		h.handle(connectionCtx, sessionID, request)
	}

	h.disconnect(connectionCtx, sessionID)
}

func (h *Handler) handle(ctx context.Context, sessionID string, request Request) {
	client := h.connections.Client(sessionID)
	if client == nil {
		return
	}

	if err := validateRequest(h.validator, request); err != nil {
		client.SendResponse(request.RID, request.Cmd, "bad_request")
		return
	}

	switch request.Cmd {
	case "auth":
		h.auth(ctx, sessionID, request)
	case "create_and_join_room":
		h.createAndJoinRoom(ctx, sessionID, request)
	case "join_room":
		h.joinRoom(ctx, sessionID, request)
	case "leave_room":
		h.leaveRoom(ctx, sessionID, request)
	case "chat":
		h.chat(ctx, sessionID, request)
	default:
		client.SendResponse(request.RID, request.Cmd, "bad_request")
	}
}

func (h *Handler) disconnect(ctx context.Context, sessionID string) {
	client := h.connections.Client(sessionID)
	if client == nil {
		return
	}
	client.Close(gorilla.CloseNormalClosure, "")
	h.connections.Unregister(sessionID)
	_ = h.sessions.MarkDisconnected(ctx, sessionID, time.Now().Add(30*time.Second))
}
