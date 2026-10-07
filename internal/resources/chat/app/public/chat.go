package public

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (h *Handler) Chat(ctx context.Context, request ChatRequest) (ChatResponse, error) {
	if request.UserID == "" || strings.TrimSpace(request.Message) == "" {
		return ChatResponse{Status: "bad_request"}, nil
	}

	roomID, err := h.coreApi.RoomForUser(ctx, request.UserID)
	if err != nil {
		return ChatResponse{}, err
	}
	if roomID == "" {
		return ChatResponse{Status: "not_in_room"}, nil
	}

	messageID, err := uuid.NewV7()
	if err != nil {
		return ChatResponse{}, fmt.Errorf("create message id: %w", err)
	}

	return ChatResponse{
		Status:    "ok",
		MessageID: messageID.String(),
		RoomID:    roomID,
		Events:    []RoomEvent{{RoomID: roomID, Cmd: "chat", User: request.UserID, Message: request.Message, MessageID: messageID.String()}},
	}, nil
}
