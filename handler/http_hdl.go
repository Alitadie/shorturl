package handler

import (
	"net/http"
	"shorturl/model"
	"shorturl/repository"
	"shorturl/service"

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

	var id string
	var err error
	for i := 0; i < 10; i++ {
		id = service.GenerateShortID(6)
		link := &model.ShortLink{
			ShortID:     id,
			OriginalURL: req.URL,
		}

		if err = repository.SaveLink(link); err == nil {
			break
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate"})
		return
	}

	shortURL := "http://localhost:8080/" + id
	c.JSON(http.StatusOK, gin.H{"short_url": shortURL, "id": id})
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
