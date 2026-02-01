package redis

import (
	"context"
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/service"
)

// TestRedisDump_String 测试 String 类型导出（回归测试用例 RS01）
func TestRedisDump_String(t *testing.T) {
	t.Run("RS01: 单个 String dump", func(t *testing.T) {
		// 启动 miniredis
		s := miniredis.RunT(t)
		defer s.Close()

		// 准备测试数据
		s.Set("session:user:10001", `{"token":"abc123","login":1706789123,"ip":"192.168.1.100"}`)

		// 创建 RedisManager 连接到 miniredis
		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 创建导出器
		dumper := NewDumper(mgr)

		// 执行导出
		output := "session_user_10001.zip"
		result, err := dumper.Dump(context.Background(), "session:user:10001", output)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysExported)
		assert.Equal(t, output, result.FileName)

		// 验证 ZIP 文件存在
		_, err = os.Stat(output)
		require.NoError(t, err)

		// 验证 ZIP 内容
		zipReader, err := zip.OpenReader(output)
		require.NoError(t, err)
		defer zipReader.Close()

		// 应该有一个目录：session:user:10001
		assert.Equal(t, 1, len(zipReader.File))

		file := zipReader.File[0]
		assert.Equal(t, "session:user:10001/value", file.Name)

		// 读取文件内容
		rc, err := file.Open()
		require.NoError(t, err)
		defer rc.Close()

		content, err := io.ReadAll(rc)
		require.NoError(t, err)

		// 验证数据内容
		expectedValue := `{"token":"abc123","login":1706789123,"ip":"192.168.1.100"}`
		assert.Equal(t, expectedValue, string(content))

		// 清理
		os.Remove(output)
	})

	t.Run("RS02: 多个 String dump", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 准备测试数据
		s.Set("session:user:10001", `{"token":"abc123"}`)
		s.Set("session:user:10002", `{"token":"def456"}`)

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		dumper := NewDumper(mgr)
		output := "sessions.zip"
		result, err := dumper.Dump(context.Background(), "session:user:*", output)
		require.NoError(t, err)
		assert.Equal(t, 2, result.KeysExported)

		// 验证 ZIP 内容
		zipReader, err := zip.OpenReader(output)
		require.NoError(t, err)
		defer zipReader.Close()
		defer os.Remove(output)

		// 应该有两个目录
		assert.Equal(t, 2, len(zipReader.File))

		// 验证两个 key 都被导出
		keys := make(map[string]bool)
		for _, f := range zipReader.File {
			keys[f.Name] = true
		}

		assert.True(t, keys["session:user:10001/value"])
		assert.True(t, keys["session:user:10002/value"])
	})

	t.Run("RS03: Pattern 匹配", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 准备测试数据 - 不同前缀
		s.Set("session:user:10001", "value1")
		s.Set("session:user:10002", "value2")
		s.Set("config:app", "value3")

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		dumper := NewDumper(mgr)
		output := "test_pattern.zip"
		result, err := dumper.Dump(context.Background(), "session:*", output)
		require.NoError(t, err)

		// 应该只导出 session: 开头的 key
		assert.Equal(t, 2, result.KeysExported)

		// 验证
		zipReader, err := zip.OpenReader(output)
		require.NoError(t, err)
		defer zipReader.Close()
		defer os.Remove(output)

		// 应该只有两个 session key
		assert.Equal(t, 2, len(zipReader.File))
	})
}

// TestRedisDump_Hash 测试 Hash 类型导出（回归测试用例 RH01）
func TestRedisDump_Hash(t *testing.T) {
	t.Run("RH01: Hash dump", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 准备 Hash 数据
		s.HSet("inventory:user:10001", "gold", "10000")
		s.HSet("inventory:user:10001", "gems", "500")
		s.HSet("inventory:user:10001", "capacity", "100")

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		dumper := NewDumper(mgr)
		output := "inventory.zip"
		result, err := dumper.Dump(context.Background(), "inventory:user:10001", output)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysExported)

		// 验证 ZIP 内容
		zipReader, err := zip.OpenReader(output)
		require.NoError(t, err)
		defer zipReader.Close()
		defer os.Remove(output)

		// 应该有 3 个字段文件
		assert.Equal(t, 3, len(zipReader.File))

		// 验证字段都被导出
		fields := make(map[string]bool)
		for _, f := range zipReader.File {
			// 提取字段名（路径格式：inventory:user:10001/field_name）
			parts := strings.Split(f.Name, "/")
			if len(parts) == 2 {
				fields[parts[1]] = true
			}
		}

		assert.True(t, fields["gold"])
		assert.True(t, fields["gems"])
		assert.True(t, fields["capacity"])

		// 验证字段值
		for _, f := range zipReader.File {
			rc, _ := f.Open()
			content, _ := io.ReadAll(rc)
			rc.Close()

			// 从路径中提取字段名：inventory:user:10001/field_name
			parts := strings.Split(f.Name, "/")
			if len(parts) == 2 {
				fieldName := parts[1]
				switch fieldName {
				case "gold":
					assert.Equal(t, "10000", string(content))
				case "gems":
					assert.Equal(t, "500", string(content))
				case "capacity":
					assert.Equal(t, "100", string(content))
				}
			}
		}
	})
}

// TestRedisDump_ZSet 测试 ZSET 类型导出（回归测试用例 RZ01）
func TestRedisDump_ZSet(t *testing.T) {
	t.Run("RZ01: ZSET dump", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 准备 ZSET 数据
		s.ZAdd("leaderboard:level", 99, "user:10001")
		s.ZAdd("leaderboard:level", 85, "user:10002")
		s.ZAdd("leaderboard:level", 72, "user:10003")

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		dumper := NewDumper(mgr)
		output := "leaderboard.zip"
		result, err := dumper.Dump(context.Background(), "leaderboard:level", output)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysExported)

		// 验证 ZIP 内容
		zipReader, err := zip.OpenReader(output)
		require.NoError(t, err)
		defer zipReader.Close()
		defer os.Remove(output)

		// 应该有 3 个成员文件
		assert.Equal(t, 3, len(zipReader.File))

		// 验证文件名格式：score_member
		// 注意：sanitizeFileName 会把 : 替换为 _
		// 文件名格式是：000072_user_10003（score=72, member=user:10003）
		expectedOrder := []struct {
			fileName string
			member   string
			score    int64
		}{
			{"000072_user_10003", "user:10003", 72},
			{"000085_user_10002", "user:10002", 85},
			{"000099_user_10001", "user:10001", 99},
		}

		for i, f := range zipReader.File {
			// 路径格式：leaderboard:level/000072_user_10003
			// 我们只需要验证文件名部分
			fileName := filepath.Base(f.Name)

			// 验证文件名
			assert.Equal(t, expectedOrder[i].fileName, fileName)

			// 验证内容是 member 名称
			rc, _ := f.Open()
			content, _ := io.ReadAll(rc)
			rc.Close()

			assert.Equal(t, expectedOrder[i].member, string(content))
		}
	})
}

// TestRedisDump_Set 测试 Set 类型导出（回归测试用例 RS01）
func TestRedisDump_Set(t *testing.T) {
	t.Run("RS01: Set dump", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 准备 Set 数据
		s.SAdd("friends:user:10001", "user:10002", "user:10003", "user:10005")

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		dumper := NewDumper(mgr)
		output := "friends.zip"
		result, err := dumper.Dump(context.Background(), "friends:user:10001", output)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysExported)

		// 验证 ZIP 内容
		zipReader, err := zip.OpenReader(output)
		require.NoError(t, err)
		defer zipReader.Close()
		defer os.Remove(output)

		// 应该有 3 个成员文件
		assert.Equal(t, 3, len(zipReader.File))

		// 验证所有成员都被导出
		members := make(map[string]bool)
		for _, f := range zipReader.File {
			rc, _ := f.Open()
			content, _ := io.ReadAll(rc)
			rc.Close()
			members[string(content)] = true
		}

		assert.True(t, members["user:10002"])
		assert.True(t, members["user:10003"])
		assert.True(t, members["user:10005"])
	})
}

// TestRedisDump_EmptyPattern 测试空匹配
func TestRedisDump_EmptyPattern(t *testing.T) {
	s := miniredis.RunT(t)
	defer s.Close()

	addr := s.Addr()
	mgr, err := service.NewRedisManager(context.Background(), service.Config{
		Host: addr,
		Port: 0,
	})
	require.NoError(t, err)
	defer mgr.Close()

	dumper := NewDumper(mgr)
	_, err = dumper.Dump(context.Background(), "nonexistent:*", "output.zip")

	// 应该返回错误
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未找到匹配的 key")
}
