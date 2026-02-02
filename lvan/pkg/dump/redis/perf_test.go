package redis_test

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/redis"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/service"
)

// TestRedisDump_PERF01 测试导出大量 String (PERF01)
// 需求：10,000 keys < 10秒
func TestRedisDump_PERF01(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试（使用 -short 标志）")
	}

	t.Run("PERF01: 导出大量 String", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		const keyCount = 10000
		t.Logf("准备 %d 个 String keys...", keyCount)

		// 准备测试数据
		for i := 0; i < keyCount; i++ {
			key := "perf:test:" + strconv.Itoa(i)
			value := "value_" + strconv.Itoa(i)
			s.Set(key, value)
		}

		// 创建 Redis manager
		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导出并计时
		output := "perf_test.zip"
		defer os.Remove(output)

		t.Log("开始导出...")
		start := time.Now()

		dumper := redis.NewDumper(mgr)
		result, err := dumper.Dump(context.Background(), "perf:test:*", output)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, keyCount, result.KeysExported)

		t.Logf("导出完成！耗时: %v", duration)

		// 验证性能：应该在 10 秒内完成
		assert.Less(t, duration.Seconds(), 10.0,
			"导出 %d 个 keys 应该在 10 秒内完成，实际耗时: %.2f 秒", keyCount, duration.Seconds())

		// 验证文件存在
		_, err = os.Stat(output)
		assert.NoError(t, err, "导出文件应该存在")
	})
}

// TestRedisDump_PERF02 测试导出大型 Hash (PERF02)
// 需求：10,000 fields < 15秒
func TestRedisDump_PERF02(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试（使用 -short 标志）")
	}

	t.Run("PERF02: 导出大型 Hash", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		const fieldCount = 10000
		t.Logf("准备包含 %d 个字段的 Hash...", fieldCount)

		// 准备测试数据：一个包含 10000 个字段的 Hash
		key := "perf:large:hash"
		for i := 0; i < fieldCount; i++ {
			field := "field_" + strconv.Itoa(i)
			value := "value_" + strconv.Itoa(i)
			s.HSet(key, field, value)
		}

		// 创建 Redis manager
		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导出并计时
		output := "perf_hash_test.zip"
		defer os.Remove(output)

		t.Log("开始导出...")
		start := time.Now()

		dumper := redis.NewDumper(mgr)
		result, err := dumper.Dump(context.Background(), "perf:large:*", output)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysExported, "应该导出 1 个 key")

		t.Logf("导出完成！耗时: %v", duration)

		// 验证性能：应该在 15 秒内完成
		assert.Less(t, duration.Seconds(), 15.0,
			"导出包含 %d 个字段的 Hash 应该在 15 秒内完成，实际耗时: %.2f 秒", fieldCount, duration.Seconds())

		// 验证文件存在
		_, err = os.Stat(output)
		assert.NoError(t, err, "导出文件应该存在")
	})
}

// TestRedisDump_PERF03 测试导出大型 ZSET (PERF03)
// 需求：100,000 members < 30秒
func TestRedisDump_PERF03(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试（使用 -short 标志）")
	}

	t.Run("PERF03: 导出大型 ZSET", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		const memberCount = 100000
		t.Logf("准备包含 %d 个成员的 ZSET...", memberCount)

		// 准备测试数据：一个包含 100000 个成员的 ZSET
		key := "perf:large:zset"
		for i := 0; i < memberCount; i++ {
			member := "member_" + strconv.Itoa(i)
			score := float64(i)
			s.ZAdd(key, score, member)
		}

		// 创建 Redis manager
		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导出并计时
		output := "perf_zset_test.zip"
		defer os.Remove(output)

		t.Log("开始导出...")
		start := time.Now()

		dumper := redis.NewDumper(mgr)
		result, err := dumper.Dump(context.Background(), "perf:large:*", output)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysExported, "应该导出 1 个 key")

		t.Logf("导出完成！耗时: %v", duration)

		// 验证性能：应该在 30 秒内完成
		assert.Less(t, duration.Seconds(), 30.0,
			"导出包含 %d 个成员的 ZSET 应该在 30 秒内完成，实际耗时: %.2f 秒", memberCount, duration.Seconds())

		// 验证文件存在
		_, err = os.Stat(output)
		assert.NoError(t, err, "导出文件应该存在")
	})
}

// TestRedisImport_PERF04 测试导入大量数据 (PERF04)
// 需求：10,000 keys < 20秒
func TestRedisImport_PERF04(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试（使用 -short 标志）")
	}

	t.Run("PERF04: 导入大量数据", func(t *testing.T) {
		// 步骤1: 准备导出数据
		s1 := miniredis.RunT(t)
		defer s1.Close()

		const keyCount = 10000
		t.Logf("准备 %d 个 keys 用于导出...", keyCount)

		// 准备测试数据
		for i := 0; i < keyCount; i++ {
			key := "perf:import:" + strconv.Itoa(i)
			value := "value_" + strconv.Itoa(i)
			s1.Set(key, value)
		}

		// 导出数据
		addr1 := s1.Addr()
		mgr1, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr1,
			Port: 0,
		})
		require.NoError(t, err)

		output := "perf_import_test.zip"
		defer os.Remove(output)

		dumper := redis.NewDumper(mgr1)
		_, err = dumper.Dump(context.Background(), "perf:import:*", output)
		require.NoError(t, err)
		mgr1.Close()

		t.Log("导出完成，开始导入测试...")

		// 步骤2: 导入到新的 Redis 实例并计时
		s2 := miniredis.RunT(t)
		defer s2.Close()

		addr2 := s2.Addr()
		mgr2, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr2,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr2.Close()

		// 执行导入并计时
		t.Log("开始导入...")
		start := time.Now()

		importer := redis.NewImporter(mgr2)
		result, err := importer.Import(context.Background(), output, 0)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, keyCount, result.KeysImported)

		t.Logf("导入完成！耗时: %v", duration)

		// 验证性能：应该在 20 秒内完成
		assert.Less(t, duration.Seconds(), 20.0,
			"导入 %d 个 keys 应该在 20 秒内完成，实际耗时: %.2f 秒", keyCount, duration.Seconds())

		// 验证数据完整性
		keys := s2.Keys()
		assert.Len(t, keys, keyCount, "应该有 %d 个 keys", keyCount)
	})
}

// BenchmarkRedisDump_String 性能基准测试：String 导出
func BenchmarkRedisDump_String(b *testing.B) {
	s := miniredis.RunT(b)
	defer s.Close()

	// 准备 1000 个测试 keys
	const keyCount = 1000
	for i := 0; i < keyCount; i++ {
		s.Set("bench:test:"+strconv.Itoa(i), "value_"+strconv.Itoa(i))
	}

	addr := s.Addr()
	mgr, _ := service.NewRedisManager(context.Background(), service.Config{
		Host: addr,
		Port: 0,
	})
	defer mgr.Close()

	dumper := redis.NewDumper(mgr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output := "bench_test_" + strconv.Itoa(i) + ".zip"
		_, err := dumper.Dump(context.Background(), "bench:test:*", output)
		if err != nil {
			b.Fatal(err)
		}
		os.Remove(output)
	}
}

// BenchmarkRedisImport_String 性能基准测试：String 导入
func BenchmarkRedisImport_String(b *testing.B) {
	// 准备测试数据
	s1 := miniredis.RunT(b)
	defer s1.Close()

	const keyCount = 1000
	for i := 0; i < keyCount; i++ {
		s1.Set("bench:import:"+strconv.Itoa(i), "value_"+strconv.Itoa(i))
	}

	// 导出
	addr1 := s1.Addr()
	mgr1, _ := service.NewRedisManager(context.Background(), service.Config{
		Host: addr1,
		Port: 0,
	})

	dumper := redis.NewDumper(mgr1)
	_, err := dumper.Dump(context.Background(), "bench:import:*", "bench_import_test.zip")
	if err != nil {
		b.Fatal(err)
	}
	mgr1.Close()
	defer os.Remove("bench_import_test.zip")

	// 基准测试导入
	s2 := miniredis.RunT(b)
	defer s2.Close()

	addr2 := s2.Addr()
	mgr2, _ := service.NewRedisManager(context.Background(), service.Config{
		Host: addr2,
		Port: 0,
	})
	defer mgr2.Close()

	importer := redis.NewImporter(mgr2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := importer.Import(context.Background(), "bench_import_test.zip", 0)
		if err != nil {
			b.Fatal(err)
		}
		if result.KeysImported != keyCount {
			b.Fatalf("导入数量不匹配: got %d, want %d", result.KeysImported, keyCount)
		}
		// 清空数据以备下次测试
		s2.FlushAll()
	}
}

// Helper function to check if running in CI environment
func isCI() bool {
	return os.Getenv("CI") != ""
}
