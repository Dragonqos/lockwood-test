package public

import chatapp "github.com/Dragonqos/lockwood-test/internal/resources/chat/app"

type Handler struct {
	coreApi chatapp.ChatCoreAPI
}

func NewHandler(coreApi chatapp.ChatCoreAPI) *Handler {
	return &Handler{coreApi: coreApi}
}
