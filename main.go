package main

import (
	"shorturl/config"
	"shorturl/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Init()

	r := gin.Default()
	r.POST("/shorten", handler.CreateShortLink)
	r.GET("/:id", handler.RedirectLink)
	r.Run(":8080")
}
