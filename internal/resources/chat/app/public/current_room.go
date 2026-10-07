package public

import (
	"context"

	chatdomain "github.com/Dragonqos/lockwood-test/internal/resources/chat/domain"
)

func (h *Handler) CurrentRoom(ctx context.Context, userID string) (chatdomain.RoomID, error) {
	return h.coreApi.RoomForUser(ctx, userID)
}
