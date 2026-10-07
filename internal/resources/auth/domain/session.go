package domain

import (
	"context"
	"time"
)

type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
}

const SessionReplacedTopic = "session.replaced"

type SessionReplaced struct {
	UserID          string
	PreviousSession string
}

//go:generate go tool mockery
type SessionRepository interface {
	Claim(ctx context.Context, userID, sessionID string) error
	Current(ctx context.Context, userID string) (string, error)
	UserID(ctx context.Context, sessionID string) (string, error)
	Touch(ctx context.Context, sessionID string) error
	MarkDisconnected(ctx context.Context, sessionID string, expiresAt time.Time) error
	Expired(ctx context.Context, now time.Time) ([]Session, error)
	Release(ctx context.Context, userID, sessionID string) error
}
