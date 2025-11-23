package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	rdb *redis.Client
	ctx = context.Background()
)

type ShortLink struct {
	gorm.Model
	ShortID     string `gorm:"type:varchar(20);uniqueIndex;not null`
	OriginalURL string `gorm:"type:text;not null"`
}

func initSystem() {
	var err error
	db, err = gorm.Open(sqlite.Open("shorturl.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", err)
	}
	db.AutoMigrate(&ShortLink{})

	rdb = redis.NewClient(&redis.Options{
		Addr:     "192.168.123.220:6379",
		Password: "123456", // no password set
		DB:       0,        // use default DB
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("failed to connect redis", err)
	}
	log.Println("System Init: DB connected, Redis connected")

}

func generateShortURL(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func crateShortLink(originalURL string) (string, error) {
	for i := 0; i < 10; i++ {
		id := generateShortURL(6)
		link := ShortLink{
			ShortID:     id,
			OriginalURL: originalURL,
		}
		result := db.Create(&link)
		if result.Error != nil {
			continue
		}
		return id, nil
	}
	return "", errors.New("failed to create short link")
}

func getOriginalURL(shortID string) (string, error) {
	val, err := rdb.Get(ctx, shortID).Result()
	if err == nil {
		return val, nil
	} else if err != redis.Nil {
		log.Println("Redis Error:", err)
	}

	var link ShortLink
	if err := db.Where("short_id = ?", shortID).First(&link).Error; err != nil {
		return "", err
	}

	err = rdb.Set(ctx, shortID, link.OriginalURL, 24*time.Hour).Err()
	if err != nil {
		log.Println("回填Redis Error:", err)
	}

	return link.OriginalURL, nil
}

type ShortRequest struct {
	URL string `json:"url" binding:"required"`
}

func main() {

	initSystem()

	r := gin.Default()
	r.POST("/shorten", func(c *gin.Context) {
		var req ShortRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		shortID, err := crateShortLink(req.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		shortURL := "http://localhost:8080/" + shortID
		c.JSON(http.StatusOK, gin.H{"short_url": shortURL, "id": shortID})
	})

	r.GET("/:id", func(c *gin.Context) {
		id := c.Param("id")

		url, err := getOriginalURL(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "short url not found"})
			return
		}
		c.Redirect(http.StatusFound, url)
	})

	r.Run(":8080")
}
