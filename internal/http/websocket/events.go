package websocket

import chatpublic "github.com/Dragonqos/lockwood-test/internal/resources/chat/app/public"

func (h *Handler) publishEvents(events []chatpublic.RoomEvent) {
	for _, event := range events {
		h.eventBus.Publish(string(event.RoomID), event)
	}
}
