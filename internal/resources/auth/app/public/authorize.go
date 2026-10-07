package public

import (
	"context"
	"errors"
)

func (h *Handler) Authorize(ctx context.Context, sessionID string) (string, error) {
	userID, err := h.sessions.UserID(ctx, sessionID)
	if err != nil || userID == "" {
		return "", errors.New("unauthorized")
	}

	current, err := h.sessions.Current(ctx, userID)
	if err != nil || current != sessionID {
		return "", errors.New("unauthorized")
	}

	return userID, nil
}
