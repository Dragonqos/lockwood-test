package infra

import (
	"context"
	"testing"
)

func TestMemorySessionRepositoryClaimReplacesUserBoundToSession(t *testing.T) {
	repository := NewMemorySessionRepository()
	ctx := context.Background()

	if err := repository.Claim(ctx, "user-a", "session-1"); err != nil {
		t.Fatal(err)
	}
	if err := repository.Claim(ctx, "user-b", "session-1"); err != nil {
		t.Fatal(err)
	}

	if current, _ := repository.Current(ctx, "user-a"); current != "" {
		t.Fatalf("expected user-a session to be released, got %q", current)
	}
	if current, _ := repository.Current(ctx, "user-b"); current != "session-1" {
		t.Fatalf("expected user-b to own session-1, got %q", current)
	}
	if userID, _ := repository.UserID(ctx, "session-1"); userID != "user-b" {
		t.Fatalf("expected session-1 to resolve to user-b, got %q", userID)
	}
}
