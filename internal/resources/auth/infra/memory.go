package infra

import (
	"context"
	"sync"
	"time"

	"github.com/Dragonqos/lockwood-test/internal/resources/auth/domain"
)

type MemorySessionRepository struct {
	mu        sync.RWMutex
	sessions  map[string]string
	usersByID map[string]string
	expiresAt map[string]time.Time
}

func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{
		sessions:  make(map[string]string),
		usersByID: make(map[string]string),
		expiresAt: make(map[string]time.Time),
	}
}

func (r *MemorySessionRepository) Claim(_ context.Context, userID, sessionID string) error {
	r.mu.Lock() // exclusive lock
	defer r.mu.Unlock()

	if previous := r.sessions[userID]; previous != "" {
		delete(r.usersByID, previous)
		delete(r.expiresAt, previous)
	}
	if previousUser := r.usersByID[sessionID]; previousUser != "" && previousUser != userID {
		if r.sessions[previousUser] == sessionID {
			delete(r.sessions, previousUser)
		}
	}
	r.sessions[userID] = sessionID
	r.usersByID[sessionID] = userID
	delete(r.expiresAt, sessionID)

	return nil
}

func (r *MemorySessionRepository) Current(_ context.Context, userID string) (string, error) {
	r.mu.RLock() // concurrent read
	defer r.mu.RUnlock()

	return r.sessions[userID], nil
}

func (r *MemorySessionRepository) UserID(_ context.Context, sessionID string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.usersByID[sessionID], nil
}

func (r *MemorySessionRepository) Touch(_ context.Context, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.expiresAt, sessionID)
	return nil
}

func (r *MemorySessionRepository) MarkDisconnected(_ context.Context, sessionID string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.usersByID[sessionID]; exists {
		r.expiresAt[sessionID] = expiresAt
	}
	return nil
}

func (r *MemorySessionRepository) Expired(_ context.Context, now time.Time) ([]domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.Session, 0)
	for sessionID, expiresAt := range r.expiresAt {
		if !expiresAt.After(now) {
			result = append(result, domain.Session{
				ID:        sessionID,
				UserID:    r.usersByID[sessionID],
				ExpiresAt: expiresAt,
			})
		}
	}
	return result, nil
}

func (r *MemorySessionRepository) Release(_ context.Context, userID, sessionID string) error {
	r.mu.Lock() // exclusive lock
	defer r.mu.Unlock()

	if r.sessions[userID] == sessionID {
		delete(r.sessions, userID)
		delete(r.usersByID, sessionID)
		delete(r.expiresAt, sessionID)
	}

	return nil
}
