package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/testutil"
)

func TestSessionCreateAndGetValid(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()

	user := testutil.CreateUser(t, pool, "reader")
	sessions := repository.NewSessionRepository(pool)

	sess, err := sessions.Create(ctx, user.ID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := sessions.GetValid(ctx, sess.ID)
	if err != nil {
		t.Fatalf("GetValid: %v", err)
	}
	if got.UserID != user.ID {
		t.Errorf("UserID = %d, want %d", got.UserID, user.ID)
	}
}

func TestSessionGetValid_NotFoundForUnknownID(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	sessions := repository.NewSessionRepository(pool)

	_, err := sessions.GetValid(context.Background(), "does-not-exist")
	if !errors.Is(err, repository.ErrSessionNotFound) {
		t.Errorf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestSessionDelete_RemovesSession(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()

	user := testutil.CreateUser(t, pool, "reader")
	sessions := repository.NewSessionRepository(pool)
	sess, err := sessions.Create(ctx, user.ID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := sessions.Delete(ctx, sess.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = sessions.GetValid(ctx, sess.ID)
	if !errors.Is(err, repository.ErrSessionNotFound) {
		t.Errorf("err after delete = %v, want ErrSessionNotFound", err)
	}
}

func TestSessionDelete_UnknownIDIsNotAnError(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	sessions := repository.NewSessionRepository(pool)

	if err := sessions.Delete(context.Background(), "does-not-exist"); err != nil {
		t.Errorf("Delete of unknown session should not error, got %v", err)
	}
}
