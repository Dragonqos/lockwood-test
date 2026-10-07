package public

import (
	"context"
	"strings"

	chatapp "github.com/Dragonqos/lockwood-test/internal/resources/chat/app"
	chatdomain "github.com/Dragonqos/lockwood-test/internal/resources/chat/domain"
)

func (h *Handler) JoinRoom(ctx context.Context, request JoinRoomRequest) (JoinRoomResponse, error) {
	if request.UserID == "" || strings.TrimSpace(request.Name) == "" {
		return JoinRoomResponse{Status: "bad_request"}, nil
	}

	oldRoom, err := h.coreApi.RoomForUser(ctx, request.UserID)
	if err != nil {
		return JoinRoomResponse{}, err
	}
	roomID := chatdomain.RoomID(request.Name)
	status, err := h.coreApi.JoinRoom(ctx, chatapp.JoinCommand{
		Membership: chatdomain.Membership{RoomID: roomID, UserID: request.UserID},
	})
	if err != nil {
		return JoinRoomResponse{PreviousRoomID: oldRoom}, err
	}
	if status != chatdomain.JoinOK {
		return JoinRoomResponse{Status: string(status), PreviousRoomID: oldRoom}, nil
	}

	events := make([]RoomEvent, 0, 2)
	if oldRoom != "" && oldRoom != roomID {
		events = append(events, RoomEvent{RoomID: oldRoom, Cmd: "user_left", User: request.UserID})
	}
	events = append(events, RoomEvent{RoomID: roomID, Cmd: "user_joined", User: request.UserID})
	return JoinRoomResponse{Status: "ok", RoomID: roomID, PreviousRoomID: oldRoom, Events: events}, nil
}
