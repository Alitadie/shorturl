package repository

import (
	"errors"
	"log"
	"shorturl/config"
	"shorturl/model"
	"shorturl/pkg/base62"
	"sync"
	"time"

	"github.com/bits-and-blooms/bloom/v3"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	CacheKeyPrefix = "short:"
	CacheTTL       = 24 * time.Hour
	EmptyFlag      = "EMPTY_Result"
	EmptyTTL       = 5 * time.Minute
)

// BloomFilter
var (
	// NewWithEstimates(预计存放的数据量, 可接受的误判率)
	// 例如：预计存 100万条，允许 1% 的误判
	bloomFilter = bloom.NewWithEstimates(1000000, 0.01)
	// 布隆过滤器本身是非并发安全的，需要加锁
	bloomMu sync.RWMutex
)

// InitBloomFilter: 系统启动时调用，进行预热
func InitBloomFilter() {
	var offset int
	limit := 1000
	log.Println("🔥 正在预热布隆过滤器...")
	// 分页读取所有数据的 ID (实际生产中可能是读取专门的索引文件或由数据中心推送)
	for {
		var links []model.ShortLink
		// 只查询 ID 和 ShortID 字段，节省内存
		result := config.DB.Select("short_id").Offset(offset).Limit(limit).Find(&links)
		if result.Error != nil || len(links) == 0 {
			break
		}

		bloomMu.Lock()

		for _, link := range links {
			bloomFilter.AddString(link.ShortID)
		}
		bloomMu.Unlock()

		offset += limit
		log.Printf("已加载 %d 条数据...", offset)
	}
	log.Println("✅ 布隆过滤器预热完成！恶意请求防御屏障已开启。")

}

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

		// 3. 【重点】同步添加到布隆过滤器
		bloomMu.Lock()
		bloomFilter.AddString(link.ShortID)
		bloomMu.Unlock()

		// 4. 写入预热缓存
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

	// --- 第一道防线：内存级拦截 (纳秒级) ---
	bloomMu.RLock()
	exists := bloomFilter.TestString(shortID)
	bloomMu.RUnlock()

	if !exists {
		// 如果布隆过滤器说不存在，那就一定不存在
		log.Printf("🛡️ Bloom Filter Blocked: %s", shortID)
		return "", errors.New("link not found (bloom blocked)")
	}

	// --- 第二道防线：Redis (毫秒级) ---
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

	// --- 第三道防线：DB (最慢) ---
	var link model.ShortLink
	if err := config.DB.Where("short_id = ?", shortID).First(&link).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			//防御数据穿透
			//查库没有找到数据
			config.RDB.Set(config.Ctx, key, EmptyFlag, EmptyTTL)
			return "", errors.New("link not found (db intercept)")
		}
		// 理论上能走到这的概率只有 1% (误判率)
		return "", err
	}

	//找到数据回填redis
	err = config.RDB.Set(config.Ctx, key, link.OriginalURL, CacheTTL).Err()
	if err != nil {
		log.Println("回填Redis Error:", err)
	}

	return link.OriginalURL, nil
}
