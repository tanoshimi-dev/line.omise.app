package repository_test

import (
	"context"
	"testing"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/testutil"
)

func TestArticleListPublished_FiltersByTag(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()
	repo := repository.NewArticleRepository(pool)

	tagged, err := repo.Create(ctx, "line-operation", "tagged", "Tagged", "", "published", nil)
	if err != nil {
		t.Fatalf("Create tagged: %v", err)
	}
	if _, err := repo.Create(ctx, "line-operation", "untagged", "Untagged", "", "published", nil); err != nil {
		t.Fatalf("Create untagged: %v", err)
	}

	tag, err := repo.GetOrCreateTag(ctx, "Rich Menu", "rich-menu")
	if err != nil {
		t.Fatalf("GetOrCreateTag: %v", err)
	}
	if err := repo.AttachTag(ctx, tagged.ID, tag.ID); err != nil {
		t.Fatalf("AttachTag: %v", err)
	}

	// Filtering by the tag should hide the untagged article, but the tag
	// itself should still be discoverable from the unfiltered list — see
	// dev-plan-09-frontend-learn's ArticleListView design.
	filtered, err := repo.ListPublished(ctx, "line-operation", "rich-menu")
	if err != nil {
		t.Fatalf("ListPublished(tag filter): %v", err)
	}
	if len(filtered) != 1 || filtered[0].Slug != "tagged" {
		t.Errorf("tag-filtered list = %+v, want only the tagged article", filtered)
	}

	all, err := repo.ListPublished(ctx, "line-operation", "")
	if err != nil {
		t.Fatalf("ListPublished(no filter): %v", err)
	}
	if len(all) != 2 {
		t.Errorf("unfiltered list = %+v, want both articles", all)
	}
}

func TestArticleCreate_DefaultsPublishedAtWhenPublished(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	repo := repository.NewArticleRepository(pool)

	article, err := repo.Create(context.Background(), "ai", "auto-publish", "Title", "", "published", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if article.PublishedAt == nil {
		t.Error("PublishedAt = nil, want auto-set to now() when creating as published")
	}
}

func TestArticleCreate_NoPublishedAtWhenDraft(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	repo := repository.NewArticleRepository(pool)

	article, err := repo.Create(context.Background(), "ai", "still-draft", "Title", "", "draft", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if article.PublishedAt != nil {
		t.Error("PublishedAt should stay nil for a draft article")
	}
}

func TestArticleUpdate_PreservesPublishedAtAcrossEdits(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()
	repo := repository.NewArticleRepository(pool)

	created, err := repo.Create(ctx, "ai", "slug", "Title", "", "published", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	firstPublishedAt := created.PublishedAt

	updated, err := repo.Update(ctx, created.ID, "ai", "slug", "New Title", "body", "published", nil)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.PublishedAt == nil || !updated.PublishedAt.Equal(*firstPublishedAt) {
		t.Errorf("PublishedAt changed on edit: got %v, want unchanged %v", updated.PublishedAt, firstPublishedAt)
	}
}

func TestAttachTag_IsIdempotent(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()
	repo := repository.NewArticleRepository(pool)

	article, err := repo.Create(ctx, "ai", "slug", "Title", "", "published", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	tag, err := repo.GetOrCreateTag(ctx, "Tag", "tag")
	if err != nil {
		t.Fatalf("GetOrCreateTag: %v", err)
	}

	if err := repo.AttachTag(ctx, article.ID, tag.ID); err != nil {
		t.Fatalf("first AttachTag: %v", err)
	}
	if err := repo.AttachTag(ctx, article.ID, tag.ID); err != nil {
		t.Fatalf("second AttachTag (should be a no-op, not an error): %v", err)
	}

	tags, err := repo.ListTagsForArticle(ctx, article.ID)
	if err != nil {
		t.Fatalf("ListTagsForArticle: %v", err)
	}
	if len(tags) != 1 {
		t.Errorf("tags = %+v, want exactly one (attach is idempotent)", tags)
	}
}
