package main

import (
	"log"
	"shorturl/config"
	"shorturl/handler"
	"shorturl/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	//1. 初始化配置
	config.Init()
	//2. 初始化布隆过滤器
	repository.InitBloomFilter()

	r := gin.Default()
	r.POST("/shorten", handler.CreateShortLink)
	r.GET("/:id", handler.RedirectLink)
	err := r.Run(":8080")
	if err != nil {
		log.Fatal("failed to run server", err)
	}
}
