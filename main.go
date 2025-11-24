package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"shorturl/config"
	"shorturl/handler"
	"shorturl/middleware"
	"shorturl/repository"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1.初始化各类组件
	middleware.InitLogger()
	middleware.Log.Info("Starting server...")

	config.Init()
	repository.InitBloomFilter()

	//2. 配置Gin
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())

	r.POST("/shorten", handler.CreateShortLink)
	r.GET("/:id", handler.RedirectLink)

	// 3. 定义 HTTP Server (这是为了优雅停机必须单独定义的)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 4. 启动 HTTP Server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			middleware.Log.Fatal("listen: failed to run server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 5. 等待优雅停机
	<-quit
	middleware.Log.Info("Shutting down server...")
	// 5.1 设置5秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 5.2 等待5秒后强制关闭
	if err := srv.Shutdown(ctx); err != nil {
		middleware.Log.Fatal("shutdown: failed to shutdown server", zap.Error(err))
	}

	//清理资源
	config.Close()

	middleware.Log.Info("Server stopped")

}
