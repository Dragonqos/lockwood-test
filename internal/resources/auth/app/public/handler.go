package public

import (
	authdomain "github.com/Dragonqos/lockwood-test/internal/resources/auth/domain"
)

type Handler struct {
	sessions authdomain.SessionRepository
}

func NewHandler(sessions authdomain.SessionRepository) *Handler {
	return &Handler{sessions: sessions}
}
