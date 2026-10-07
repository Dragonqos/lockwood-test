package websocket

import (
	"context"

	chatpublic "github.com/Dragonqos/lockwood-test/internal/resources/chat/app/public"
)

func (h *Handler) createAndJoinRoom(ctx context.Context, sessionID string, request Request) {
	client := h.connections.Client(sessionID)
	if client == nil {
		return
	}

	userID, ok := h.authorize(ctx, sessionID, request.RID, "create_room")
	if !ok {
		return
	}

	response, err := h.chatPublic.CreateAndJoinRoom(ctx, chatpublic.CreateRoomRequest{UserID: userID, Name: request.Name})
	if err != nil {
		client.SendResponse(request.RID, "create_room", "internal_error")
		return
	}

	if response.Status != "ok" {
		client.SendResponse(request.RID, "create_room", response.Status)
		return
	}

	if response.PreviousRoomID != "" && response.PreviousRoomID != response.RoomID {
		h.subscriptions.Unsubscribe(string(response.PreviousRoomID), sessionID)
	}

	h.subscriptions.Subscribe(string(response.RoomID), sessionID)
	h.publishEvents(response.Events)

	// IMPORTANT: The request command is `create_and_join_room`, so `create_room` in the task's
	// response example looks like a specification typo. Keep it for compatibility.
	client.SendResponse(request.RID, "create_room", response.Status)
}
