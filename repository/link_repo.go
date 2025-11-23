package repository

import (
	"errors"
	"log"
	"shorturl/config"
	"shorturl/model"
	"shorturl/pkg/base62"
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

// SaveLink 使用 Base62 策略
// 输入: 只含 OriginalURL 的对象
// 输出: 存好的完整对象 (含 ID 和 ShortID)
func SaveLinkV2(link *model.ShortLink) error {
	// 开启事务 (Transaction)
	return config.DB.Transaction(func(tx *gorm.DB) error {
		// 生成短链 ID
		// 1. 先插入数据，获取数据库自增 ID (MySQL/SQLite 自动生成)
		// 此时 link.ShortID 是空的，link.ID 会被填入值
		if err := tx.Create(link).Error; err != nil {
			return err
		}
		// 2. 根据自增 ID 生成 Base62 编码
		// 为了防止太短 (比如 ID=1 -> "b")，我们可以加个偏移量 (Start from 1000000)
		link.ShortID = base62.Encode(uint64(link.ID) + 1000000)
		if err := tx.Model(link).Update("short_id", link.ShortID).Error; err != nil {
			return err
		}
		// 3. 写入预热缓存
		if err := config.RDB.Set(config.Ctx, CacheKeyPrefix+link.ShortID, link.OriginalURL, CacheTTL).Err(); err != nil {
			return err
		}
		return nil
	})
}

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
