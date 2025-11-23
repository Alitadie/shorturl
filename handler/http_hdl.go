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

func CreateShortLink(c *gin.Context) {
	var req RequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link := &model.ShortLink{
		OriginalURL: req.URL,
	}

	if err := repository.SaveLinkV2(link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate"})
		return
	}

	shortURL := "http://localhost:8080/" + link.ShortID
	c.JSON(http.StatusOK, gin.H{"short_url": shortURL, "id": link.ShortID})
}

func RedirectLink(c *gin.Context) {
	id := c.Param("id")
	url, err := repository.GetOriginalURL(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "short url not found"})
		return
	}
	c.Redirect(http.StatusFound, url)
}
