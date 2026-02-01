package redis

import (
	"archive/zip"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/service"
)

// TestRedisDump_TTL 测试 TTL 导出（回归测试用例 TTL01-TTL04）
//
// 业务场景：
// - 会话缓存：1小时过期
// - 限流计数：60秒过期
// - 临时缓存：5分钟过期
// - 持久数据：永不过期
func TestRedisDump_TTL(t *testing.T) {
	t.Run("TTL01: 导出带 TTL 的 Key", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 业务数据：创建带 TTL 的会话缓存
		s.Set("session:user:10001", `{"token":"abc123"}`)
		s.Set("session:user:10002", `{"token":"def456"}`)
		s.Set("cache:hot:items", `{"items":[1,2,3]}`)

		// 设置过期时间（秒）
		s.FastForward(10 * time.Second) // 模拟时间流逝
		s.SetTTL("session:user:10001", 3600*time.Second) // 1小时
		s.SetTTL("session:user:10002", 7200*time.Second) // 2小时
		s.SetTTL("cache:hot:items", 300*time.Second)     // 5分钟

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导出
		dumper := NewDumper(mgr)
		output := "sessions_with_ttl.zip"
		result, err := dumper.Dump(context.Background(), "session:*", output)
		require.NoError(t, err, "带 TTL 的会话导出应该成功")
		assert.Equal(t, 2, result.KeysExported, "应该导出 2 个会话 key")
		defer os.Remove(output)

		// 验证 ZIP 结构：每个 key 应该有 .metadata.json 文件
		zipReader, err := zip.OpenReader(output)
		require.NoError(t, err)
		defer zipReader.Close()

		// 应该有 4 个文件：2 个 value + 2 个 metadata
		assert.Equal(t, 4, len(zipReader.File),
			"应该导出 2 个数据文件 + 2 个元数据文件")

		// 验证元数据文件存在
		metadataFiles := make(map[string]*zip.File)
		for _, f := range zipReader.File {
			if f.Name[len(f.Name)-14:] == ".metadata.json" {
				metadataFiles[f.Name] = f
			}
		}

		assert.Contains(t, metadataFiles, "session:user:10001/.metadata.json",
			"应该包含 session:user:10001 的元数据")
		assert.Contains(t, metadataFiles, "session:user:10002/.metadata.json",
			"应该包含 session:user:10002 的元数据")

		// 验证元数据内容
		metaFile := metadataFiles["session:user:10001/.metadata.json"]
		rc, err := metaFile.Open()
		require.NoError(t, err)
		var metadata map[string]interface{}
		err = json.NewDecoder(rc).Decode(&metadata)
		rc.Close()
		require.NoError(t, err, "元数据应该是有效的 JSON")

		assert.Equal(t, "session:user:10001", metadata["key"],
			"元数据应该包含 key")
		assert.Equal(t, "string", metadata["type"],
			"元数据应该包含类型")
		assert.Equal(t, float64(3600), metadata["ttl"],
			"元数据应该包含正确的 TTL 值")
	})

	t.Run("TTL02: 导出永久 Key（TTL=-1）", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 业务数据：永久配置
		s.Set("config:game:version", "1.2.3")
		s.Set("config:drop:rates", `{"common":0.8}`)

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导出
		dumper := NewDumper(mgr)
		output := "config_permanent.zip"
		result, err := dumper.Dump(context.Background(), "config:*", output)
		require.NoError(t, err)
		_ = result // 验证在后面通过检查文件
		defer os.Remove(output)

		// 验证永久 key 不应该有元数据文件
		zipReader, _ := zip.OpenReader(output)
		defer zipReader.Close()

		// 确保没有任何元数据文件
		for _, f := range zipReader.File {
			assert.NotContains(t, f.Name, ".metadata.json",
				"永久 key 不应该有元数据文件")
		}

		// 文件数量应该只有 2 个（2 个 value 文件），没有元数据
		assert.Equal(t, 2, len(zipReader.File),
			"永久 key 导出应该只有数据文件，没有元数据文件")
	})

	t.Run("TTL03: 导出混合 TTL 数据", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 业务数据：混合有 TTL 和无 TTL 的数据
		s.Set("session:user:10001", "data1")
		s.Set("config:version", "1.0.0") // 永久 (TTL=-1，不会有元数据)
		s.Set("cache:temp", "temp_data")

		s.SetTTL("session:user:10001", 1800*time.Second) // 30分钟
		s.SetTTL("cache:temp", 300*time.Second)          // 5分钟

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 导出所有数据
		dumper := NewDumper(mgr)
		output := "mixed_ttl.zip"
		result, err := dumper.Dump(context.Background(), "*", output)
		require.NoError(t, err)
		assert.Equal(t, 3, result.KeysExported)
		defer os.Remove(output)

		// 验证：只有带 TTL 的 key 才有元数据
		zipReader, _ := zip.OpenReader(output)
		defer zipReader.Close()

		metadataCount := 0
		metadataKeys := make([]string, 0)
		for _, f := range zipReader.File {
			if len(f.Name) >= 14 && f.Name[len(f.Name)-14:] == ".metadata.json" {
				metadataCount++
				// 提取 key 名称（去掉 /.metadata.json 后缀和末尾的斜杠）
				key := strings.TrimSuffix(f.Name[:len(f.Name)-14], "/")
				metadataKeys = append(metadataKeys, key)
			}
		}

		assert.Equal(t, 2, metadataCount,
			"只有带 TTL 的 key 才应该有元数据文件（永久 key 没有）")

		// 验证具体的元数据 key
		assert.Contains(t, metadataKeys, "session:user:10001",
			"session:user:10001 应该有元数据（TTL=1800）")
		assert.Contains(t, metadataKeys, "cache:temp",
			"cache:temp 应该有元数据（TTL=300）")
		assert.NotContains(t, metadataKeys, "config:version",
			"config:version 不应该有元数据（永久 key）")
	})

	t.Run("TTL04: Hash 带 TTL 导出", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 业务数据：玩家背包 Hash
		s.HSet("inventory:user:10001", "gold", "10000")
		s.HSet("inventory:user:10001", "gems", "500")
		s.SetTTL("inventory:user:10001", 86400*time.Second) // 24小时

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导出
		dumper := NewDumper(mgr)
		output := "inventory_ttl.zip"
		result, err := dumper.Dump(context.Background(), "inventory:*", output)
		require.NoError(t, err)
		_ = result // 验证在后面通过检查文件
		defer os.Remove(output)

		// 验证 Hash 也有元数据
		zipReader, _ := zip.OpenReader(output)
		defer zipReader.Close()

		var foundMetadata bool
		for _, f := range zipReader.File {
			if f.Name == "inventory:user:10001/.metadata.json" {
				foundMetadata = true
				rc, _ := f.Open()
				var metadata map[string]interface{}
				json.NewDecoder(rc).Decode(&metadata)
				rc.Close()

				assert.Equal(t, "hash", metadata["type"],
					"元数据应该标识为 Hash 类型")
				assert.Equal(t, float64(86400), metadata["ttl"],
					"TTL 应该是 86400 秒")
			}
		}

		assert.True(t, foundMetadata,
			"Hash 应该有元数据文件")
	})
}

// TestRedisImport_TTL 测试 TTL 导入（回归测试用例 TTL01-TTL04）
//
// 业务场景：
// - 从备份恢复会话（包括过期时间）
// - 迁移数据到新环境保持 TTL
func TestRedisImport_TTL(t *testing.T) {
	t.Run("TTL01: 导入时恢复 TTL", func(t *testing.T) {
		ctx := context.Background()
		s := miniredis.RunT(t)
		defer s.Close()

		// 创建测试 ZIP（手动构造带元数据的 ZIP）
		zipPath := "import_with_ttl.zip"
		createTestZipWithTTL(t, zipPath, map[string]string{
			"session:user:10001/value":       `{"token":"abc123","login":1706789123}`,
			"session:user:10001/.metadata.json": `{"key":"session:user:10001","type":"string","ttl":3600}`,
		})
		defer os.Remove(zipPath)

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导入
		importer := NewImporter(mgr)
		result, err := importer.Import(ctx, zipPath, 0)
		require.NoError(t, err, "带 TTL 的导入应该成功")
		assert.Equal(t, 1, result.KeysImported)

		// 验证 TTL 被正确恢复
		client := mgr.GetClient()
		ttl, _ := client.TTL(ctx, "session:user:10001").Result()

		// miniredis 返回的 TTL 是 time.Duration
		assert.Equal(t, 3600*time.Second, ttl,
			"导入后的 TTL 应该是 3600 秒")
	})

	t.Run("TTL02: 导入永久数据（TTL=-1）", func(t *testing.T) {
		ctx := context.Background()
		s := miniredis.RunT(t)
		defer s.Close()

		// 创建永久数据的 ZIP
		zipPath := "import_permanent.zip"
		createTestZipWithTTL(t, zipPath, map[string]string{
			"config:version/value":       "1.2.3",
			"config:version/.metadata.json": `{"key":"config:version","type":"string","ttl":-1}`,
		})
		defer os.Remove(zipPath)

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导入
		importer := NewImporter(mgr)
		result, err := importer.Import(ctx, zipPath, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysImported)

		// 验证永久数据的 TTL = -1
		client := mgr.GetClient()
		ttl, _ := client.TTL(ctx, "config:version").Result()

		// miniredis 对于永久 key 返回 0 秒
		// 真实 Redis 返回 -1 秒
		ttlSeconds := int64(ttl.Seconds())
		assert.Contains(t, []int64{-1, 0}, ttlSeconds,
			"永久数据的 TTL 应该是 -1 或 0（miniredis 返回 0）")
	})

	t.Run("TTL03: 导入 Hash 并恢复 TTL", func(t *testing.T) {
		ctx := context.Background()
		s := miniredis.RunT(t)
		defer s.Close()

		// 创建 Hash 的 ZIP
		zipPath := "import_hash_ttl.zip"
		createTestZipWithTTL(t, zipPath, map[string]string{
			"inventory:user:10001/gold":       "10000",
			"inventory:user:10001/gems":       "500",
			"inventory:user:10001/.metadata.json": `{"key":"inventory:user:10001","type":"hash","ttl":86400}`,
		})
		defer os.Remove(zipPath)

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导入
		importer := NewImporter(mgr)
		result, err := importer.Import(ctx, zipPath, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysImported)

		// 验证 Hash TTL
		client := mgr.GetClient()
		ttl, _ := client.TTL(ctx, "inventory:user:10001").Result()

		assert.Equal(t, 86400*time.Second, ttl,
			"Hash 的 TTL 应该是 86400 秒")
	})

	t.Run("TTL04: 导出导入循环保持 TTL", func(t *testing.T) {
		ctx := context.Background()
		s := miniredis.RunT(t)
		defer s.Close()

		// 准备原始数据（带不同 TTL）
		s.Set("session:user:10001", "data1")
		s.Set("session:user:10002", "data2")
		s.SetTTL("session:user:10001", 1800*time.Second) // 30分钟
		s.SetTTL("session:user:10002", 3600*time.Second) // 1小时

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// Dump
		dumper := NewDumper(mgr)
		zipPath := "ttl_loop.zip"
		_, err = dumper.Dump(ctx, "session:*", zipPath)
		require.NoError(t, err)
		defer os.Remove(zipPath)

		// 清空 Redis
		s.Del("session:user:10001")
		s.Del("session:user:10002")

		// Import
		importer := NewImporter(mgr)
		result, err := importer.Import(ctx, zipPath, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, result.KeysImported)

		// 验证 TTL 保持一致
		client := mgr.GetClient()
		ttl1, _ := client.TTL(ctx, "session:user:10001").Result()
		ttl2, _ := client.TTL(ctx, "session:user:10002").Result()

		assert.Equal(t, 1800*time.Second, ttl1,
			"TTL 循环后应该保持 1800 秒")
		assert.Equal(t, 3600*time.Second, ttl2,
			"TTL 循环后应该保持 3600 秒")
	})
}

// createTestZipWithTTL 创建带 TTL 元数据的测试 ZIP 文件
func createTestZipWithTTL(t *testing.T, zipPath string, files map[string]string) {
	zipFile, err := os.Create(zipPath)
	require.NoError(t, err, "创建 ZIP 文件应该成功")
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for path, content := range files {
		w, err := zipWriter.Create(path)
		require.NoError(t, err, "创建 ZIP 条目应该成功")

		_, err = w.Write([]byte(content))
		require.NoError(t, err, "写入 ZIP 内容应该成功")
	}
}
