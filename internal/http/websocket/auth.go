package websocket

import (
	"context"

	authpublic "github.com/Dragonqos/lockwood-test/internal/resources/auth/app/public"
	authdomain "github.com/Dragonqos/lockwood-test/internal/resources/auth/domain"
	gorilla "github.com/gorilla/websocket"
)

func (h *Handler) authorize(ctx context.Context, sessionID string, rid any, command string) (string, bool) {
	client := h.connections.Client(sessionID)
	if client == nil {
		return "", false
	}

	userID, err := h.authPublic.Authorize(ctx, sessionID)
	if err != nil {
		client.SendResponse(rid, command, "unauthorized")
		return "", false
	}

	return userID, true
}

func (h *Handler) auth(ctx context.Context, sessionID string, request Request) {
	client := h.connections.Client(sessionID)
	if client == nil {
		return
	}

	response, err := h.authPublic.Authenticate(ctx, authpublic.Request{
		SessionID: sessionID,
		User:      request.User,
	})
	if err != nil {
		client.SendResponse(request.RID, "auth", "internal_error")
		return
	}

	if response.Status != "ok" {
		client.SendResponse(request.RID, "auth", response.Status)
		return
	}

	roomID, err := h.chatPublic.CurrentRoom(ctx, response.UserID)
	if err != nil {
		client.SendResponse(request.RID, "auth", "internal_error")
		return
	}
	if roomID != "" {
		h.subscriptions.Subscribe(string(roomID), sessionID)
	}

	if response.PreviousSession != "" && response.PreviousSession != sessionID {
		h.eventBus.Publish(authdomain.SessionReplacedTopic, authdomain.SessionReplaced{
			UserID:          response.UserID,
			PreviousSession: response.PreviousSession,
		})
		if previousClient := h.connections.Client(response.PreviousSession); previousClient != nil {
			previousClient.Close(gorilla.ClosePolicyViolation, "session replaced")
		}
	}

	client.SendResponse(request.RID, "auth", response.Status)
}
