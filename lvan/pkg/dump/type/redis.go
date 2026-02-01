package _type

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

// RedisConfig Redis 输出配置
type RedisConfig struct {
	Host     string // Redis 主机地址
	Port     int    // Redis 端口
	Password string // Redis 密码
	DB       int    // Redis 数据库编号
	KeyPrefix string // Redis key 前缀
	TTL      int64  // 过期时间（秒），0 表示永不过期
}

// Redis 创建一个将 MySQL 数据写入 Redis 的导出函数
// 用于 MySQL → Redis 迁移场景
func Redis(config RedisConfig) func(records []dump.Record, pks ...string) string {
	return func(records []dump.Record, pks ...string) string {
		writeToRedis(records, config, pks...)
		return fmt.Sprintf("redis:%s:%d", config.Host, config.Port)
	}
}

// writeToRedis 将记录写入 Redis
func writeToRedis(records []dump.Record, config RedisConfig, pks ...string) {
	// 构建 Redis 客户端地址
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	if config.Port == 0 {
		addr = config.Host // 如果端口为0，假设 Host 已经包含完整地址
	}

	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB,
	})
	defer client.Close()

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Panicf("连接 Redis 失败: %v", err)
	}

	// 遍历所有记录
	successCount := 0
	for _, record := range records {
		// 构造 Redis key
		key := config.KeyPrefix
		if len(pks) > 0 {
			// 使用主键值作为 key 的一部分
			var pkValues []string
			for _, pk := range pks {
				if pkValue, exists := record[pk]; exists {
					pkValues = append(pkValues, string(pkValue))
				}
			}
			if len(pkValues) > 0 {
				key = config.KeyPrefix + joinPrimaryKey(pkValues)
			}
		} else {
			// 没有主键时，使用序列号
			key = fmt.Sprintf("%s%d", config.KeyPrefix, successCount+1)
		}

		// 将记录转换为 JSON
		recordJSON, err := json.Marshal(record)
		if err != nil {
			log.Printf("序列化记录失败: %v", err)
			continue
		}

		// 写入 Redis (使用 String 类型存储 JSON)
		if err := client.Set(ctx, key, recordJSON, 0).Err(); err != nil {
			log.Printf("写入 Redis 失败 [key=%s]: %v", key, err)
			continue
		}

		// 设置 TTL（如果指定了）
		if config.TTL > 0 {
			if err := client.Expire(ctx, key, time.Duration(config.TTL)*time.Second).Err(); err != nil {
				log.Printf("设置 TTL 失败 [key=%s, ttl=%d]: %v", key, config.TTL, err)
			}
		}

		successCount++
	}

	log.Printf("成功写入 %d/%d 条记录到 Redis (prefix: %s)", successCount, len(records), config.KeyPrefix)
}

// joinPrimaryKey 将主键值组合成字符串
func joinPrimaryKey(values []string) string {
	if len(values) == 0 {
		return ""
	}
	if len(values) == 1 {
		return values[0]
	}
	result := values[0]
	for i := 1; i < len(values); i++ {
		result += "_" + values[i]
	}
	return result
}
