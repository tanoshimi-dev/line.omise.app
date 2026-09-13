package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/service"
)

// UsecaseHandler implements dev-plan-05-content-api's usecase endpoints.
type UsecaseHandler struct {
	Usecases *repository.UsecaseRepository
}

type usecaseRequest struct {
	Slug       string `json:"slug"`
	ClientName string `json:"client_name"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Status     string `json:"status"`
}

// ListUsecases handles GET /api/usecases.
func (h *UsecaseHandler) ListUsecases(c *gin.Context) {
	usecases, err := h.Usecases.ListPublished(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list usecases"})
		return
	}
	items := make([]gin.H, 0, len(usecases))
	for i := range usecases {
		items = append(items, usecaseJSON(&usecases[i]))
	}
	c.JSON(http.StatusOK, gin.H{"usecases": items})
}

// GetUsecase handles GET /api/usecases/:slug.
func (h *UsecaseHandler) GetUsecase(c *gin.Context) {
	usecase, err := h.Usecases.GetPublishedBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		respondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, usecaseJSON(usecase))
}

// AdminCreateUsecase handles POST /api/admin/usecases.
func (h *UsecaseHandler) AdminCreateUsecase(c *gin.Context) {
	var req usecaseRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Slug == "" || req.Title == "" || req.ClientName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug, client_name and title are required"})
		return
	}
	if err := service.ValidateStatus(req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usecase, err := h.Usecases.Create(c.Request.Context(), req.Slug, req.ClientName, req.Title, req.Body, req.Status, nil)
	if err != nil {
		respondUsecaseWriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, usecaseJSON(usecase))
}

// AdminUpdateUsecase handles PUT /api/admin/usecases/:id.
func (h *UsecaseHandler) AdminUpdateUsecase(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req usecaseRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Slug == "" || req.Title == "" || req.ClientName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug, client_name and title are required"})
		return
	}
	if err := service.ValidateStatus(req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usecase, err := h.Usecases.Update(c.Request.Context(), id, req.Slug, req.ClientName, req.Title, req.Body, req.Status, nil)
	if err != nil {
		respondUsecaseWriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, usecaseJSON(usecase))
}

// AdminDeleteUsecase handles DELETE /api/admin/usecases/:id.
func (h *UsecaseHandler) AdminDeleteUsecase(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.Usecases.Delete(c.Request.Context(), id); err != nil {
		respondUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func respondUsecaseError(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}

func respondUsecaseWriteError(c *gin.Context, err error) {
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

func usecaseJSON(usecase *repository.Usecase) gin.H {
	return gin.H{
		"id":           strconv.FormatInt(usecase.ID, 10),
		"slug":         usecase.Slug,
		"client_name":  usecase.ClientName,
		"title":        usecase.Title,
		"body":         usecase.Body,
		"status":       usecase.Status,
		"published_at": usecase.PublishedAt,
		"created_at":   usecase.CreatedAt,
		"updated_at":   usecase.UpdatedAt,
	}
}
