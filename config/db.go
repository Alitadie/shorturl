package config

import (
	"context"
	"log"
	"os"
	"shorturl/model"
	"strconv"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
	Ctx = context.Background()
)

// getEnv 读取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func Init() {
	var err error
	// 1. SQLite 路径配置
	// 容器内路径一般用 /data/shorturl.db，方便挂载卷
	dbPath := getEnv("DB_PATH", "shorturl.db")

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", err)
	}

	if err := DB.AutoMigrate(&model.ShortLink{}); err != nil {
		log.Fatal("failed to migrate database", err)
	}

	// 2. Redis 地址配置 (核心修改)
	// 默认 localhost，但在 Docker 中我们要改成 "redis-service" 这种名字
	redisAddr := getEnv("REDIS_ADDR", "192.168.123.220:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisUsername := getEnv("REDIS_USERNAME", "")
	redisDbStr := getEnv("REDIS_DB", "0")

	// 将字符串转换为 int
	redisDb, err := strconv.Atoi(redisDbStr)
	if err != nil {
		log.Fatal("Error converting REDIS_DB:", err)
	}

	// 创建 Redis 客户端配置
	options := &redis.Options{
		Addr: redisAddr,
		DB:   redisDb,
	}

	// 如果环境变量 REDIS_PASSWORD 存在且不为空，设置密码
	if redisPassword != "" {
		options.Password = redisPassword
	}

	// 如果环境变量 REDIS_USERNAME 存在且不为空，设置用户名
	if redisUsername != "" {
		options.Username = redisUsername
	}

	// 使用配置创建 Redis 客户端
	RDB = redis.NewClient(options)

	if err := RDB.Ping(Ctx).Err(); err != nil {
		log.Fatal("failed to connect redis", err)
	}
	log.Printf("Config Init: DB at %s, Redis at %s", dbPath, redisAddr)
}

func Close() {
	if err := RDB.Close(); err != nil {
		log.Fatal("failed to close redis", err)
	}
	sqlDB, err := DB.DB()
	if err == nil {
		err = sqlDB.Close()
		if err != nil {
			log.Fatal("failed to close Database", err)
		}
	}
	log.Println("Resources closed")
}
