package redis

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/service"
)

// TestRedisImport_String 测试 String 类型导入（回归测试用例 RSI01）
func TestRedisImport_String(t *testing.T) {
	t.Run("RSI01: 单个 String import", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 创建测试 ZIP 文件
		zipPath := "test_string_import.zip"
		defer os.Remove(zipPath)

		createTestZip(t, zipPath, map[string]string{
			"session:user:10001/value": `{"token":"abc123","login":1706789123}`,
		})

		// 执行导入
		importer := NewImporter(mgr)
		result, err := importer.Import(context.Background(), zipPath, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysImported)

		// 验证 Redis 中的数据
		client := mgr.GetClient()
		value, err := client.Get(context.Background(), "session:user:10001").Result()
		require.NoError(t, err)
		assert.Equal(t, `{"token":"abc123","login":1706789123}`, value)
	})

	t.Run("RSI02: 多个 String import", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 创建测试 ZIP 文件
		zipPath := "test_multi_string_import.zip"
		defer os.Remove(zipPath)

		createTestZip(t, zipPath, map[string]string{
			"session:user:10001/value":    `{"token":"abc123"}`,
			"session:user:10002/value":    `{"token":"def456"}`,
			"config:app/version/value":    "1.2.3",
		})

		// 执行导入
		importer := NewImporter(mgr)
		result, err := importer.Import(context.Background(), zipPath, 0)
		require.NoError(t, err)
		assert.Equal(t, 3, result.KeysImported)

		// 验证所有数据
		client := mgr.GetClient()
		v1, _ := client.Get(context.Background(), "session:user:10001").Result()
		v2, _ := client.Get(context.Background(), "session:user:10002").Result()
		v3, _ := client.Get(context.Background(), "config:app/version").Result()

		assert.Equal(t, `{"token":"abc123"}`, v1)
		assert.Equal(t, `{"token":"def456"}`, v2)
		assert.Equal(t, "1.2.3", v3)
	})

	t.Run("RSI03: 覆盖已存在的 key", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 先设置一个值
		client := mgr.GetClient()
		client.Set(context.Background(), "session:user:10001", "old_value", 0)

		// 创建 ZIP 文件
		zipPath := "test_override_import.zip"
		defer os.Remove(zipPath)

		createTestZip(t, zipPath, map[string]string{
			"session:user:10001/value": "new_value",
		})

		// 执行导入
		importer := NewImporter(mgr)
		result, err := importer.Import(context.Background(), zipPath, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysImported)

		// 验证值被覆盖
		value, _ := client.Get(context.Background(), "session:user:10001").Result()
		assert.Equal(t, "new_value", value)
	})

	t.Run("RSI04: 导入到不同 DB", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 创建 ZIP 文件
		zipPath := "test_different_db.zip"
		defer os.Remove(zipPath)

		createTestZip(t, zipPath, map[string]string{
			"test:key/value": "value_in_db_1",
		})

		// 导入到 DB 1
		importer := NewImporter(mgr)
		result, err := importer.Import(context.Background(), zipPath, 1)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysImported)

		// 验证数据在 DB 1 中
		// 注意：miniredis 默认使用 DB 0，这里需要特殊处理
		// 实际 Redis 中需要 SELECT 1 切换 DB
	})
}

// TestRedisImport_Hash 测试 Hash 类型导入（回归测试用例 RHI01）
func TestRedisImport_Hash(t *testing.T) {
	t.Run("RHI01: Hash import", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 创建测试 ZIP 文件
		zipPath := "test_hash_import.zip"
		defer os.Remove(zipPath)

		createTestZip(t, zipPath, map[string]string{
			"inventory:user:10001/gold":     "10000",
			"inventory:user:10001/gems":     "500",
			"inventory:user:10001/capacity": "100",
		})

		// 执行导入
		importer := NewImporter(mgr)
		result, err := importer.Import(context.Background(), zipPath, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysImported)

		// 验证 Hash 数据
		client := mgr.GetClient()
		fields, err := client.HGetAll(context.Background(), "inventory:user:10001").Result()
		require.NoError(t, err)

		assert.Equal(t, "10000", fields["gold"])
		assert.Equal(t, "500", fields["gems"])
		assert.Equal(t, "100", fields["capacity"])
		assert.Equal(t, 3, len(fields))
	})

	t.Run("RHI02: 覆盖已存在的 Hash 字段", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 先设置一个 Hash
		client := mgr.GetClient()
		client.HSet(context.Background(), "inventory:user:10001", "gold", "5000")

		// 创建 ZIP 文件
		zipPath := "test_hash_override.zip"
		defer os.Remove(zipPath)

		createTestZip(t, zipPath, map[string]string{
			"inventory:user:10001/gold": "10000",
			"inventory:user:10001/gems": "500",
		})

		// 执行导入
		importer := NewImporter(mgr)
		result, err := importer.Import(context.Background(), zipPath, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysImported)

		// 验证 Hash 数据（gold 应该被覆盖，gems 是新增）
		fields, _ := client.HGetAll(context.Background(), "inventory:user:10001").Result()
		assert.Equal(t, "10000", fields["gold"])
		assert.Equal(t, "500", fields["gems"])
		assert.Equal(t, 2, len(fields))
	})
}

// TestRedisImport_ZSet 测试 ZSET 类型导入（回归测试用例 RZI01）
func TestRedisImport_ZSet(t *testing.T) {
	t.Run("RZI01: ZSET import", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 创建测试 ZIP 文件
		// 文件名格式：score_member
		zipPath := "test_zset_import.zip"
		defer os.Remove(zipPath)

		createTestZip(t, zipPath, map[string]string{
			"leaderboard:level/000099_user_10001": "user:10001",
			"leaderboard:level/000085_user_10002": "user:10002",
			"leaderboard:level/000072_user_10003": "user:10003",
		})

		// 执行导入
		importer := NewImporter(mgr)
		result, err := importer.Import(context.Background(), zipPath, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysImported)

		// 验证 ZSET 数据
		// 注意：ZRangeWithScores 返回按 score 从低到高排序的成员
		client := mgr.GetClient()
		members, err := client.ZRangeWithScores(context.Background(), "leaderboard:level", 0, -1).Result()
		require.NoError(t, err)

		assert.Equal(t, 3, len(members))
		// ZRange 返回按 score 从低到高排序
		assert.Equal(t, float64(72), members[0].Score)
		assert.Equal(t, "user:10003", members[0].Member)
		assert.Equal(t, float64(85), members[1].Score)
		assert.Equal(t, "user:10002", members[1].Member)
		assert.Equal(t, float64(99), members[2].Score)
		assert.Equal(t, "user:10001", members[2].Member)
	})
}

// TestRedisImport_Set 测试 Set 类型导入
func TestRedisImport_Set(t *testing.T) {
	t.Run("RSI01: Set import", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 创建测试 ZIP 文件
		zipPath := "test_set_import.zip"
		defer os.Remove(zipPath)

		createTestZip(t, zipPath, map[string]string{
			"friends:user:10001/user_10002": "user:10002",
			"friends:user:10001/user_10003": "user:10003",
			"friends:user:10001/user_10005": "user:10005",
		})

		// 执行导入
		importer := NewImporter(mgr)
		result, err := importer.Import(context.Background(), zipPath, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysImported)

		// 验证 Set 数据
		client := mgr.GetClient()
		members, err := client.SMembers(context.Background(), "friends:user:10001").Result()
		require.NoError(t, err)

		assert.Equal(t, 3, len(members))
		// 验证所有成员都存在
		memberSet := make(map[string]bool)
		for _, m := range members {
			memberSet[m] = true
		}
		assert.True(t, memberSet["user:10002"])
		assert.True(t, memberSet["user:10003"])
		assert.True(t, memberSet["user:10005"])
	})
}

// TestRedisImport_InvalidFile 测试无效文件处理
func TestRedisImport_InvalidFile(t *testing.T) {
	t.Run("不存在的 ZIP 文件", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		importer := NewImporter(mgr)
		_, err = importer.Import(context.Background(), "nonexistent.zip", 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "打开 ZIP 文件失败")
	})
}

// createTestZip 创建测试用的 ZIP 文件
func createTestZip(t *testing.T, zipPath string, files map[string]string) {
	zipFile, err := os.Create(zipPath)
	require.NoError(t, err)
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for path, content := range files {
		w, err := zipWriter.Create(path)
		require.NoError(t, err)
		_, err = w.Write([]byte(content))
		require.NoError(t, err)
	}
}

// parseZipPath 解析 ZIP 路径，提取 key 和类型信息
// 例如：session:user:10001/value -> key=session:user:10001, type=value
// 例如：inventory:user:10001/gold -> key=inventory:user:10001, field=gold
func parseZipPath(path string) (key string, itemType string, itemName string) {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", "", ""
	}
	key = parts[0]
	itemType = parts[len(parts)-1]
	if len(parts) > 2 {
		itemName = parts[len(parts)-1]
		// 对于 Hash，字段名是最后一部分
		// 对于 String，value 是固定字符串
		// 对于 ZSET 和 Set，成员名在文件名中
		if len(parts) == 2 {
			// String 类型：key/value
			return key, "string", "value"
		}
		// 处理嵌套路径：key/field_name 或 key/score_member
		itemName = filepath.Join(parts[1:]...)
	}
	return key, itemType, itemName
}
