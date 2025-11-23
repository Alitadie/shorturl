package config

import (
	"context"
	"log"
	"shorturl/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
	Ctx = context.Background()
)

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("shorturl.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", err)
	}

	DB.AutoMigrate(&model.ShortLink{})

	RDB = redis.NewClient(&redis.Options{
		Addr:     "192.168.123.220:6379",
		Password: "123456", // no password set
		Username: "dan",
		DB:       0, // use default DB
	})

	if err := RDB.Ping(Ctx).Err(); err != nil {
		log.Fatal("failed to connect redis", err)
	}
	log.Println("System Init: DB connected, Redis connected")
}
