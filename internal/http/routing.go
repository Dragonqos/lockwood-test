package httpapp

import (
	"net/http"

	"github.com/Dragonqos/lockwood-test/internal/http/openapi"
	"github.com/gorilla/mux"
)

type RouterDependencies struct {
	Handlers openapi.ServerInterface
}

func NewRouter(deps RouterDependencies) *mux.Router {
	router := mux.NewRouter()
	router.Use(securityHeaders)
	router.HandleFunc("/healthz", healthz).Methods(http.MethodGet).Name("healthz")

	openapi.HandlerFromMux(deps.Handlers, router)
	return router
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}
