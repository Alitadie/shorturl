package handler

import (
	"net/http"
	"shorturl/model"
	"shorturl/repository"

	"github.com/gin-gonic/gin"
)

type RequestBody struct {
	URL string `json:"url" binding:"required"`
}

// CreateShortLink
// @Summary 生成短链接
// @Description 接受 URL 并返回对应的 Base62 短链接
// @Tags Link
// @Accept json
// @Produce json
// @Param body body RequestBody true "Original URL"
// @Success 200 {object} map[string]string
// @Router /shorten [post]
func CreateShortLink(c *gin.Context) {
	var req RequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link := &model.ShortLink{
		OriginalURL: req.URL,
	}

	if err := repository.SaveLinkV2(c.Request.Context(), link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate"})
		return
	}

	shortURL := "http://localhost:8080/" + link.ShortID
	c.JSON(http.StatusOK, gin.H{"short_url": shortURL, "id": link.ShortID})
}

// RedirectLink
// @Summary 重定向短链接
// @Description 根据短链接 ID 重定向到原始 URL
// @Tags Link
// @Accept json
// @Produce json
// @Param id path string true "Short Link ID"
// @Success 200 {object} map[string]string
// @Router /:id [get]
func RedirectLink(c *gin.Context) {
	id := c.Param("id")
	url, err := repository.GetOriginalURL(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "short url not found"})
		return
	}
	c.Redirect(http.StatusFound, url)
}
