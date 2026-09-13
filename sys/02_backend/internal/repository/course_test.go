package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/testutil"
)

func TestCourseListPublished_ExcludesDraft(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()
	repo := repository.NewCourseRepository(pool)

	if _, err := repo.Create(ctx, "published-course", "Published", "", 1, "published"); err != nil {
		t.Fatalf("Create published: %v", err)
	}
	if _, err := repo.Create(ctx, "draft-course", "Draft", "", 2, "draft"); err != nil {
		t.Fatalf("Create draft: %v", err)
	}

	courses, err := repo.ListPublished(ctx)
	if err != nil {
		t.Fatalf("ListPublished: %v", err)
	}
	if len(courses) != 1 || courses[0].Slug != "published-course" {
		t.Errorf("ListPublished = %+v, want only the published course", courses)
	}
}

func TestCourseListAll_IncludesDraft(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()
	repo := repository.NewCourseRepository(pool)

	if _, err := repo.Create(ctx, "draft-course", "Draft", "", 1, "draft"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	courses, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(courses) != 1 {
		t.Errorf("ListAll = %+v, want the draft course included", courses)
	}
}

func TestCourseGetPublishedBySlug_NotFoundForDraft(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()
	repo := repository.NewCourseRepository(pool)

	if _, err := repo.Create(ctx, "draft-course", "Draft", "", 1, "draft"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err := repo.GetPublishedBySlug(ctx, "draft-course")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound (draft course must not be publicly visible)", err)
	}
}

func TestCourseCreate_DuplicateSlugIsUniqueViolation(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()
	repo := repository.NewCourseRepository(pool)

	if _, err := repo.Create(ctx, "dup", "First", "", 1, "draft"); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if _, err := repo.Create(ctx, "dup", "Second", "", 2, "draft"); err == nil {
		t.Error("expected an error creating a course with a duplicate slug, got nil")
	}
}

func TestLessonListByCourse_PublishedOnlyFiltersDraftLessons(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()
	repo := repository.NewCourseRepository(pool)

	course, err := repo.Create(ctx, "course-1", "Course", "", 1, "published")
	if err != nil {
		t.Fatalf("Create course: %v", err)
	}
	if _, err := repo.CreateLesson(ctx, course.ID, "lesson-published", "Published Lesson", "", 1, "published"); err != nil {
		t.Fatalf("CreateLesson published: %v", err)
	}
	if _, err := repo.CreateLesson(ctx, course.ID, "lesson-draft", "Draft Lesson", "", 2, "draft"); err != nil {
		t.Fatalf("CreateLesson draft: %v", err)
	}

	publishedOnly, err := repo.ListLessonsByCourse(ctx, course.ID, true)
	if err != nil {
		t.Fatalf("ListLessonsByCourse(published only): %v", err)
	}
	if len(publishedOnly) != 1 || publishedOnly[0].Slug != "lesson-published" {
		t.Errorf("published-only lessons = %+v, want only lesson-published", publishedOnly)
	}

	all, err := repo.ListLessonsByCourse(ctx, course.ID, false)
	if err != nil {
		t.Fatalf("ListLessonsByCourse(all): %v", err)
	}
	if len(all) != 2 {
		t.Errorf("all lessons = %+v, want both lessons", all)
	}
}

func TestCourseDelete_NotFoundForUnknownID(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })

	repo := repository.NewCourseRepository(pool)
	err := repo.Delete(context.Background(), 999999)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
