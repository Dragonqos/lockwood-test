package domain

import "context"

type RoomID string

type Room struct {
	ID       RoomID
	Capacity int
}

type Membership struct {
	RoomID RoomID
	UserID string
}

type JoinStatus string

const (
	JoinOK       JoinStatus = "ok"
	RoomFull     JoinStatus = "room_full"
	RoomNotFound JoinStatus = "room_not_found"
)

//go:generate go tool mockery
type Repository interface {
	CreateAndJoin(ctx context.Context, room Room, membership Membership) error
	Join(ctx context.Context, membership Membership) (JoinStatus, error)
	Leave(ctx context.Context, membership Membership) error
	CurrentRoom(ctx context.Context, userID string) (RoomID, error)
	Members(ctx context.Context, roomID RoomID) ([]string, error)
}
