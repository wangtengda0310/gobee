package redis

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/service"
)

// TestRedisDump_List 测试 List 类型导出（回归测试用例 RL01）
//
// 业务场景：
// - 消息队列：玩家聊天消息、系统通知
// - 时间线：动态流、事件日志
// - 最新列表：最近操作记录（固定长度）
// - 任务队列：异步任务列表
func TestRedisDump_List(t *testing.T) {
	t.Run("RL01: 消息队列导出", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 业务数据：模拟玩家聊天消息队列
		// LPUSH 顺序：msg1 -> msg2 -> msg3
		// List 存储：[msg3, msg2, msg1] (左端是头部)
		s.Lpush("messages:user:10001", "[10:30] Player1: 大家好")
		s.Lpush("messages:user:10001", "[10:31] Player2: 你好")
		s.Lpush("messages:user:10001", "[10:32] System: 欢迎来到游戏")

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导出
		dumper := NewDumper(mgr)
		output := "messages.zip"
		result, err := dumper.Dump(context.Background(), "messages:user:10001", output)
		require.NoError(t, err, "消息队列导出应该成功")
		assert.Equal(t, 1, result.KeysExported, "应该导出 1 个消息队列 key")
		defer os.Remove(output)

		// 验证 ZIP 结构
		zipReader, err := zip.OpenReader(output)
		require.NoError(t, err)
		defer zipReader.Close()

		// 应该有 3 个文件（3 条消息）
		assert.Equal(t, 3, len(zipReader.File),
			"应该导出 3 条聊天消息")

		// 验证文件名格式：000000_xxx, 000001_xxx, 000002_xxx
		// LPUSH 后 List 顺序：[msg3, msg2, msg1]
		// 所以 index 0 = msg3, index 1 = msg2, index 2 = msg1
		actualFiles := make(map[string]string)
		for _, f := range zipReader.File {
			// 提取文件名（去掉路径前缀）
			fileName := strings.TrimPrefix(f.Name, "messages:user:10001/")
			// 提取索引（前 6 个字符）
			index := fileName[:6]

			// 读取文件内容
			rc, _ := f.Open()
			content, _ := io.ReadAll(rc)
			rc.Close()

			actualFiles[index] = string(content)
		}

		// 验证所有文件存在
		assert.Contains(t, actualFiles, "000000",
			"应该包含索引 0 的文件")
		assert.Contains(t, actualFiles, "000001",
			"应该包含索引 1 的文件")
		assert.Contains(t, actualFiles, "000002",
			"应该包含索引 2 的文件")

		// 验证消息内容顺序
		assert.Equal(t, "[10:32] System: 欢迎来到游戏", actualFiles["000000"],
			"索引 0 应该是最新的消息（最后 LPUSH 的）")
		assert.Equal(t, "[10:31] Player2: 你好", actualFiles["000001"],
			"索引 1 应该是中间的消息")
		assert.Equal(t, "[10:30] Player1: 大家好", actualFiles["000002"],
			"索引 2 应该是最早的消息")
	})

	t.Run("RL02: 空列表处理", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 创建空列表
		s.Lpush("messages:user:10001", "only_one_item")
		s.Lpop("messages:user:10001")

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导出空列表
		dumper := NewDumper(mgr)
		output := "empty.zip"
		_, err = dumper.Dump(context.Background(), "empty:list", output)

		// 空列表应该返回错误
		assert.Error(t, err, "空列表导出应该返回错误")
		assert.Contains(t, err.Error(), "未找到匹配的 key",
			"应该提示未找到 key")
	})

	t.Run("RL03: 包含特殊字符的消息", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 业务消息：包含换行符、引号、表情符号
		messages := []string{
			"Hello\nWorld",        // 包含换行
			`Say "Hello"`,        // 包含引号
			"Emoji: 😀🎮🚀",       // 包含表情
			"中文名: 张三",         // 包含中文
		}

		for _, msg := range messages {
			s.Lpush("chat:room:1", msg)
		}

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导出
		dumper := NewDumper(mgr)
		output := "chat.zip"
		result, err := dumper.Dump(context.Background(), "chat:room:1", output)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysExported, "应该导出 1 个聊天室")
		defer os.Remove(output)

		// 验证 ZIP 内容
		zipReader, _ := zip.OpenReader(output)
		defer zipReader.Close()

		// 读取所有消息
		importedMsgs := make([]string, len(messages))
		for i, f := range zipReader.File {
			rc, _ := f.Open()
			content, _ := io.ReadAll(rc)
			rc.Close()
			importedMsgs[i] = string(content)
		}

		// 验证所有特殊字符都被保留
		// LPUSH 顺序：msg1, msg2, msg3, msg4
		// List 存储：[msg4, msg3, msg2, msg1]
		// 所以 importedMsgs[0] = msg4, importedMsgs[1] = msg3, etc.
		assert.Contains(t, importedMsgs[3], "\n",
			"换行符必须正确保留")
		assert.Contains(t, importedMsgs[2], "\"",
			"引号必须正确保留")
		assert.Contains(t, importedMsgs[1], "😀",
			"Emoji 表情必须正确保留")
		assert.Contains(t, importedMsgs[0], "张三",
			"中文字符必须正确保留")
	})

	t.Run("RL04: 大列表性能测试", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 业务场景：大量历史消息
		messageCount := 1000 // 测试用 1000，生产用 10000
		for i := 0; i < messageCount; i++ {
			s.Lpush("messages:global", fmt.Sprintf("Message #%d: 内容", i))
		}

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导出
		dumper := NewDumper(mgr)
		output := "large_messages.zip"
		result, err := dumper.Dump(context.Background(), "messages:global", output)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysExported, "应该导出 1 个大列表")
		defer os.Remove(output)

		// 验证 ZIP 内容
		zipReader, _ := zip.OpenReader(output)
		defer zipReader.Close()

		assert.Equal(t, messageCount, len(zipReader.File),
			fmt.Sprintf("应该导出 %d 条消息", messageCount))
	})
}

// TestRedisImport_List 测试 List 类型导入（回归测试用例 RLI01）
//
// 业务场景：从备份恢复消息队列、迁移聊天记录
func TestRedisImport_List(t *testing.T) {
	t.Run("RLI01: 消息队列导入", func(t *testing.T) {
		ctx := context.Background()
		s := miniredis.RunT(t)
		defer s.Close()

		// 创建测试 ZIP
		zipPath := "messages_import.zip"
		createTestZip(t, zipPath, map[string]string{
			"messages:user:10001/000000": "[10:32] System: 欢迎",
			"messages:user:10001/000001": "[10:31] Player2: 你好",
			"messages:user:10001/000002": "[10:30] Player1: 大家好",
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
		require.NoError(t, err, "消息队列导入应该成功")
		assert.Equal(t, 1, result.KeysImported, "应该导入 1 个消息队列")

		// 验证 List 顺序
		// 导入后 List 应该是：[000000的内容, 000001的内容, 000002的内容]
		client := mgr.GetClient()
		msg0, _ := client.LIndex(ctx, "messages:user:10001", 0).Result()
		msg1, _ := client.LIndex(ctx, "messages:user:10001", 1).Result()
		msg2, _ := client.LIndex(ctx, "messages:user:10001", 2).Result()

		assert.Equal(t, "[10:32] System: 欢迎", msg0,
			"索引 0 应该是文件 000000 的内容")
		assert.Equal(t, "[10:31] Player2: 你好", msg1,
			"索引 1 应该是文件 000001 的内容")
		assert.Equal(t, "[10:30] Player1: 大家好", msg2,
			"索引 2 应该是文件 000002 的内容")

		// 验证 List 长度
		length, _ := client.LLen(ctx, "messages:user:10001").Result()
		assert.Equal(t, int64(3), length,
			"List 长度应该是 3")
	})

	t.Run("RLI02: List 覆盖已存在的 key", func(t *testing.T) {
		ctx := context.Background()
		s := miniredis.RunT(t)
		defer s.Close()

		// 先创建一个 List
		s.Lpush("messages:user:10001", "old_message_1")
		s.Lpush("messages:user:10001", "old_message_2")

		// 创建新的 ZIP 用于导入
		zipPath := "messages_override.zip"
		createTestZip(t, zipPath, map[string]string{
			"messages:user:10001/000000": "new_message_1",
			"messages:user:10001/000001": "new_message_2",
		})
		defer os.Remove(zipPath)

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 导入会覆盖原有的 List
		importer := NewImporter(mgr)
		result, err := importer.Import(context.Background(), zipPath, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysImported, "应该导入 1 个消息队列")

		// 验证旧的 List 被完全覆盖
		client := mgr.GetClient()
		length, _ := client.LLen(ctx, "messages:user:10001").Result()
		assert.Equal(t, int64(2), length,
			"旧的 List 应该被覆盖，长度应该是 2")

		msg0, _ := client.LIndex(ctx, "messages:user:10001", 0).Result()
		assert.Equal(t, "new_message_1", msg0,
			"旧消息应该被新消息覆盖")
	})

	t.Run("RLI03: 导入包含特殊字符的消息", func(t *testing.T) {
		ctx := context.Background()
		s := miniredis.RunT(t)
		defer s.Close()

		// 创建包含特殊字符的 ZIP
		zipPath := "special_chars.zip"
		createTestZip(t, zipPath, map[string]string{
			"chat:room:1/000000": "Line1\nLine2",           // 包含换行
			"chat:room:1/000001": `Say "Hello"`,             // 包含引号
			"chat:room:1/000002": "Emoji: 😀🎮",              // 包含表情
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
		result, err := importer.Import(context.Background(), zipPath, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysImported)

		// 验证特殊字符被正确保留
		client := mgr.GetClient()
		msg0, _ := client.LIndex(ctx, "chat:room:1", 0).Result()
		msg1, _ := client.LIndex(ctx, "chat:room:1", 1).Result()
		msg2, _ := client.LIndex(ctx, "chat:room:1", 2).Result()

		assert.Contains(t, msg0, "\n",
			"换行符必须正确保留")
		assert.Contains(t, msg1, "\"",
			"引号必须正确保留")
		assert.Contains(t, msg2, "😀",
			"Emoji 表情必须正确保留")
	})

	t.Run("RLI04: List 导入导出循环测试", func(t *testing.T) {
		ctx := context.Background()
		s := miniredis.RunT(t)
		defer s.Close()

		// 准备原始数据
		originalMessages := []string{
			"msg_001",
			"msg_002",
			"msg_003",
		}
		for _, msg := range originalMessages {
			s.Lpush("test:loop", msg)
		}

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// Dump
		dumper := NewDumper(mgr)
		zipPath := "loop_test.zip"
		_, err = dumper.Dump(context.Background(), "test:loop", zipPath)
		require.NoError(t, err)
		defer os.Remove(zipPath)

		// 清空 List (删除 key)
		s.Del("test:loop")

		// Import
		importer := NewImporter(mgr)
		result, err := importer.Import(context.Background(), zipPath, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysImported)

		// 验证数据一致性
		client := mgr.GetClient()
		importedMessages, _ := client.LRange(ctx, "test:loop", 0, -1).Result()

		assert.Equal(t, len(originalMessages), len(importedMessages),
			"消息数量应该保持一致")

		// 注意：List 导出导入后顺序会反转
		// LPUSH: [msg1, msg2, msg3] -> List 存储顺序
		// 导出: index 0=msg3, index 1=msg2, index 2=msg1
		// 导入: RPUSH index 0 -> [0], RPUSH index 1 -> [0,1], RPUSH index 2 -> [0,1,2]
		// 结果: [msg3, msg2, msg1]
		for i, msg := range importedMessages {
			expectedMsg := originalMessages[len(originalMessages)-1-i]
			assert.Equal(t, expectedMsg, msg,
				fmt.Sprintf("索引 %d 的消息应该对应", i))
		}
	})
}
