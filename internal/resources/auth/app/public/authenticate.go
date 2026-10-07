package public

import (
	"context"
	"strings"

	authdomain "github.com/Dragonqos/lockwood-test/internal/resources/auth/domain"
)

type Request struct {
	SessionID string
	User      string
}

type Response struct {
	Status          string
	UserID          string
	PreviousSession string
}

func (h *Handler) Authenticate(ctx context.Context, request Request) (Response, error) {
	identity := h.identity(request)
	if identity.UserID == "" {
		return Response{Status: "unauthorized"}, nil
	}

	// product requirement - only one user
	previousSession, _ := h.sessions.Current(ctx, identity.UserID)
	if err := h.sessions.Claim(ctx, identity.UserID, request.SessionID); err != nil {
		return Response{Status: "unauthorized"}, nil
	}

	return Response{
		Status:          "ok",
		UserID:          identity.UserID,
		PreviousSession: previousSession,
	}, nil
}

func (h *Handler) identity(request Request) authdomain.Identity {
	// TODO(auth): here we should add auth.Authenticate(request.Token) in real scenario
	user := strings.TrimSpace(request.User)
	return authdomain.Identity{UserID: user, Username: user}
}
