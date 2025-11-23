package repository

import (
	"errors"
	"log"
	"shorturl/config"
	"shorturl/model"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	CacheKeyPrefix = "short:"
	CacheTTL       = 24 * time.Hour
	EmptyFlag      = "EMPTY_Result"
	EmptyTTL       = 5 * time.Minute
)

// Save 存储数据
func SaveLink(link *model.ShortLink) error {
	//写DB
	if err := config.DB.Create(link).Error; err != nil {
		return err
	}
	//写入预热缓存
	if err := config.RDB.Set(config.Ctx, CacheKeyPrefix+link.ShortID, link.OriginalURL, CacheTTL).Err(); err != nil {
		return err
	}
	return nil
}

func GetOriginalURL(shortID string) (string, error) {
	key := CacheKeyPrefix + shortID
	log.Println("key:", key)
	val, err := config.RDB.Get(config.Ctx, key).Result()
	if err == nil {
		if val == EmptyFlag {
			log.Println("命中缓存空对象拦截")
			return "", errors.New("link not found (cache intercept)")
		}
		return val, nil
	} else if err != redis.Nil {
		//redis 系统本身报错 记录日志 降级查库
		log.Println("Redis Error:", err)
	}

	var link model.ShortLink
	if err := config.DB.Where("short_id = ?", shortID).First(&link).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			//防御数据穿透
			//查库没有找到数据
			config.RDB.Set(config.Ctx, key, EmptyFlag, EmptyTTL)
			return "", errors.New("link not found (db intercept)")
		}
		return "", err
	}

	//找到数据回填redis
	err = config.RDB.Set(config.Ctx, key, link.OriginalURL, CacheTTL).Err()
	if err != nil {
		log.Println("回填Redis Error:", err)
	}

	return link.OriginalURL, nil
}
