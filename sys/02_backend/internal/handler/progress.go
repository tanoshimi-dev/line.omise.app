package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/middleware"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
)

// ProgressHandler implements dev-plan-06-exam-progress-api's progress
// endpoints. All routes require a logged-in Reader (dev-plan-06 6.3); the
// user id always comes from the session (middleware.CurrentUser), never from
// a request parameter, so one user can't read another's progress
// (dev-plan-06 6.4).
type ProgressHandler struct {
	Courses  *repository.CourseRepository
	Exams    *repository.ExamRepository
	Progress *repository.ProgressRepository
}

// CompleteLesson handles POST /api/lessons/:lessonId/complete.
func (h *ProgressHandler) CompleteLesson(c *gin.Context) {
	user := middleware.CurrentUser(c)
	lessonID, ok := parseIDParam(c, "lessonId")
	if !ok {
		return
	}

	lesson, err := h.Courses.GetLessonByID(c.Request.Context(), lessonID)
	if err != nil || lesson.Status != "published" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	progress, err := h.Progress.MarkLessonComplete(c.Request.Context(), user.ID, lessonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"lesson_id":    strconv.FormatInt(progress.LessonID, 10),
		"completed_at": progress.CompletedAt,
	})
}

// GetCourseProgress handles GET /api/courses/:slug/progress.
func (h *ProgressHandler) GetCourseProgress(c *gin.Context) {
	user := middleware.CurrentUser(c)
	ctx := c.Request.Context()

	course, err := h.Courses.GetPublishedBySlug(ctx, c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	lessons, err := h.Courses.ListLessonsByCourse(ctx, course.ID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	lessonIDs := make([]int64, len(lessons))
	for i, l := range lessons {
		lessonIDs[i] = l.ID
	}
	completed, err := h.Progress.CompletedLessons(ctx, user.ID, lessonIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	lessonItems := make([]gin.H, 0, len(lessons))
	completedCount := 0
	for _, l := range lessons {
		completedAt, isCompleted := completed[l.ID]
		if isCompleted {
			completedCount++
		}

		item := gin.H{
			"id":    strconv.FormatInt(l.ID, 10),
			"slug":  l.Slug,
			"title": l.Title,
		}
		if isCompleted {
			item["completed"] = true
			item["completed_at"] = completedAt
		} else {
			item["completed"] = false
			item["completed_at"] = nil
		}
		item["exam"] = h.lessonExamProgress(ctx, user.ID, l.ID)
		lessonItems = append(lessonItems, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"course": gin.H{
			"id":    strconv.FormatInt(course.ID, 10),
			"slug":  course.Slug,
			"title": course.Title,
		},
		"completed_count": completedCount,
		"total_count":     len(lessons),
		"lessons":         lessonItems,
	})
}

// GetMyProgress handles GET /api/me/progress — a summary across every
// published course (dev-plan-06 6.3, optional "マイページ用" endpoint).
func (h *ProgressHandler) GetMyProgress(c *gin.Context) {
	user := middleware.CurrentUser(c)
	ctx := c.Request.Context()

	courses, err := h.Courses.ListPublished(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	items := make([]gin.H, 0, len(courses))
	for _, course := range courses {
		lessons, err := h.Courses.ListLessonsByCourse(ctx, course.ID, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		completedCount, err := h.Progress.CountCompletedByCourse(ctx, user.ID, course.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		items = append(items, gin.H{
			"course_id":       strconv.FormatInt(course.ID, 10),
			"slug":            course.Slug,
			"title":           course.Title,
			"completed_count": completedCount,
			"total_count":     len(lessons),
		})
	}

	c.JSON(http.StatusOK, gin.H{"courses": items})
}

// lessonExamProgress returns the exam block for a course-progress lesson
// entry, or nil if the lesson has no exam.
func (h *ProgressHandler) lessonExamProgress(ctx context.Context, userID, lessonID int64) gin.H {
	exam, err := h.Exams.GetByLessonID(ctx, lessonID)
	if err != nil {
		return nil
	}

	result := gin.H{
		"id":            strconv.FormatInt(exam.ID, 10),
		"title":         exam.Title,
		"passing_score": exam.PassingScore,
		"latest_result": nil,
	}
	latest, err := h.Progress.LatestExamResult(ctx, userID, exam.ID)
	if err == nil {
		result["latest_result"] = gin.H{
			"score":        latest.Score,
			"passed":       latest.Passed,
			"submitted_at": latest.SubmittedAt,
		}
	}
	return result
}
