package redis

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/service"
)

// Dumper Redis 数据导出器
type Dumper struct {
	mgr *service.RedisManager
}

// NewDumper 创建 Redis 导出器
func NewDumper(mgr *service.RedisManager) *Dumper {
	return &Dumper{mgr: mgr}
}

// DumpResult 导出结果
type DumpResult struct {
	KeysExported int
	FileName     string
}

// Dump 导出 Redis 数据
func (d *Dumper) Dump(ctx context.Context, pattern, output string) (*DumpResult, error) {
	client := d.mgr.GetClient()

	// 1. 获取所有匹配的 keys
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("获取 keys 失败: %w", err)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("未找到匹配的 key: pattern=%s", pattern)
	}

	log.Printf("找到 %d 个匹配的 key", len(keys))

	// 2. 确定输出文件名
	if output == "" {
		output = "redis_export.zip"
	}

	// 3. 创建 ZIP 文件
	zipFile, err := os.Create(output)
	if err != nil {
		return nil, fmt.Errorf("创建 ZIP 文件失败: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 4. 导出每个 key
	for _, key := range keys {
		if err := d.dumpKeyToZip(ctx, client, key, zipWriter); err != nil {
			return nil, fmt.Errorf("导出 key %s 失败: %w", key, err)
		}
	}

	log.Printf("成功导出 %d 个 key 到 %s", len(keys), output)

	return &DumpResult{
		KeysExported: len(keys),
		FileName:     output,
	}, nil
}

// dumpKeyToZip 导出单个 key 到 ZIP
func (d *Dumper) dumpKeyToZip(ctx context.Context, client *redis.Client, key string, zipWriter *zip.Writer) error {
	// 获取 key 的类型
	keyType, err := client.Type(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("获取 key 类型失败: %w", err)
	}

	// 获取 TTL
	ttl, err := client.TTL(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("获取 TTL 失败: %w", err)
	}

	// 记录日志
	ttlSeconds := int64(ttl) / int64(time.Second)
	if ttlSeconds == 0 {
		// 在 Redis 中 TTL=-1 表示永久，但某些客户端可能返回 0
		// 我们通过检查 TTL 是否小于 0 来判断
		if ttl < 0 {
			ttlSeconds = -1
		}
	}
	log.Printf("导出 key: %s (类型: %s, TTL: %d秒)", key, keyType, ttlSeconds)

	// 根据类型导出
	switch keyType {
	case "string":
		if err := d.dumpStringToZip(ctx, client, key, zipWriter); err != nil {
			return err
		}
	case "hash":
		if err := d.dumpHashToZip(ctx, client, key, zipWriter); err != nil {
			return err
		}
	case "zset":
		if err := d.dumpZSetToZip(ctx, client, key, zipWriter); err != nil {
			return err
		}
	case "set":
		if err := d.dumpSetToZip(ctx, client, key, zipWriter); err != nil {
			return err
		}
	case "list":
		if err := d.dumpListToZip(ctx, client, key, zipWriter); err != nil {
			return err
		}
	default:
		return fmt.Errorf("不支持的 key 类型: %s", keyType)
	}

	// 创建元数据文件（仅当 TTL != -1 时）
	// 永久 key (TTL=-1) 不需要元数据，保持向后兼容
	if ttlSeconds != -1 {
		if err := d.createMetadataFile(key, keyType, ttlSeconds, zipWriter); err != nil {
			return fmt.Errorf("创建元数据文件失败: %w", err)
		}
	}

	return nil
}

// dumpStringToZip 导出 String 类型到 ZIP
func (d *Dumper) dumpStringToZip(ctx context.Context, client *redis.Client, key string, zipWriter *zip.Writer) error {
	// 获取值 - 使用 Bytes() 确保二进制安全
	// 这对于包含 NULL 字节、Protobuf 数据、非 UTF-8 编码的数据非常重要
	value, err := client.Get(ctx, key).Bytes()
	if err != nil {
		return fmt.Errorf("获取 String 值失败: %w", err)
	}

	// 创建 ZIP 中的文件路径：key/value
	// 注意：ZIP 标准使用 / 作为路径分隔符，即使在 Windows 上
	filePath := key + "/value"

	// 创建 ZIP 中的文件
	w, err := zipWriter.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建 ZIP 文件失败: %w", err)
	}

	// 写入数据 - 直接写入 []byte，无需转换
	_, err = w.Write(value)
	if err != nil {
		return fmt.Errorf("写入数据失败: %w", err)
	}

	return nil
}

// dumpHashToZip 导出 Hash 类型到 ZIP
func (d *Dumper) dumpHashToZip(ctx context.Context, client *redis.Client, key string, zipWriter *zip.Writer) error {
	// 获取所有字段
	fields, err := client.HGetAll(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("获取 Hash 字段失败: %w", err)
	}

	// 写入每个字段
	for field, value := range fields {
		filePath := key + "/" + field

		w, err := zipWriter.Create(filePath)
		if err != nil {
			return fmt.Errorf("创建 ZIP 文件失败: %w", err)
		}

		if _, err := w.Write([]byte(value)); err != nil {
			return fmt.Errorf("写入数据失败: %w", err)
		}
	}

	return nil
}

// dumpZSetToZip 导出 ZSET 类型到 ZIP
func (d *Dumper) dumpZSetToZip(ctx context.Context, client *redis.Client, key string, zipWriter *zip.Writer) error {
	// 获取所有成员和分数
	members, err := client.ZRangeWithScores(ctx, key, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("获取 ZSET 成员失败: %w", err)
	}

	// 写入每个成员，文件名格式：score_member
	for _, m := range members {
		// 格式化分数为 6 位整数，保证排序
		scoreStr := fmt.Sprintf("%06d", int64(m.Score))
		memberName := fmt.Sprintf("%s_%s", scoreStr, sanitizeFileName(m.Member.(string)))
		filePath := key + "/" + memberName

		w, err := zipWriter.Create(filePath)
		if err != nil {
			return fmt.Errorf("创建 ZIP 文件失败: %w", err)
		}

		if _, err := w.Write([]byte(m.Member.(string))); err != nil {
			return fmt.Errorf("写入数据失败: %w", err)
		}
	}

	return nil
}

// dumpSetToZip 导出 Set 类型到 ZIP
func (d *Dumper) dumpSetToZip(ctx context.Context, client *redis.Client, key string, zipWriter *zip.Writer) error {
	// 获取所有成员
	members, err := client.SMembers(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("获取 Set 成员失败: %w", err)
	}

	// 写入每个成员
	for _, member := range members {
		filePath := key + "/" + sanitizeFileName(member)

		w, err := zipWriter.Create(filePath)
		if err != nil {
			return fmt.Errorf("创建 ZIP 文件失败: %w", err)
		}

		if _, err := w.Write([]byte(member)); err != nil {
			return fmt.Errorf("写入数据失败: %w", err)
		}
	}

	return nil
}

// dumpListToZip 导出 List 类型到 ZIP
func (d *Dumper) dumpListToZip(ctx context.Context, client *redis.Client, key string, zipWriter *zip.Writer) error {
	// 获取所有元素
	elements, err := client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("获取 List 元素失败: %w", err)
	}

	// 写入每个元素，文件名格式：index_element (从 0 开始)
	for i, element := range elements {
		indexStr := fmt.Sprintf("%06d", i)
		elementName := fmt.Sprintf("%s_%s", indexStr, sanitizeFileName(element))
		filePath := key + "/" + elementName

		w, err := zipWriter.Create(filePath)
		if err != nil {
			return fmt.Errorf("创建 ZIP 文件失败: %w", err)
		}

		if _, err := w.Write([]byte(element)); err != nil {
			return fmt.Errorf("写入数据失败: %w", err)
		}
	}

	return nil
}

// sanitizeFileName 清理文件名中的非法字符
func sanitizeFileName(name string) string {
	// 替换文件名中的非法字符
	replacements := map[string]string{
		"/": "_",
		"\\": "_",
		":": "_",
		"*": "_",
		"?": "_",
		"\"": "_",
		"<": "_",
		">": "_",
		"|": "_",
		"\n": "_",
		"\r": "_",
		"\t": "_",
	}

	result := name
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// 限制文件名长度
	if len(result) > 200 {
		result = result[:200]
	}

	return result
}

// createMetadataFile 创建元数据文件
func (d *Dumper) createMetadataFile(key, keyType string, ttlSeconds int64, zipWriter *zip.Writer) error {
	// 构建元数据结构
	metadata := map[string]interface{}{
		"key":  key,
		"type": keyType,
		"ttl":  ttlSeconds, // -1 表示永久
	}

	// 转换为 JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("序列化元数据失败: %w", err)
	}

	// 创建元数据文件：key/.metadata.json
	metadataPath := key + "/.metadata.json"
	w, err := zipWriter.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("创建元数据文件失败: %w", err)
	}

	// 写入元数据
	_, err = w.Write(metadataJSON)
	if err != nil {
		return fmt.Errorf("写入元数据失败: %w", err)
	}

	return nil
}
