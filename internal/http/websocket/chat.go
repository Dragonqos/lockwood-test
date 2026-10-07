package websocket

import (
	"context"

	chatpublic "github.com/Dragonqos/lockwood-test/internal/resources/chat/app/public"
)

func (h *Handler) chat(ctx context.Context, sessionID string, request Request) {
	client := h.connections.Client(sessionID)
	if client == nil {
		return
	}

	userID, ok := h.authorize(ctx, sessionID, request.RID, "chat")
	if !ok {
		return
	}

	response, err := h.chatPublic.Chat(ctx, chatpublic.ChatRequest{UserID: userID, Message: request.Message})
	if err != nil {
		client.SendResponse(request.RID, "chat", "internal_error")
		return
	}

	h.publishEvents(response.Events)
	client.SendResponse(request.RID, "chat", response.Status)
}
