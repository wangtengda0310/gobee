package redis_test

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/redis"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/service"
)

// TestProgress_Demo 演示进度条功能
// 使用方法: go test -v -run TestProgress_Demo
// 注意: 此测试会显示实际进度条，建议在终端中运行
func TestProgress_Demo(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过演示测试（使用 -short 标志）")
	}

	t.Run("导出进度演示", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 创建 100 个测试 keys 来演示进度条
		const keyCount = 100
		t.Logf("创建 %d 个测试 keys...", keyCount)

		for i := 0; i < keyCount; i++ {
			key := "progress:demo:" + strconv.Itoa(i)
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

		// 导出数据并查看进度条
		output := "progress_demo.zip"
		defer os.Remove(output)

		t.Log("开始导出（将显示进度条）...")
		dumper := redis.NewDumper(mgr)
		result, err := dumper.Dump(context.Background(), "progress:demo:*", output)

		require.NoError(t, err)
		assert.Equal(t, keyCount, result.KeysExported)
		t.Logf("导出完成！共导出 %d 个 key", result.KeysExported)
	})

	t.Run("导入进度演示", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 准备导出数据
		const keyCount = 50
		for i := 0; i < keyCount; i++ {
			s.Set("progress:import:"+strconv.Itoa(i), "value_"+strconv.Itoa(i))
		}

		// 导出
		addr := s.Addr()
		mgr1, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)

		output := "progress_import_demo.zip"
		defer os.Remove(output)

		dumper := redis.NewDumper(mgr1)
		_, err = dumper.Dump(context.Background(), "progress:import:*", output)
		require.NoError(t, err)
		mgr1.Close()

		// 导入并查看进度条
		t.Log("开始导入（将显示进度条）...")

		s2 := miniredis.RunT(t)
		defer s2.Close()

		addr2 := s2.Addr()
		mgr2, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr2,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr2.Close()

		importer := redis.NewImporter(mgr2)
		result, err := importer.Import(context.Background(), output, 0)

		require.NoError(t, err)
		assert.Equal(t, keyCount, result.KeysImported)
		t.Logf("导入完成！共导入 %d 个 key", result.KeysImported)
	})
}
