package repository_test

import (
	"context"
	"testing"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/testutil"
)

func TestUpsertByProvider_CreatesReaderByDefault(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	repo := repository.NewUserRepository(pool)

	user, err := repo.UpsertByProvider(context.Background(), "google", "u1", "u1@example.com", "User One", "", false)
	if err != nil {
		t.Fatalf("UpsertByProvider: %v", err)
	}
	if user.Role != "reader" {
		t.Errorf("Role = %q, want reader", user.Role)
	}
}

func TestUpsertByProvider_PromotesToAdminWhenEmailMatches(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	repo := repository.NewUserRepository(pool)

	user, err := repo.UpsertByProvider(context.Background(), "google", "u2", "admin@example.com", "Admin", "", true)
	if err != nil {
		t.Fatalf("UpsertByProvider: %v", err)
	}
	if user.Role != "admin" {
		t.Errorf("Role = %q, want admin", user.Role)
	}
}

// dev-plan-04-auth 4.5: removing an address from ADMIN_EMAILS must not
// silently revoke access granted some other way — login should never demote.
func TestUpsertByProvider_DoesNotDemoteOnSubsequentLogin(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	repo := repository.NewUserRepository(pool)
	ctx := context.Background()

	first, err := repo.UpsertByProvider(ctx, "google", "u3", "u3@example.com", "User Three", "", true)
	if err != nil {
		t.Fatalf("first UpsertByProvider: %v", err)
	}
	if first.Role != "admin" {
		t.Fatalf("precondition failed: first login role = %q, want admin", first.Role)
	}

	second, err := repo.UpsertByProvider(ctx, "google", "u3", "u3@example.com", "User Three", "", false)
	if err != nil {
		t.Fatalf("second UpsertByProvider: %v", err)
	}
	if second.Role != "admin" {
		t.Errorf("Role after re-login without admin match = %q, want admin (must not demote)", second.Role)
	}
}

func TestUpsertByProvider_RefreshesProfileFields(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	repo := repository.NewUserRepository(pool)
	ctx := context.Background()

	_, err := repo.UpsertByProvider(ctx, "google", "u4", "old@example.com", "Old Name", "", false)
	if err != nil {
		t.Fatalf("first UpsertByProvider: %v", err)
	}

	updated, err := repo.UpsertByProvider(ctx, "google", "u4", "new@example.com", "New Name", "https://example.com/avatar.png", false)
	if err != nil {
		t.Fatalf("second UpsertByProvider: %v", err)
	}
	if updated.Email != "new@example.com" || updated.DisplayName != "New Name" || updated.AvatarURL != "https://example.com/avatar.png" {
		t.Errorf("profile fields not refreshed: %+v", updated)
	}
}

func TestUserGetByID_HandlesNullableColumns(t *testing.T) {
	// Regression test for the bug found in dev-plan-04-auth: scanning a row
	// with NULL email/display_name/avatar_url into plain strings used to
	// panic with "cannot scan NULL into *string" before COALESCE was added.
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })

	var id int64
	err := pool.QueryRow(context.Background(), `
		INSERT INTO users (provider, provider_user_id, role) VALUES ('line', 'u5', 'reader') RETURNING id
	`).Scan(&id)
	if err != nil {
		t.Fatalf("failed to insert user with NULL columns: %v", err)
	}

	repo := repository.NewUserRepository(pool)
	user, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if user.Email != "" || user.DisplayName != "" || user.AvatarURL != "" {
		t.Errorf("expected empty strings for NULL columns, got %+v", user)
	}
}
