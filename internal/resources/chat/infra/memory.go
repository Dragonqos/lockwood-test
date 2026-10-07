package infra

import (
	"context"
	"errors"
	"sync"

	"github.com/Dragonqos/lockwood-test/internal/resources/chat/domain"
)

var ErrRoomExists = errors.New("room already exists")

type memoryRoom struct {
	domain.Room
	members map[string]struct{}
}

type MemoryRepository struct {
	mu       sync.RWMutex
	rooms    map[domain.RoomID]*memoryRoom
	userRoom map[string]domain.RoomID
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{rooms: make(map[domain.RoomID]*memoryRoom), userRoom: make(map[string]domain.RoomID)}
}

func (r *MemoryRepository) CreateAndJoin(_ context.Context, room domain.Room, membership domain.Membership) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.rooms[room.ID]; exists {
		return ErrRoomExists
	}
	if previous := r.userRoom[membership.UserID]; previous != "" {
		r.leaveLocked(membership.UserID, previous)
	}
	r.rooms[room.ID] = &memoryRoom{Room: room, members: map[string]struct{}{membership.UserID: {}}}
	r.userRoom[membership.UserID] = room.ID
	return nil
}

func (r *MemoryRepository) Join(_ context.Context, membership domain.Membership) (domain.JoinStatus, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	room, exists := r.rooms[membership.RoomID]
	if !exists {
		return domain.RoomNotFound, nil
	}
	if _, already := room.members[membership.UserID]; already {
		return domain.JoinOK, nil
	}
	if len(room.members) >= room.Capacity {
		return domain.RoomFull, nil
	}
	if previous := r.userRoom[membership.UserID]; previous != "" {
		r.leaveLocked(membership.UserID, previous)
	}
	room.members[membership.UserID] = struct{}{}
	r.userRoom[membership.UserID] = membership.RoomID
	return domain.JoinOK, nil
}

func (r *MemoryRepository) Leave(_ context.Context, membership domain.Membership) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.leaveLocked(membership.UserID, membership.RoomID)
	return nil
}

func (r *MemoryRepository) CurrentRoom(_ context.Context, userID string) (domain.RoomID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.userRoom[userID], nil
}

func (r *MemoryRepository) Members(_ context.Context, roomID domain.RoomID) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	room := r.rooms[roomID]
	if room == nil {
		return nil, nil
	}
	result := make([]string, 0, len(room.members))
	for userID := range room.members {
		result = append(result, userID)
	}
	return result, nil
}

func (r *MemoryRepository) leaveLocked(userID string, roomID domain.RoomID) {
	room := r.rooms[roomID]
	if room == nil {
		return
	}
	delete(room.members, userID)
	if r.userRoom[userID] == roomID {
		delete(r.userRoom, userID)
	}
	if len(room.members) == 0 {
		delete(r.rooms, roomID)
	}
}
