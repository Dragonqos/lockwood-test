package public

import (
	"context"

	chatapp "github.com/Dragonqos/lockwood-test/internal/resources/chat/app"
	chatdomain "github.com/Dragonqos/lockwood-test/internal/resources/chat/domain"
)

func (h *Handler) LeaveRoom(ctx context.Context, request LeaveRoomRequest) (LeaveRoomResponse, error) {
	if request.UserID == "" {
		return LeaveRoomResponse{Status: "bad_request"}, nil
	}

	roomID, err := h.coreApi.RoomForUser(ctx, request.UserID)
	if err != nil {
		return LeaveRoomResponse{}, err
	}
	if roomID == "" {
		return LeaveRoomResponse{Status: "ok"}, nil
	}
	if err := h.coreApi.LeaveRoom(ctx, chatapp.LeaveCommand{
		Membership: chatdomain.Membership{RoomID: roomID, UserID: request.UserID},
	}); err != nil {
		return LeaveRoomResponse{}, err
	}
	return LeaveRoomResponse{
		Status: "ok",
		RoomID: roomID,
		Events: []RoomEvent{{RoomID: roomID, Cmd: "user_left", User: request.UserID}},
	}, nil
}
