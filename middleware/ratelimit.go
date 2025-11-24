package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type IPLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPLimiter(r rate.Limit, b int) *IPLimiter {
	return &IPLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

func (i *IPLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		//限流器 每秒1个令牌 桶容量3个
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}
	return limiter
}

// RateLimitMiddleware Gin 中间件
func RateLimitMiddleware(limiter *IPLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		l := limiter.GetLimiter(ip)

		// 尝试取 1 个令牌，如果取不到，说明超速了
		if !l.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "request too frequent, please try again later",
			})
			return
		}

		c.Next()
	}
}
