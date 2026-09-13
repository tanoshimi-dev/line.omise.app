package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Course mirrors the `courses` table (dev-plan-02-database 2.3).
type Course struct {
	ID          int64
	Slug        string
	Title       string
	Description string
	SortOrder   int
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Lesson mirrors the `lessons` table.
type Lesson struct {
	ID        int64
	CourseID  int64
	Slug      string
	Title     string
	Body      string
	SortOrder int
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

const coursePublished = "published"

// CourseRepository queries `courses` and `lessons`.
type CourseRepository struct {
	db *pgxpool.Pool
}

func NewCourseRepository(db *pgxpool.Pool) *CourseRepository {
	return &CourseRepository{db: db}
}

// ListPublished returns published courses ordered for display (dev-plan-05
// 5.1: draft courses never appear in the public API).
func (r *CourseRepository) ListPublished(ctx context.Context) ([]Course, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, title, COALESCE(description, ''), sort_order, status, created_at, updated_at
		FROM courses WHERE status = $1
		ORDER BY sort_order, id
	`, coursePublished)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCourses(rows)
}

// ListAll returns every course regardless of status, for the admin UI
// (dev-plan-11-frontend-admin) which needs to see and manage drafts too.
func (r *CourseRepository) ListAll(ctx context.Context) ([]Course, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, title, COALESCE(description, ''), sort_order, status, created_at, updated_at
		FROM courses
		ORDER BY sort_order, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCourses(rows)
}

// GetPublishedBySlug returns a published course, or ErrNotFound if it
// doesn't exist or isn't published.
func (r *CourseRepository) GetPublishedBySlug(ctx context.Context, slug string) (*Course, error) {
	return r.getBySlug(ctx, slug, true)
}

// GetBySlug returns a course regardless of status (admin use).
func (r *CourseRepository) GetBySlug(ctx context.Context, slug string) (*Course, error) {
	return r.getBySlug(ctx, slug, false)
}

func (r *CourseRepository) getBySlug(ctx context.Context, slug string, publishedOnly bool) (*Course, error) {
	query := `SELECT id, slug, title, COALESCE(description, ''), sort_order, status, created_at, updated_at
		FROM courses WHERE slug = $1`
	args := []any{slug}
	if publishedOnly {
		query += " AND status = $2"
		args = append(args, coursePublished)
	}
	row := r.db.QueryRow(ctx, query, args...)
	return scanCourse(row)
}

func (r *CourseRepository) GetByID(ctx context.Context, id int64) (*Course, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, slug, title, COALESCE(description, ''), sort_order, status, created_at, updated_at
		FROM courses WHERE id = $1
	`, id)
	return scanCourse(row)
}

func (r *CourseRepository) Create(ctx context.Context, slug, title, description string, sortOrder int, status string) (*Course, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO courses (slug, title, description, sort_order, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, slug, title, COALESCE(description, ''), sort_order, status, created_at, updated_at
	`, slug, title, description, sortOrder, status)
	return scanCourse(row)
}

func (r *CourseRepository) Update(ctx context.Context, id int64, slug, title, description string, sortOrder int, status string) (*Course, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE courses SET slug = $2, title = $3, description = $4, sort_order = $5, status = $6, updated_at = now()
		WHERE id = $1
		RETURNING id, slug, title, COALESCE(description, ''), sort_order, status, created_at, updated_at
	`, id, slug, title, description, sortOrder, status)
	return scanCourse(row)
}

func (r *CourseRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM courses WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListLessonsByCourse returns lessons for a course. publishedOnly restricts
// to status='published' for the public API; admin callers pass false.
func (r *CourseRepository) ListLessonsByCourse(ctx context.Context, courseID int64, publishedOnly bool) ([]Lesson, error) {
	query := `SELECT id, course_id, slug, title, COALESCE(body, ''), sort_order, status, created_at, updated_at
		FROM lessons WHERE course_id = $1`
	args := []any{courseID}
	if publishedOnly {
		query += " AND status = $2"
		args = append(args, coursePublished)
	}
	query += " ORDER BY sort_order, id"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lessons []Lesson
	for rows.Next() {
		var l Lesson
		if err := rows.Scan(&l.ID, &l.CourseID, &l.Slug, &l.Title, &l.Body, &l.SortOrder, &l.Status, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}
	return lessons, rows.Err()
}

// GetLessonBySlug returns a lesson within a course by slug. publishedOnly
// also requires the lesson itself to be published (the caller is expected to
// have already confirmed the parent course is published/visible).
func (r *CourseRepository) GetLessonBySlug(ctx context.Context, courseID int64, lessonSlug string, publishedOnly bool) (*Lesson, error) {
	query := `SELECT id, course_id, slug, title, COALESCE(body, ''), sort_order, status, created_at, updated_at
		FROM lessons WHERE course_id = $1 AND slug = $2`
	args := []any{courseID, lessonSlug}
	if publishedOnly {
		query += " AND status = $3"
		args = append(args, coursePublished)
	}
	return scanLesson(r.db.QueryRow(ctx, query, args...))
}

func (r *CourseRepository) GetLessonByID(ctx context.Context, id int64) (*Lesson, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, course_id, slug, title, COALESCE(body, ''), sort_order, status, created_at, updated_at
		FROM lessons WHERE id = $1
	`, id)
	return scanLesson(row)
}

func (r *CourseRepository) CreateLesson(ctx context.Context, courseID int64, slug, title, body string, sortOrder int, status string) (*Lesson, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO lessons (course_id, slug, title, body, sort_order, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, course_id, slug, title, COALESCE(body, ''), sort_order, status, created_at, updated_at
	`, courseID, slug, title, body, sortOrder, status)
	return scanLesson(row)
}

func (r *CourseRepository) UpdateLesson(ctx context.Context, id int64, slug, title, body string, sortOrder int, status string) (*Lesson, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE lessons SET slug = $2, title = $3, body = $4, sort_order = $5, status = $6, updated_at = now()
		WHERE id = $1
		RETURNING id, course_id, slug, title, COALESCE(body, ''), sort_order, status, created_at, updated_at
	`, id, slug, title, body, sortOrder, status)
	return scanLesson(row)
}

func (r *CourseRepository) DeleteLesson(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM lessons WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanCourse(row rowScanner) (*Course, error) {
	var c Course
	err := row.Scan(&c.ID, &c.Slug, &c.Title, &c.Description, &c.SortOrder, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &c, nil
}

func scanCourses(rows pgx.Rows) ([]Course, error) {
	var courses []Course
	for rows.Next() {
		c, err := scanCourse(rows)
		if err != nil {
			return nil, err
		}
		courses = append(courses, *c)
	}
	return courses, rows.Err()
}

func scanLesson(row rowScanner) (*Lesson, error) {
	var l Lesson
	err := row.Scan(&l.ID, &l.CourseID, &l.Slug, &l.Title, &l.Body, &l.SortOrder, &l.Status, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &l, nil
}
