package httpapp

import (
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/Dragonqos/lockwood-test/internal/http/openapi"
	websocketapp "github.com/Dragonqos/lockwood-test/internal/http/websocket"
	authpublic "github.com/Dragonqos/lockwood-test/internal/resources/auth/app/public"
	authdomain "github.com/Dragonqos/lockwood-test/internal/resources/auth/domain"
	chatpublic "github.com/Dragonqos/lockwood-test/internal/resources/chat/app/public"
	transport "github.com/Dragonqos/lockwood-test/pkg/websocket"
)

type Handlers struct {
	webSocket http.Handler
}

func NewHandlers(authPublic *authpublic.Handler, chatPublic *chatpublic.Handler, sessions authdomain.SessionRepository, requestValidator *validator.Validate, eventBus transport.EventBus, connections transport.ConnectionRegistry, subscriptions transport.SubscriptionRegistry, clientFactory transport.ClientFactory) openapi.ServerInterface {
	return Handlers{
		webSocket: websocketapp.NewHandler(authPublic, chatPublic, sessions, requestValidator, eventBus, connections, subscriptions, clientFactory),
	}
}

func (Handlers) Healthcheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func (h Handlers) GetWs(w http.ResponseWriter, r *http.Request) {
	h.webSocket.ServeHTTP(w, r)
}
