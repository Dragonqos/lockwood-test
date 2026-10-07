package app

import (
	"context"

	"github.com/Dragonqos/lockwood-test/internal/resources/chat/domain"
)

//go:generate go tool mockery
type ChatCoreAPI interface {
	CreateAndJoinRoom(ctx context.Context, command CreateAndJoinCommand) error
	JoinRoom(ctx context.Context, command JoinCommand) (domain.JoinStatus, error)
	LeaveRoom(ctx context.Context, command LeaveCommand) error
	RoomForUser(ctx context.Context, userID string) (domain.RoomID, error)
	ListMembers(ctx context.Context, roomID domain.RoomID) ([]string, error)
}

type CreateAndJoinCommand struct {
	Room       domain.Room
	Membership domain.Membership
}

type JoinCommand struct {
	Membership domain.Membership
}

type LeaveCommand struct {
	Membership domain.Membership
}

type ChatCore struct {
	repository domain.Repository
}

func NewChatCore(repository domain.Repository) *ChatCore {
	return &ChatCore{repository: repository}
}

func (s *ChatCore) CreateAndJoinRoom(ctx context.Context, command CreateAndJoinCommand) error {
	return s.repository.CreateAndJoin(ctx, command.Room, command.Membership)
}

func (s *ChatCore) JoinRoom(ctx context.Context, command JoinCommand) (domain.JoinStatus, error) {
	return s.repository.Join(ctx, command.Membership)
}

func (s *ChatCore) LeaveRoom(ctx context.Context, command LeaveCommand) error {
	return s.repository.Leave(ctx, command.Membership)
}

func (s *ChatCore) RoomForUser(ctx context.Context, userID string) (domain.RoomID, error) {
	return s.repository.CurrentRoom(ctx, userID)
}

func (s *ChatCore) ListMembers(ctx context.Context, roomID domain.RoomID) ([]string, error) {
	return s.repository.Members(ctx, roomID)
}
