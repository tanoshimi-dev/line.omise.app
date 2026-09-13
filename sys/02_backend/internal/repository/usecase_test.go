package repository_test

import (
	"context"
	"testing"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/testutil"
)

func TestUsecaseListPublished_ExcludesDraft(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()
	repo := repository.NewUsecaseRepository(pool)

	if _, err := repo.Create(ctx, "published", "Client A", "Title", "", "published", "", "", nil); err != nil {
		t.Fatalf("Create published: %v", err)
	}
	if _, err := repo.Create(ctx, "draft", "Client B", "Title", "", "draft", "", "", nil); err != nil {
		t.Fatalf("Create draft: %v", err)
	}

	usecases, err := repo.ListPublished(ctx)
	if err != nil {
		t.Fatalf("ListPublished: %v", err)
	}
	if len(usecases) != 1 || usecases[0].Slug != "published" {
		t.Errorf("ListPublished = %+v, want only the published usecase", usecases)
	}
}

func TestUsecaseCreate_StoresThumbnailAndRelatedApp(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	repo := repository.NewUsecaseRepository(pool)

	usecase, err := repo.Create(context.Background(), "slug", "Client", "Title", "", "draft", "/images/x.png", "salon-reservation", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if usecase.ThumbnailURL != "/images/x.png" || usecase.RelatedDemoApp != "salon-reservation" {
		t.Errorf("usecase = %+v, want thumbnail/related_demo_app preserved", usecase)
	}
}

func TestUsecaseCreate_EmptyOptionalFieldsStoreAsEmpty(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	repo := repository.NewUsecaseRepository(pool)

	// Regression coverage: these columns are nullable, and nullIfEmpty
	// converts "" to SQL NULL (so the related_demo_app CHECK constraint
	// doesn't reject it) — GetByID must still read them back as "", not
	// error scanning NULL into a plain string.
	usecase, err := repo.Create(context.Background(), "slug", "Client", "Title", "", "draft", "", "", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	fetched, err := repo.GetByID(context.Background(), usecase.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if fetched.ThumbnailURL != "" || fetched.RelatedDemoApp != "" {
		t.Errorf("fetched = %+v, want empty strings", fetched)
	}
}
