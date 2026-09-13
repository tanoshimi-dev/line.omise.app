package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/service"
)

// ArticleHandler implements dev-plan-05-content-api's article/tag endpoints.
type ArticleHandler struct {
	Articles *repository.ArticleRepository
}

type articleRequest struct {
	Category string `json:"category"`
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Status   string `json:"status"`
}

type attachTagRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ListArticles handles GET /api/articles?category=...&tag=... (dev-plan-05 5.2).
func (h *ArticleHandler) ListArticles(c *gin.Context) {
	category := c.Query("category")
	if err := service.ValidateArticleCategory(category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	articles, err := h.Articles.ListPublished(c.Request.Context(), category, c.Query("tag"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list articles"})
		return
	}

	items := make([]gin.H, 0, len(articles))
	for i := range articles {
		items = append(items, h.articleJSON(c, &articles[i]))
	}
	c.JSON(http.StatusOK, gin.H{"articles": items})
}

// GetArticle handles GET /api/articles/:category/:slug.
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	article, err := h.Articles.GetPublishedByCategoryAndSlug(c.Request.Context(), c.Param("category"), c.Param("slug"))
	if err != nil {
		respondArticleError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.articleJSON(c, article))
}

// AdminListArticles handles GET /api/admin/articles?category=... — all
// statuses; category is optional (omit to list both) (dev-plan-11 11.3).
func (h *ArticleHandler) AdminListArticles(c *gin.Context) {
	category := c.Query("category")
	if category != "" {
		if err := service.ValidateArticleCategory(category); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	articles, err := h.Articles.ListAll(c.Request.Context(), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list articles"})
		return
	}

	items := make([]gin.H, 0, len(articles))
	for i := range articles {
		items = append(items, h.articleJSON(c, &articles[i]))
	}
	c.JSON(http.StatusOK, gin.H{"articles": items})
}

// AdminGetArticle handles GET /api/admin/articles/:id — any status, for
// prefilling the edit form.
func (h *ArticleHandler) AdminGetArticle(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	article, err := h.Articles.GetByID(c.Request.Context(), id)
	if err != nil {
		respondArticleError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.articleJSON(c, article))
}

// AdminCreateArticle handles POST /api/admin/articles.
func (h *ArticleHandler) AdminCreateArticle(c *gin.Context) {
	var req articleRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Slug == "" || req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and title are required"})
		return
	}
	if err := service.ValidateArticleCategory(req.Category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.ValidateStatus(req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article, err := h.Articles.Create(c.Request.Context(), req.Category, req.Slug, req.Title, req.Body, req.Status, nil)
	if err != nil {
		respondArticleWriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, articleJSON(article, nil))
}

// AdminUpdateArticle handles PUT /api/admin/articles/:id.
func (h *ArticleHandler) AdminUpdateArticle(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req articleRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Slug == "" || req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and title are required"})
		return
	}
	if err := service.ValidateArticleCategory(req.Category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.ValidateStatus(req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article, err := h.Articles.Update(c.Request.Context(), id, req.Category, req.Slug, req.Title, req.Body, req.Status, nil)
	if err != nil {
		respondArticleWriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, articleJSON(article, nil))
}

// AdminDeleteArticle handles DELETE /api/admin/articles/:id.
func (h *ArticleHandler) AdminDeleteArticle(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.Articles.Delete(c.Request.Context(), id); err != nil {
		respondArticleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// AdminAttachTag handles POST /api/admin/articles/:id/tags — creates the tag
// if it doesn't exist yet, then attaches it to the article.
func (h *ArticleHandler) AdminAttachTag(c *gin.Context) {
	articleID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req attachTagRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" || req.Slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and slug are required"})
		return
	}

	if _, err := h.Articles.GetByID(c.Request.Context(), articleID); err != nil {
		respondArticleError(c, err)
		return
	}

	tag, err := h.Articles.GetOrCreateTag(c.Request.Context(), req.Name, req.Slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if err := h.Articles.AttachTag(c.Request.Context(), articleID, tag.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": strconv.FormatInt(tag.ID, 10), "name": tag.Name, "slug": tag.Slug})
}

func respondArticleError(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}

func respondArticleWriteError(c *gin.Context, err error) {
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

// articleJSON includes tags when the repository is available to fetch them
// (ListArticles/GetArticle via the handler method below); write-path
// responses (create/update) skip the extra query and return an empty list.
func (h *ArticleHandler) articleJSON(c *gin.Context, article *repository.Article) gin.H {
	tags, err := h.Articles.ListTagsForArticle(c.Request.Context(), article.ID)
	if err != nil {
		tags = nil
	}
	return articleJSON(article, tags)
}

func articleJSON(article *repository.Article, tags []repository.Tag) gin.H {
	tagItems := make([]gin.H, 0, len(tags))
	for _, t := range tags {
		tagItems = append(tagItems, gin.H{"id": strconv.FormatInt(t.ID, 10), "name": t.Name, "slug": t.Slug})
	}
	return gin.H{
		"id":           strconv.FormatInt(article.ID, 10),
		"category":     article.Category,
		"slug":         article.Slug,
		"title":        article.Title,
		"body":         article.Body,
		"status":       article.Status,
		"published_at": article.PublishedAt,
		"tags":         tagItems,
		"created_at":   article.CreatedAt,
		"updated_at":   article.UpdatedAt,
	}
}
