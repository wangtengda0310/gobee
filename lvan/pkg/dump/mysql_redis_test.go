package dump_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_type "github.com/wangtengda0310/gobee/lvan/pkg/dump/type"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

// TestMySQLToRedis_MR101 测试 MySQL 导出到 Redis (MR101)
//
// 业务场景：生产环境 MySQL → Redis 中转 → 本地环境 MySQL
// 用于跨环境数据迁移，例如玩家充值问题调试
func TestMySQLToRedis_MR101(t *testing.T) {
	t.Run("MR101: MySQL 导出到 Redis", func(t *testing.T) {
		// 启动 miniredis 用于测试
		s := miniredis.RunT(t)
		defer s.Close()

		// 模拟 MySQL 导出的记录
		records := []dump.Record{
			{
				"uid":      []byte("10001"),
				"username": []byte("player1"),
				"level":    []byte("50"),
			},
			{
				"uid":      []byte("10002"),
				"username": []byte("player2"),
				"level":    []byte("30"),
			},
		}

		// 创建 Redis 导出函数
		config := _type.RedisConfig{
			Host:      s.Addr(),
			Port:      0, // miniredis 地址已包含端口
			KeyPrefix: "export:user:",
			TTL:       0, // 永不过期
		}
		exportFunc := _type.Redis(config)

		// 执行导出
		result := exportFunc(records, "uid")

		// 验证返回值
		assert.Contains(t, result, "redis:")

		// 检查 key 是否存在
		exists1 := s.Exists("export:user:10001")
		assert.True(t, exists1, "key export:user:10001 应该存在")

		exists2 := s.Exists("export:user:10002")
		assert.True(t, exists2, "key export:user:10002 应该存在")

		// 验证数据内容
		// 注意：JSON 编码时，[]byte 会被 base64 编码
		data1, _ := s.Get("export:user:10001")
		require.True(t, s.Exists("export:user:10001"), "key export:user:10001 应该存在")
		var record1 map[string]string
		err := json.Unmarshal([]byte(data1), &record1)
		require.NoError(t, err)

		// 验证 base64 编码的值（JSON 对 []byte 的标准处理）
		assert.Equal(t, "MTAwMDE=", record1["uid"], "uid 应该被 base64 编码")
		assert.Equal(t, "cGxheWVyMQ==", record1["username"], "username 应该被 base64 编码")
		assert.Equal(t, "NTA=", record1["level"], "level 应该被 base64 编码")

		data2, _ := s.Get("export:user:10002")
		require.True(t, s.Exists("export:user:10002"), "key export:user:10002 应该存在")
		var record2 map[string]string
		err = json.Unmarshal([]byte(data2), &record2)
		require.NoError(t, err)

		assert.Equal(t, "MTAwMDI=", record2["uid"], "uid 应该被 base64 编码")
		assert.Equal(t, "cGxheWVyMg==", record2["username"], "username 应该被 base64 编码")
		assert.Equal(t, "MzA=", record2["level"], "level 应该被 base64 编码")
	})
}

// TestMySQLToRedis_MR102 测试指定 TTL (MR102)
func TestMySQLToRedis_MR102(t *testing.T) {
	t.Run("MR102: 指定 TTL", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		records := []dump.Record{
			{
				"uid":      []byte("10001"),
				"username": []byte("player1"),
			},
		}

		// 设置 TTL 为 3600 秒
		config := _type.RedisConfig{
			Host:      s.Addr(),
			Port:      0,
			KeyPrefix: "export:session:",
			TTL:       3600, // 1小时
		}
		exportFunc := _type.Redis(config)

		// 执行导出
		exportFunc(records, "uid")

		// 验证 TTL
		ttl := s.TTL("export:session:10001")

		// miniredis 的 TTL 返回值可能略有不同，检查是否接近 3600 秒
		assert.Greater(t, ttl, time.Duration(3500)*time.Second)
		assert.LessOrEqual(t, ttl, time.Duration(3600)*time.Second)
	})
}

// TestMySQLToRedis_MR103 测试批量导出 (MR103)
func TestMySQLToRedis_MR103(t *testing.T) {
	t.Run("MR103: 批量导出", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 模拟批量导出多个记录
		records := []dump.Record{
			{"uid": []byte("10001"), "username": []byte("player1")},
			{"uid": []byte("10002"), "username": []byte("player2")},
			{"uid": []byte("10003"), "username": []byte("player3")},
			{"uid": []byte("10004"), "username": []byte("player4")},
			{"uid": []byte("10005"), "username": []byte("player5")},
		}

		config := _type.RedisConfig{
			Host:      s.Addr(),
			Port:      0,
			KeyPrefix: "export:user:",
			TTL:       0,
		}
		exportFunc := _type.Redis(config)

		// 执行导出
		exportFunc(records, "uid")

		// 验证所有 key 都已创建
		keys := s.Keys()
		assert.Len(t, keys, 5, "应该有 5 个 key")

		// 验证 key 格式
		for _, key := range keys {
			assert.True(t, strings.HasPrefix(key, "export:user:"), "key 应该有正确的前缀")
		}

		// 验证可以使用 KEYS 命令计数
		matchingKeys := s.Keys() // miniredis 的 Keys() 返回所有 key
		assert.Len(t, matchingKeys, 5, "KEYS 计数应该正确")
	})
}

// TestMySQLToRedis_EdgeCases 测试边缘情况
func TestMySQLToRedis_EdgeCases(t *testing.T) {
	t.Run("无主键时使用序列号", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		records := []dump.Record{
			{"username": []byte("player1")},
			{"username": []byte("player2")},
		}

		config := _type.RedisConfig{
			Host:      s.Addr(),
			Port:      0,
			KeyPrefix: "export:user:",
			TTL:       0,
		}
		exportFunc := _type.Redis(config)

		// 执行导出，不传主键
		exportFunc(records)

		// 验证使用序列号作为 key
		exists1 := s.Exists("export:user:1")
		exists2 := s.Exists("export:user:2")

		assert.True(t, exists1, "应该使用序列号 1")
		assert.True(t, exists2, "应该使用序列号 2")
	})

	t.Run("复合主键", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		records := []dump.Record{
			{
				"guild_id": []byte("100"),
				"user_id":  []byte("10001"),
				"role":     []byte("admin"),
			},
		}

		config := _type.RedisConfig{
			Host:      s.Addr(),
			Port:      0,
			KeyPrefix: "export:guild_member:",
			TTL:       0,
		}
		exportFunc := _type.Redis(config)

		// 执行导出，使用复合主键
		exportFunc(records, "guild_id", "user_id")

		// 验证复合主键用下划线连接
		exists := s.Exists("export:guild_member:100_10001")
		assert.True(t, exists, "复合主键应该用下划线连接")
	})

	t.Run("空记录集", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		records := []dump.Record{}

		config := _type.RedisConfig{
			Host:      s.Addr(),
			Port:      0,
			KeyPrefix: "export:user:",
			TTL:       0,
		}
		exportFunc := _type.Redis(config)

		// 执行导出空记录集
		result := exportFunc(records, "uid")

		// 验证返回值
		assert.Contains(t, result, "redis:")

		// 验证 Redis 中没有 key
		keys := s.Keys()
		assert.Len(t, keys, 0, "空记录集不应该创建任何 key")
	})
}

// TestMySQLToRedis_DataIntegrity 测试数据完整性 (V101-V105)
func TestMySQLToRedis_DataIntegrity(t *testing.T) {
	t.Run("V101: String 数据完整性", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		records := []dump.Record{
			{
				"uid":      []byte("10001"),
				"username": []byte("player1"),
				"email":    []byte("player1@example.com"),
			},
		}

		config := _type.RedisConfig{
			Host:      s.Addr(),
			Port:      0,
			KeyPrefix: "export:user:",
			TTL:       0,
		}
		exportFunc := _type.Redis(config)
		exportFunc(records, "uid")

		// 验证数据完全相等
		// 注意：JSON 编码时，[]byte 会被 base64 编码
		require.True(t, s.Exists("export:user:10001"), "key export:user:10001 应该存在")
		data, _ := s.Get("export:user:10001")
		var result map[string]string
		json.Unmarshal([]byte(data), &result)

		assert.Equal(t, "MTAwMDE=", result["uid"], "uid 应该被 base64 编码")
		assert.Equal(t, "cGxheWVyMQ==", result["username"], "username 应该被 base64 编码")
		assert.Equal(t, "cGxheWVyMUBleGFtcGxlLmNvbQ==", result["email"], "email 应该被 base64 编码")
	})

	t.Run("V105: 二进制安全", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 包含 NULL 字节的二进制数据
		binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0x00, 0xFF}

		records := []dump.Record{
			{
				"uid":  []byte("10001"),
				"data": binaryData,
			},
		}

		config := _type.RedisConfig{
			Host:      s.Addr(),
			Port:      0,
			KeyPrefix: "export:binary:",
			TTL:       0,
		}
		exportFunc := _type.Redis(config)
		exportFunc(records, "uid")

		// 验证二进制数据完整
		// 包含 NULL 字节的二进制数据会被 base64 编码
		require.True(t, s.Exists("export:binary:10001"), "key export:binary:10001 应该存在")
		data, _ := s.Get("export:binary:10001")
		var result map[string]string
		json.Unmarshal([]byte(data), &result)

		// JSON 中 base64 编码的二进制数据
		assert.Equal(t, "MTAwMDE=", result["uid"], "uid 应该被 base64 编码")
		// 二进制数据 {0x00, 0x01, 0x02, 0x03, 0x00, 0xFF} 被编码为 "AAECAwD/"
		assert.Equal(t, "AAECAwD/", result["data"], "二进制数据应该被正确 base64 编码")
	})
}
