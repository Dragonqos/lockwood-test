package websocket

import (
	"context"

	chatpublic "github.com/Dragonqos/lockwood-test/internal/resources/chat/app/public"
)

func (h *Handler) leaveRoom(ctx context.Context, sessionID string, request Request) {
	client := h.connections.Client(sessionID)
	if client == nil {
		return
	}

	userID, ok := h.authorize(ctx, sessionID, request.RID, "leave_room")
	if !ok {
		return
	}

	response, err := h.chatPublic.LeaveRoom(ctx, chatpublic.LeaveRoomRequest{UserID: userID})
	if err != nil {
		client.SendResponse(request.RID, "leave_room", "internal_error")
		return
	}

	if response.RoomID != "" {
		h.subscriptions.Unsubscribe(string(response.RoomID), sessionID)
	}

	h.publishEvents(response.Events)

	client.SendResponse(request.RID, "leave_room", response.Status)
}
