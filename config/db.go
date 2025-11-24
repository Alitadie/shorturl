package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"shorturl/model"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	initDB()
	initRedis()
}

func initDB() {
	var dialector gorm.Dialector

	// 1.读取驱动类型默认sqlite
	driver := getEnv("DB_DRIVER", "sqlite")
	dsn := ""

	log.Printf("正在初始化数据库，使用 %s", driver)

	// 2.根据类型创建Dialector
	switch driver {
	case "sqlite":
		dialector = sqlite.Open(getEnv("DB_PATH", "data/shorturl.db"))
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			getEnv("DB_USER", "root"),
			getEnv("DB_PASSWORD", "root"),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "3306"),
			getEnv("DB_NAME", "shorturl"),
		)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASSWORD", "postgres"),
			getEnv("DB_NAME", "shorturl"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_SSL_MODE", "disable"),
		)
		dialector = postgres.Open(dsn)
	default:
		log.Fatalf("不支持的数据库类型: %s", driver)
	}

	// 3 连接数据库
	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatal("failed to connect database", err)
	}

	// 4. 【关键】配置连接池 (Production Ready)
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("获取底层 SQL 对象失败")
	}

	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns 设置打开数据库连接的最大数量
	// 注意：这个值不要超过数据库本身的 max_connections 配置
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime 设置了连接可复用的最大时间
	// 避免连接长时间不用被防火墙切断导致 "Broken Pipe"
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 5. 自动迁移表结构
	log.Println("执行数据库自动迁移...")
	if err := DB.AutoMigrate(&model.ShortLink{}); err != nil {
		log.Fatal("failed to migrate database", err)
	}

}

func initRedis() {
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
	log.Printf("Config Init Redis at %s", redisAddr)
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
