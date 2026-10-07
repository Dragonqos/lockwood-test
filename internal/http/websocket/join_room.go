package websocket

import (
	"context"

	chatpublic "github.com/Dragonqos/lockwood-test/internal/resources/chat/app/public"
)

func (h *Handler) joinRoom(ctx context.Context, sessionID string, request Request) {
	client := h.connections.Client(sessionID)
	if client == nil {
		return
	}

	userID, ok := h.authorize(ctx, sessionID, request.RID, "join_room")
	if !ok {
		return
	}

	response, err := h.chatPublic.JoinRoom(ctx, chatpublic.JoinRoomRequest{UserID: userID, Name: request.Name})
	if err != nil {
		client.SendResponse(request.RID, "join_room", "internal_error")
		return
	}

	if response.Status != "ok" {
		client.SendResponse(request.RID, "join_room", response.Status)
		return
	}

	if response.PreviousRoomID != "" && response.PreviousRoomID != response.RoomID {
		h.subscriptions.Unsubscribe(string(response.PreviousRoomID), sessionID)
	}

	h.subscriptions.Subscribe(string(response.RoomID), sessionID)
	h.publishEvents(response.Events)

	client.SendResponse(request.RID, "join_room", response.Status)
}
