package public

import (
	"context"
	"strings"

	chatapp "github.com/Dragonqos/lockwood-test/internal/resources/chat/app"
	chatdomain "github.com/Dragonqos/lockwood-test/internal/resources/chat/domain"
)

func (h *Handler) CreateAndJoinRoom(ctx context.Context, request CreateRoomRequest) (CreateRoomResponse, error) {
	if request.UserID == "" || strings.TrimSpace(request.Name) == "" {
		return CreateRoomResponse{Status: "bad_request"}, nil
	}

	oldRoom, err := h.coreApi.RoomForUser(ctx, request.UserID)
	if err != nil {
		return CreateRoomResponse{}, err
	}

	roomID := chatdomain.RoomID(request.Name)
	if err := h.coreApi.CreateAndJoinRoom(ctx, chatapp.CreateAndJoinCommand{
		Room:       chatdomain.Room{ID: roomID, Capacity: RoomCapacity},
		Membership: chatdomain.Membership{RoomID: roomID, UserID: request.UserID},
	}); err != nil {
		return CreateRoomResponse{Status: "room_exists", PreviousRoomID: oldRoom}, nil
	}

	events := make([]RoomEvent, 0, 2)
	if oldRoom != "" && oldRoom != roomID {
		events = append(events, RoomEvent{RoomID: oldRoom, Cmd: "user_left", User: request.UserID})
	}
	events = append(events, RoomEvent{RoomID: roomID, Cmd: "user_joined", User: request.UserID})
	return CreateRoomResponse{Status: "ok", RoomID: roomID, PreviousRoomID: oldRoom, Events: events}, nil
}
