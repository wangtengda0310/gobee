package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisManager Redis 数据源管理器
type RedisManager struct {
	client *redis.Client
	config Config
}

// NewRedisManager 创建 Redis 管理器
func NewRedisManager(ctx context.Context, config Config) (*RedisManager, error) {
	// 构建 Redis 客户端地址
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	if config.Port == 0 {
		addr = config.Host // 如果端口为0，假设 Host 已经包含完整地址
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       0, // 默认使用 DB 0
	})

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %w", err)
	}

	return &RedisManager{
		client: client,
		config: config,
	}, nil
}

// GetClient 获取 Redis 客户端
func (m *RedisManager) GetClient() *redis.Client {
	return m.client
}

// GetConfig 获取配置
func (m *RedisManager) GetConfig() Config {
	return m.config
}

// Close 关闭连接
func (m *RedisManager) Close() error {
	if m.client != nil {
		return m.client.Close()
	}
	return nil
}

// GetDB 实现 Manager 接口（Redis 不使用 SQL DB，返回 nil）
func (m *RedisManager) GetDB() *sql.DB {
	return nil
}

// Keys 根据 pattern 获取所有匹配的 key
func (m *RedisManager) Keys(ctx context.Context, pattern string) ([]string, error) {
	return m.client.Keys(ctx, pattern).Result()
}

// Type 获取 key 的类型
func (m *RedisManager) Type(ctx context.Context, key string) (string, error) {
	return m.client.Type(ctx, key).Result()
}

// Get 获取 String 类型的值
func (m *RedisManager) Get(ctx context.Context, key string) (string, error) {
	return m.client.Get(ctx, key).Result()
}

// HGetAll 获取 Hash 类型的所有字段和值
func (m *RedisManager) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return m.client.HGetAll(ctx, key).Result()
}

// ZRangeWithScores 获取 ZSET 的所有成员和分数
func (m *RedisManager) ZRangeWithScores(ctx context.Context, key string) ([]redis.Z, error) {
	return m.client.ZRangeWithScores(ctx, key, 0, -1).Result()
}

// SMembers 获取 Set 的所有成员
func (m *RedisManager) SMembers(ctx context.Context, key string) ([]string, error) {
	return m.client.SMembers(ctx, key).Result()
}

// TTL 获取 key 的过期时间
func (m *RedisManager) TTL(ctx context.Context, key string) (time.Duration, error) {
	return m.client.TTL(ctx, key).Result()
}
