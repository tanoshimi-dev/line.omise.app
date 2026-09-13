package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/service"
)

// CourseHandler implements dev-plan-05-content-api's course/lesson endpoints.
type CourseHandler struct {
	Courses *repository.CourseRepository
}

type courseRequest struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	Status      string `json:"status"`
}

type lessonRequest struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	SortOrder int    `json:"sort_order"`
	Status    string `json:"status"`
}

// ListCourses handles GET /api/courses — published only.
func (h *CourseHandler) ListCourses(c *gin.Context) {
	courses, err := h.Courses.ListPublished(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list courses"})
		return
	}
	items := make([]gin.H, 0, len(courses))
	for i := range courses {
		items = append(items, courseJSON(&courses[i], nil))
	}
	c.JSON(http.StatusOK, gin.H{"courses": items})
}

// GetCourse handles GET /api/courses/:slug — published course with its
// published lessons embedded.
func (h *CourseHandler) GetCourse(c *gin.Context) {
	course, err := h.Courses.GetPublishedBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		respondCourseError(c, err)
		return
	}
	lessons, err := h.Courses.ListLessonsByCourse(c.Request.Context(), course.ID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list lessons"})
		return
	}
	c.JSON(http.StatusOK, courseJSON(course, lessons))
}

// GetLesson handles GET /api/courses/:slug/lessons/:lessonSlug.
func (h *CourseHandler) GetLesson(c *gin.Context) {
	course, err := h.Courses.GetPublishedBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		respondCourseError(c, err)
		return
	}
	lesson, err := h.Courses.GetLessonBySlug(c.Request.Context(), course.ID, c.Param("lessonSlug"), true)
	if err != nil {
		respondCourseError(c, err)
		return
	}
	c.JSON(http.StatusOK, lessonJSON(lesson))
}

// AdminListCourses handles GET /api/admin/courses — all statuses, for the
// admin UI's course list (dev-plan-11-frontend-admin 11.2).
func (h *CourseHandler) AdminListCourses(c *gin.Context) {
	courses, err := h.Courses.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list courses"})
		return
	}
	items := make([]gin.H, 0, len(courses))
	for i := range courses {
		items = append(items, courseJSON(&courses[i], nil))
	}
	c.JSON(http.StatusOK, gin.H{"courses": items})
}

// AdminGetCourse handles GET /api/admin/courses/:id — any status, with all
// of its lessons (any status) embedded, for prefilling the edit form.
func (h *CourseHandler) AdminGetCourse(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	course, err := h.Courses.GetByID(c.Request.Context(), id)
	if err != nil {
		respondCourseError(c, err)
		return
	}
	lessons, err := h.Courses.ListLessonsByCourse(c.Request.Context(), course.ID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list lessons"})
		return
	}
	c.JSON(http.StatusOK, courseJSON(course, lessons))
}

// AdminGetLesson handles GET /api/admin/lessons/:id — any status, for
// prefilling the lesson edit form.
func (h *CourseHandler) AdminGetLesson(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	lesson, err := h.Courses.GetLessonByID(c.Request.Context(), id)
	if err != nil {
		respondCourseError(c, err)
		return
	}
	c.JSON(http.StatusOK, lessonJSON(lesson))
}

// AdminCreateCourse handles POST /api/admin/courses.
func (h *CourseHandler) AdminCreateCourse(c *gin.Context) {
	var req courseRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Slug == "" || req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and title are required"})
		return
	}
	if err := service.ValidateStatus(req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	course, err := h.Courses.Create(c.Request.Context(), req.Slug, req.Title, req.Description, req.SortOrder, req.Status)
	if err != nil {
		respondCourseWriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, courseJSON(course, nil))
}

// AdminUpdateCourse handles PUT /api/admin/courses/:id.
func (h *CourseHandler) AdminUpdateCourse(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req courseRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Slug == "" || req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and title are required"})
		return
	}
	if err := service.ValidateStatus(req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	course, err := h.Courses.Update(c.Request.Context(), id, req.Slug, req.Title, req.Description, req.SortOrder, req.Status)
	if err != nil {
		respondCourseWriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, courseJSON(course, nil))
}

// AdminDeleteCourse handles DELETE /api/admin/courses/:id.
func (h *CourseHandler) AdminDeleteCourse(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.Courses.Delete(c.Request.Context(), id); err != nil {
		respondCourseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// AdminCreateLesson handles POST /api/admin/courses/:id/lessons.
func (h *CourseHandler) AdminCreateLesson(c *gin.Context) {
	courseID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req lessonRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Slug == "" || req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and title are required"})
		return
	}
	if err := service.ValidateStatus(req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lesson, err := h.Courses.CreateLesson(c.Request.Context(), courseID, req.Slug, req.Title, req.Body, req.SortOrder, req.Status)
	if err != nil {
		respondCourseWriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, lessonJSON(lesson))
}

// AdminUpdateLesson handles PUT /api/admin/lessons/:id.
func (h *CourseHandler) AdminUpdateLesson(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req lessonRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Slug == "" || req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and title are required"})
		return
	}
	if err := service.ValidateStatus(req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lesson, err := h.Courses.UpdateLesson(c.Request.Context(), id, req.Slug, req.Title, req.Body, req.SortOrder, req.Status)
	if err != nil {
		respondCourseWriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, lessonJSON(lesson))
}

// AdminDeleteLesson handles DELETE /api/admin/lessons/:id.
func (h *CourseHandler) AdminDeleteLesson(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.Courses.DeleteLesson(c.Request.Context(), id); err != nil {
		respondCourseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func respondCourseError(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}

func respondCourseWriteError(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if errors.Is(service.AsDuplicateSlug(err), service.ErrDuplicateSlug) {
		c.JSON(http.StatusConflict, gin.H{"error": "slug already in use"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}

func courseJSON(course *repository.Course, lessons []repository.Lesson) gin.H {
	h := gin.H{
		"id":          strconv.FormatInt(course.ID, 10),
		"slug":        course.Slug,
		"title":       course.Title,
		"description": course.Description,
		"sort_order":  course.SortOrder,
		"status":      course.Status,
		"created_at":  course.CreatedAt,
		"updated_at":  course.UpdatedAt,
	}
	if lessons != nil {
		items := make([]gin.H, 0, len(lessons))
		for i := range lessons {
			items = append(items, lessonJSON(&lessons[i]))
		}
		h["lessons"] = items
	}
	return h
}

func lessonJSON(lesson *repository.Lesson) gin.H {
	return gin.H{
		"id":         strconv.FormatInt(lesson.ID, 10),
		"course_id":  strconv.FormatInt(lesson.CourseID, 10),
		"slug":       lesson.Slug,
		"title":      lesson.Title,
		"body":       lesson.Body,
		"sort_order": lesson.SortOrder,
		"status":     lesson.Status,
		"created_at": lesson.CreatedAt,
		"updated_at": lesson.UpdatedAt,
	}
}
