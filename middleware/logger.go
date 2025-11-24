package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var Log *zap.Logger

// 初始化全局日志
func InitLogger() {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	var err error
	Log, err = config.Build()
	if err != nil {
		panic(err)
	}
}

// RequestLogger 中间件:注入TraceID并记录访问日志
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成TraceID
		traceID := c.GetHeader("X-Request-Id")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		// 将TraceID注入到Context中
		c.Set("trace_id", traceID)
		// 也在响应头里带回，方便前端或测试人员拿着 ID 来报 Bug
		c.Header("X-Request-ID", traceID)

		startTime := time.Now()

		c.Next()

		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// 记录访问日志
		Log.Info("HTTP Request",
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("trace_id", traceID), // 关键：以后拿这个 ID 搜日志
		)
	}
}
