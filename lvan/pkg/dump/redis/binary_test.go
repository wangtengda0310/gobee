package redis

import (
	"context"
	"crypto/rand"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/service"
)

// TestRedisDump_Binary_Protobuf 测试 Protobuf 二进制数据导入导出（回归测试用例 BIN01）
//
// 业务场景：游戏玩家存档使用 Protobuf 序列化，包含大量二进制数据
// 数据特点：包含 varint 编码、字段标记、可能的 NULL 字节
func TestRedisDump_Binary_Protobuf(t *testing.T) {
	t.Run("BIN01: Protobuf 数据完整性", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 模拟真实的 Protobuf 数据（游戏玩家存档）
		// field 1 (varint): player_id = 10001
		// field 2 (string): player_name = "测试玩家"
		// field 3 (bytes): raw_data = 包含 NULL 字节的二进制数据
		protobufData := []byte{
			// field 1, wire type 0 (varint), value 10001
			0x08, 0x81, 0x9C, 0x04,
			// field 2, wire type 2 (length-delimited), length 12 (UTF-8 "测试玩家")
			0x12, 0x0C,
			// UTF-8 编码的中文
			0xE6, 0xB5, 0x8B, 0xE8, 0xAF, 0x95, 0xE7, 0x8E, 0xB0, 0xE5, 0xAE, 0xB6,
			// field 3, wire type 2 (length-delimited), length 8
			0x1A, 0x08,
			// 二进制数据（包含 NULL 字节）
			0xAB, 0x00, 0xCD, 0x00, 0xEF, 0xFF, 0x01, 0x02,
		}

		// 业务数据：将 Protobuf 存档存储到 Redis
		s.Set("player:save:10001", string(protobufData))

		// 执行导出
		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err, "Redis 管理器初始化应该成功")
		defer mgr.Close()

		dumper := NewDumper(mgr)
		output := "player_save.zip"
		result, err := dumper.Dump(context.Background(), "player:save:10001", output)
		require.NoError(t, err, "导出 Protobuf 数据应该成功")
		assert.Equal(t, 1, result.KeysExported, "应该导出 1 个 key")
		defer os.Remove(output)

		// 业务验证：导出的数据必须与原始 Protobuf 数据完全一致
		// Protobuf 解析对字节顺序非常敏感，任何一个字节错误都会导致解析失败
		client := mgr.GetClient()
		exportedData, err := client.Get(context.Background(), "player:save:10001").Bytes()
		require.NoError(t, err, "从 Redis 读取导出的数据应该成功")

		assert.Equal(t, protobufData, exportedData,
			"Protobuf 数据必须完全一致，否则游戏将无法解析玩家存档")

		// 额外验证：检查 Protobuf 头部字节
		assert.Equal(t, []byte{0x08, 0x81, 0x9C, 0x04}, exportedData[0:4],
			"Protobuf 头部字段标记必须正确")
	})

	t.Run("BIN01: Protobuf 导入循环验证", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 准备测试数据：模拟装备的 Protobuf 数据
		equipmentProto := []byte{
			0x08, 0x64,             // field 1: equipment_id = 100
			0x10, 0x05,             // field 2: level = 5
			0x20, 0xFF, 0xFF, 0xFF, 0xFF, 0x07, // field 3: durability = 2147483647 (int32 max)
			0x2A, 0x04, 0x00, 0x01, 0x02, 0x03, // field 5: enchantments = [0,1,2,3]
		}

		s.Set("equipment:10001", string(equipmentProto))

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行 dump → import 循环
		dumper := NewDumper(mgr)
		zipFile := "equipment.zip"
		_, err = dumper.Dump(context.Background(), "equipment:10001", zipFile)
		require.NoError(t, err)
		defer os.Remove(zipFile)

		// 清空 Redis，模拟导入到新环境
		s.Del("equipment:10001")

		// 执行导入
		importer := NewImporter(mgr)
		importResult, err := importer.Import(context.Background(), zipFile, 0)
		require.NoError(t, err, "装备数据导入应该成功")
		assert.Equal(t, 1, importResult.KeysImported, "应该导入 1 个装备 key")

		// 业务验证：导入后的装备数据必须与原始数据完全一致
		// 否则可能导致装备属性错误、耐久度丢失等游戏问题
		client := mgr.GetClient()
		importedData, err := client.Get(context.Background(), "equipment:10001").Bytes()
		require.NoError(t, err)

		assert.Equal(t, equipmentProto, importedData,
			"装备 Protobuf 数据必须完全一致，否则可能导致装备属性错误")
	})
}

// TestRedisDump_Binary_NullBytes 测试 NULL 字节处理（回归测试用例 BIN02）
//
// 业务场景：C/C++ 程序经常使用 NULL 字节作为字符串终止符
// 数据特点：包含 \x00，使用 string 处理会被截断
func TestRedisDump_Binary_NullBytes(t *testing.T) {
	t.Run("BIN02: NULL 字节在字符串中间", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 业务数据：C 结构体序列化数据，NULL 字节作为字段分隔符
		// 格式: "field1\x00field2\x00field3"
		cStructData := []byte{
			'a', 'b', 'c', 0x00,  // field 1: "abc" + NULL
			'd', 'e', 'f', 0x00,  // field 2: "def" + NULL
			'g', 'h', 'i', 0x00,  // field 3: "ghi" + NULL
		}

		s.Set("struct:user:10001", string(cStructData))

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导出
		dumper := NewDumper(mgr)
		result, err := dumper.Dump(context.Background(), "struct:user:10001", "struct.zip")
		require.NoError(t, err, "C 结构体数据导出应该成功")
		assert.Equal(t, 1, result.KeysExported, "应该导出 1 个结构体")
		defer os.Remove("struct.zip")

		// 验证 NULL 字节被正确保留
		// 如果 NULL 被截断，C 程序将无法正确解析结构体
		client := mgr.GetClient()
		exported, err := client.Get(context.Background(), "struct:user:10001").Bytes()
		require.NoError(t, err, "读取导出的结构体数据应该成功")

		assert.Equal(t, cStructData, exported,
			"C 结构体中的 NULL 字节必须保留，否则字段解析将出错")

		// 显式检查 NULL 字节的位置
		assert.Equal(t, byte(0x00), exported[3],
			"第 4 个字节应该是 NULL（字段分隔符）")
		assert.Equal(t, byte(0x00), exported[7],
			"第 8 个字节应该是 NULL（字段分隔符）")
	})

	t.Run("BIN02: 开头和结尾都是 NULL 字节", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 边界情况：数据开头和结尾都是 NULL
		boundaryData := []byte{0x00, 0x01, 0x02, 0x00}
		s.Set("boundary:test", string(boundaryData))

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		dumper := NewDumper(mgr)
		_, err = dumper.Dump(context.Background(), "boundary:test", "boundary.zip")
		require.NoError(t, err)
		defer os.Remove("boundary.zip")

		// 验证：开头和结尾的 NULL 都必须保留
		client := mgr.GetClient()
		exported, _ := client.Get(context.Background(), "boundary:test").Bytes()

		assert.Equal(t, byte(0x00), exported[0],
			"开头的 NULL 字节必须保留")
		assert.Equal(t, byte(0x00), exported[len(exported)-1],
			"结尾的 NULL 字节必须保留")
		assert.Equal(t, boundaryData, exported,
			"边界情况数据必须完全一致")
	})
}

// TestRedisDump_Binary_UTF8 测试 UTF-8 编码保持（回归测试用例 BIN03）
//
// 业务场景：国际化游戏，玩家名称、聊天记录包含多语言文本
// 数据特点：中文、日文、韩文、Emoji 等多字节 UTF-8 字符
func TestRedisDump_Binary_UTF8(t *testing.T) {
	t.Run("BIN03: 中文文本编码", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 业务数据：玩家信息
		chineseText := `{"name":"张三","guild":"测试公会","location":"北京"}`
		s.Set("player:info:10001", chineseText)

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导出
		dumper := NewDumper(mgr)
		result, err := dumper.Dump(context.Background(), "player:info:10001", "player.zip")
		require.NoError(t, err, "中文数据导出应该成功")
		assert.Equal(t, 1, result.KeysExported, "应该导出 1 个玩家信息")
		defer os.Remove("player.zip")

		// 业务验证：中文字符必须完整保留
		// UTF-8 中文字符通常占用 3 个字节，不能被截断或破坏
		client := mgr.GetClient()
		exported, err := client.Get(context.Background(), "player:info:10001").Result()
		require.NoError(t, err)

		assert.Equal(t, chineseText, exported,
			"中文字符必须完整保留，否则玩家信息将显示乱码")
		assert.Contains(t, exported, "张三",
			"玩家姓名必须正确导出")
		assert.Contains(t, exported, "测试公会",
			"公会名称必须正确导出")
	})

	t.Run("BIN03: 多语言混合文本", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 业务数据：国际化聊天记录
		multilingualText := "Hello 世界 🎮 こんにちは 🌍 مرحبا"
		s.Set("chat:global", multilingualText)

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		dumper := NewDumper(mgr)
		_, err = dumper.Dump(context.Background(), "chat:global", "chat.zip")
		require.NoError(t, err)
		defer os.Remove("chat.zip")

		// 验证：所有语言的字符都必须正确保留
		client := mgr.GetClient()
		exported, _ := client.Get(context.Background(), "chat:global").Result()

		assert.Equal(t, multilingualText, exported,
			"多语言聊天记录必须完整保留")
		assert.Contains(t, exported, "🎮",
			"Emoji 表情必须正确导出")
		assert.Contains(t, exported, "こんにちは",
			"日文字符必须正确导出")
		assert.Contains(t, exported, "مرحبا",
			"阿拉伯文字符必须正确导出")
	})
}

// TestRedisDump_Binary_LargeValue 测试大 Value 处理（回归测试用例 BIN04）
//
// 业务场景：大型对象存储，如战斗录像、游戏地图数据、用户上传文件
// 数据特点：单个 value 可达 1MB 或更大
func TestRedisDump_Binary_LargeValue(t *testing.T) {
	t.Run("BIN04: 1MB 随机二进制数据", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 业务数据：战斗录像（1MB 随机二进制数据）
		largeData := make([]byte, 1024*1024) // 1MB
		_, err := rand.Read(largeData)
		require.NoError(t, err, "生成随机数据应该成功")

		s.Set("battle:replay:12345", string(largeData))

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导出
		dumper := NewDumper(mgr)
		result, err := dumper.Dump(context.Background(), "battle:replay:12345", "replay.zip")
		require.NoError(t, err, "1MB 战斗录像导出应该成功")
		assert.Equal(t, 1, result.KeysExported, "应该导出 1 个战斗录像")
		defer os.Remove("replay.zip")

		// 业务验证：大型战斗录像必须完整导出，任何字节丢失都会导致回放失败
		client := mgr.GetClient()
		exported, err := client.Get(context.Background(), "battle:replay:12345").Bytes()
		require.NoError(t, err, "读取 1MB 战斗录像应该成功")

		assert.Equal(t, len(largeData), len(exported),
			"战斗录像大小必须保持一致，丢失任何字节都会导致回放中断")
		assert.Equal(t, largeData, exported,
			"战斗录像内容必须完全一致，否则回放时会出错")

		// 验证数据完整性：对比开头和结尾
		assert.Equal(t, largeData[:100], exported[:100],
			"战斗录像开头数据必须一致")
		assert.Equal(t, largeData[len(largeData)-100:], exported[len(exported)-100:],
			"战斗录像结尾数据必须一致")
	})

	t.Run("BIN04: 大数据导入导出循环", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 业务数据：游戏地图数据 (512KB)
		mapData := make([]byte, 512*1024)
		for i := range mapData {
			mapData[i] = byte(i % 256) // 填充模式数据
		}

		s.Set("map:data:001", string(mapData))

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// dump → import 循环
		dumper := NewDumper(mgr)
		zipFile := "map_data.zip"
		_, err = dumper.Dump(context.Background(), "map:data:001", zipFile)
		require.NoError(t, err)
		defer os.Remove(zipFile)

		s.Del("map:data:001")

		importer := NewImporter(mgr)
		result, err := importer.Import(context.Background(), zipFile, 0)
		require.NoError(t, err, "地图数据导入应该成功")
		assert.Equal(t, 1, result.KeysImported, "应该导入 1 个地图数据")

		// 验证：地图数据必须完整，否则会导致游戏地图显示错误
		client := mgr.GetClient()
		imported, err := client.Get(context.Background(), "map:data:001").Bytes()
		require.NoError(t, err)

		assert.Equal(t, mapData, imported,
			"地图数据必须完全一致，否则游戏地图将显示错误区域")
	})
}

// TestRedisDump_Binary_MixedTypes 测试混合二进制数据类型
//
// 业务场景：同一个 Hash 中包含字符串、数字、二进制数据
func TestRedisDump_Binary_MixedTypes(t *testing.T) {
	t.Run("BIN05: Hash 包含二进制字段", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		// 业务数据：用户配置 Hash
		// 包含：文本字段、数字字段、二进制字段
		s.HSet("user:config:10001", "username", "player001")
		s.HSet("user:config:10001", "level", "99")

		// 二进制字段：用户头像图片数据（模拟 PNG 头部）
		avatarData := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG 签名
			0x00, 0x00, 0x00, 0x0D, // IHDR 长度
			0x49, 0x48, 0x44, 0x52, // IHDR 类型
			0xFF, 0x00, 0x01, 0x00, // 包含 NULL 字节的图像数据
		}
		s.HSet("user:config:10001", "avatar", string(avatarData))

		addr := s.Addr()
		mgr, err := service.NewRedisManager(context.Background(), service.Config{
			Host: addr,
			Port: 0,
		})
		require.NoError(t, err)
		defer mgr.Close()

		// 执行导出
		dumper := NewDumper(mgr)
		result, err := dumper.Dump(context.Background(), "user:config:10001", "config.zip")
		require.NoError(t, err)
		assert.Equal(t, 1, result.KeysExported, "应该导出 1 个配置 Hash")
		defer os.Remove("config.zip")

		// 验证：所有字段类型都必须正确处理
		client := mgr.GetClient()
		username, _ := client.HGet(context.Background(), "user:config:10001", "username").Result()
		level, _ := client.HGet(context.Background(), "user:config:10001", "level").Result()
		avatar, _ := client.HGet(context.Background(), "user:config:10001", "avatar").Bytes()

		assert.Equal(t, "player001", username,
			"文本字段必须正确导出")
		assert.Equal(t, "99", level,
			"数字字段必须正确导出")
		assert.Equal(t, avatarData, avatar,
			"二进制头像数据必须完整保留，否则图片将损坏")
	})
}
