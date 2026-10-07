package public

import chatdomain "github.com/Dragonqos/lockwood-test/internal/resources/chat/domain"

const RoomCapacity = 8

type CreateRoomRequest struct {
	UserID string
	Name   string
}

type CreateRoomResponse struct {
	Status         string
	RoomID         chatdomain.RoomID
	PreviousRoomID chatdomain.RoomID
	Events         []RoomEvent
}

type JoinRoomRequest struct {
	UserID string
	Name   string
}

type JoinRoomResponse struct {
	Status         string
	RoomID         chatdomain.RoomID
	PreviousRoomID chatdomain.RoomID
	Events         []RoomEvent
}

type LeaveRoomRequest struct {
	UserID string
}

type LeaveRoomResponse struct {
	Status string
	RoomID chatdomain.RoomID
	Events []RoomEvent
}

type ChatRequest struct {
	UserID  string
	Message string
}

type ChatResponse struct {
	Status    string
	MessageID string
	RoomID    chatdomain.RoomID
	Events    []RoomEvent
}

type RoomEvent struct {
	RoomID    chatdomain.RoomID `json:"-"`
	Cmd       string            `json:"cmd"`
	User      string            `json:"user,omitempty"`
	Message   string            `json:"message,omitempty"`
	MessageID string            `json:"message_id,omitempty"`
}
