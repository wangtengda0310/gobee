package redis

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/service"
)

// Importer Redis 数据导入器
type Importer struct {
	mgr *service.RedisManager
}

// NewImporter 创建 Redis 导入器
func NewImporter(mgr *service.RedisManager) *Importer {
	return &Importer{mgr: mgr}
}

// ImportResult 导入结果
type ImportResult struct {
	KeysImported int
	FileName     string
}

// Import 从 ZIP 文件导入 Redis 数据
func (imp *Importer) Import(ctx context.Context, zipPath string, db int) (*ImportResult, error) {
	client := imp.mgr.GetClient()

	// 打开 ZIP 文件
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("打开 ZIP 文件失败: %w", err)
	}
	defer zipReader.Close()

	log.Printf("开始导入: %s", zipPath)

	// 按键分组收集数据
	keyGroups := make(map[string][]zipFileEntry)
	keyMetadatas := make(map[string]keyMetadata) // 存储每个 key 的元数据
	for _, file := range zipReader.File {
		// 跳过目录
		if file.FileInfo().IsDir() {
			continue
		}

		fileName := filepath.Base(file.Name)

		// 跳过元数据文件
		if fileName == ".metadata.json" {
			// 读取元数据
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("打开元数据文件 %s 失败: %w", file.Name, err)
			}

			content, err := io.ReadAll(rc)
			rc.Close()

			if err != nil {
				return nil, fmt.Errorf("读取元数据文件失败: %w", err)
			}

			// 解析元数据
			var metadata keyMetadata
			if err := json.Unmarshal(content, &metadata); err != nil {
				return nil, fmt.Errorf("解析元数据失败: %w", err)
			}

			// 从路径中提取 key
			key := filepath.Dir(file.Name)
			keyMetadatas[key] = metadata
			continue
		}

		// 解析路径：key/value (String), key/field (Hash), key/score_member (ZSET), key/member (Set)
		key, entryType, entryName := parseZipFilePath(file.Name)

		// 将文件内容读入内存
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("打开文件 %s 失败: %w", file.Name, err)
		}

		content, err := io.ReadAll(rc)
		rc.Close()

		if err != nil {
			return nil, fmt.Errorf("读取文件 %s 失败: %w", file.Name, err)
		}

		keyGroups[key] = append(keyGroups[key], zipFileEntry{
			path:      file.Name,
			entryType: entryType,
			entryName: entryName,
			content:   content, // 直接使用 []byte
		})
	}

	// 导入每个 key
	keysImported := 0
	totalKeys := len(keyGroups)

	// 创建进度条
	progress := NewProgressManager(int64(totalKeys), "导入进度")
	defer progress.Finish()

	// 在安静模式下输出到 stderr，避免干扰进度条
	if progress.IsQuiet() || !progress.IsTTY() {
		log.Printf("开始导入 %d 个 key", totalKeys)
	}

	currentKey := 0
	for key, entries := range keyGroups {
		// 获取元数据
		metadata, hasMetadata := keyMetadatas[key]

		if err := imp.importKey(ctx, client, key, entries, db, metadata); err != nil {
			progress.Close()
			return nil, fmt.Errorf("导入 key %s 失败: %w", key, err)
		}

		// 恢复 TTL
		if hasMetadata && metadata.TTL >= 0 {
			// TTL >= 0 表示有过期时间
			if err := client.Expire(ctx, key, time.Duration(metadata.TTL)*time.Second).Err(); err != nil {
				return nil, fmt.Errorf("恢复 TTL 失败: %w", err)
			}
			// 只在安静模式或非终端环境输出 TTL 恢复日志
			if progress.IsQuiet() || !progress.IsTTY() {
				log.Printf("  恢复 TTL: %d秒", metadata.TTL)
			}
		}

		keysImported++
		currentKey++
		progress.AddOne()

		// 在安静模式下每 100 个 key 输出一次日志
		if progress.IsQuiet() && currentKey%100 == 0 {
			log.Printf("已导入 %d/%d 个 key", currentKey, totalKeys)
		}
	}

	// 完成进度条
	progress.Close()

	if progress.IsQuiet() || !progress.IsTTY() {
		log.Printf("成功导入 %d 个 key", keysImported)
	}

	return &ImportResult{
		KeysImported: keysImported,
		FileName:     zipPath,
	}, nil
}

// zipFileEntry ZIP 文件条目
type zipFileEntry struct {
	path      string
	entryType string // "string", "hash", "zset", "set"
	entryName string // 字段名、成员名等
	content   []byte // 使用 []byte 支持二进制数据
}

// keyMetadata Key 元数据
type keyMetadata struct {
	Key  string `json:"key"`
	Type string `json:"type"`
	TTL  int64  `json:"ttl"` // -1 表示永久
}

// parseZipFilePath 解析 ZIP 文件路径
// 例如：
// - session:user:10001/value -> key=session:user:10001, type=string, name=value
// - config:app/version/value -> key=config:app/version, type=string, name=value
// - inventory:user:10001/gold -> key=inventory:user:10001, type=hash, name=gold
// - leaderboard:level/000099_user_10001 -> key=leaderboard:level, type=zset, name=user:10001,score=99
// - friends:user:10001/user_10002 -> key=friends:user:10001, type=set, name=user:10002
func parseZipFilePath(path string) (key, entryType, entryName string) {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", "", ""
	}

	// String 类型：key/value
	// 注意：key 本身可能包含 /，所以如果最后一个部分是 value，
	// 则 key 是前面所有部分的连接
	if parts[len(parts)-1] == "value" {
		key = strings.Join(parts[:len(parts)-1], "/")
		return key, "string", "value"
	}

	// 对于其他类型，key 是第一部分
	// 这是基于当前 dump 格式的假设
	// 如果 key 包含 /，这里需要调整
	key = parts[0]

	// 其他类型：key/field_name 或 key/score_member 或 key/member_name
	if len(parts) >= 2 {
		lastPart := parts[len(parts)-1]

		// 判断是否是 ZSET 格式：000099_user_10001
		if isZSetFileName(lastPart) {
			// 从文件名中提取 member 名称
			// 格式：000099_user_10001 -> member=user:10001, score=99
			member, _ := parseZSetFileName(lastPart)
			return key, "zset", member
		}

		// Hash 类型：key/field_name
		// Set 类型：key/member_name
		return key, detectTypeFromEntries(lastPart), lastPart
	}

	return key, "unknown", ""
}

// isZSetFileName 判断是否是 ZSET 文件名格式（数字开头）
func isZSetFileName(name string) bool {
	matched, _ := regexp.MatchString(`^\d{6}_`, name)
	return matched
}

// isListFileName 判断是否是 List 文件名格式（连续的 6 位数字索引）
// List 文件名格式：000000, 000001, 000002, ...（连续索引）
// ZSET 文件名格式：000099_user_10001（非连续的分数）
//
// 检测策略：检查所有文件名是否形成从 0 开始的连续序列
func isListFileName(entries []zipFileEntry) bool {
	if len(entries) == 0 {
		return false
	}

	// 收集所有索引
	indices := make(map[int]bool)
	for _, entry := range entries {
		fileName := filepath.Base(entry.path)
		// 检查是否是 6 位数字开头
		matched, _ := regexp.MatchString(`^\d{6}`, fileName)
		if !matched {
			return false
		}

		// 提取索引
		indexStr := fileName[:6]
		index, _ := strconv.ParseInt(indexStr, 10, 64)
		indices[int(index)] = true
	}

	// 检查是否是从 0 开始的连续索引
	for i := 0; i < len(indices); i++ {
		if !indices[i] {
			return false
		}
	}

	return true
}

// parseZSetFileName 解析 ZSET 文件名
// 例如：000099_user_10001 -> member=user:10001, score=99
func parseZSetFileName(name string) (member string, score int64) {
	parts := strings.SplitN(name, "_", 2)
	if len(parts) < 2 {
		return "", 0
	}

	scoreStr := strings.TrimLeft(parts[0], "0")
	if scoreStr == "" {
		scoreStr = "0"
	}

	score, _ = strconv.ParseInt(scoreStr, 10, 64)
	member = unsanitizeFileName(parts[1])

	return member, score
}

// detectTypeFromEntries 从条目名称检测类型
// 这里需要根据上下文判断是 Hash 还是 Set
// 暂时统一返回 "hash"，后续可根据实际情况调整
func detectTypeFromEntries(name string) string {
	// 对于 Hash 和 Set，在导入时需要根据 key 的实际类型来判断
	// 这里返回 "unknown"，让 importKey 处理
	return "unknown"
}

// unsanitizeFileName 反向清理文件名
// 将 sanitizeFileName 替换的字符还原
func unsanitizeFileName(name string) string {
	// 将 _ 替换回可能的特殊字符
	// 注意：这里无法完全还原，因为 _ 是多对一映射
	// 只能尽量还原常见的特殊字符

	// 对于 user_10002 这样的格式，尝试只替换最后一个 _ 为 :
	// 这样 user_10002 -> user:10002
	lastUnderscore := strings.LastIndex(name, "_")
	if lastUnderscore > 0 {
		// 检查是否是数字格式（如 10002）
		afterPart := name[lastUnderscore+1:]
		if _, err := strconv.ParseInt(afterPart, 10, 64); err == nil {
			// 如果后面部分是数字，将 _ 替换为 :
			return name[:lastUnderscore] + ":" + afterPart
		}
	}

	// 默认：将所有 _ 替换为 :
	return strings.ReplaceAll(name, "_", ":")
}

// importKey 导入单个 key
func (imp *Importer) importKey(ctx context.Context, client *redis.Client, key string, entries []zipFileEntry, db int, metadata keyMetadata) error {
	_ = db // 暂时未使用
	_ = metadata // 暂时未使用，TTL 恢复在调用方处理

	if len(entries) == 0 {
		return nil
	}

	// 确定数据类型
	entryType := imp.detectDataType(entries)

	log.Printf("导入 key: %s (类型: %s)", key, entryType)

	// 删除已存在的 key
	client.Del(ctx, key)

	// 根据类型导入
	switch entryType {
	case "string":
		return imp.importString(ctx, client, key, entries)
	case "hash":
		return imp.importHash(ctx, client, key, entries)
	case "list":
		return imp.importList(ctx, client, key, entries)
	case "zset":
		return imp.importZSet(ctx, client, key, entries)
	case "set":
		return imp.importSet(ctx, client, key, entries)
	default:
		return fmt.Errorf("未知的数据类型: %s", entryType)
	}
}

// detectDataType 检测数据类型
func (imp *Importer) detectDataType(entries []zipFileEntry) string {
	if len(entries) == 0 {
		return "unknown"
	}

	// 只有一个条目且是 value -> String
	if len(entries) == 1 && entries[0].entryType == "string" {
		return "string"
	}

	// 检查是否是 List 类型（连续索引从 0 开始）
	// 需要在 ZSET 之前检查，因为 ZSET 也有 6 位数字前缀
	if isListFileName(entries) {
		return "list"
	}

	// 检查是否是 ZSET 文件名格式（6 位数字 + 下划线 + 内容）
	fileName := filepath.Base(entries[0].path)
	if isZSetFileName(fileName) {
		return "zset"
	}

	// 检查是否所有条目的内容都等于文件名（经过反向清理）-> Set
	// Set 的特点是：文件名 = sanitizeFileName(member)，内容 = member
	// 如果内容与文件名（反向清理后）相同，则判定为 Set
	allSet := true
	for _, e := range entries {
		// 反向清理文件名
		unsanitized := unsanitizeFileName(filepath.Base(e.path))
		// 将 content ([]byte) 转换为 string 进行比较
		contentStr := string(e.content)
		// 如果内容与反向清理后的文件名不同，则不是 Set
		if contentStr != unsanitized {
			allSet = false
			break
		}
	}
	if allSet && len(entries) > 1 {
		return "set"
	}

	// 否则，根据条目数量判断
	// 单个条目 -> String
	// 多个条目 -> Hash
	if len(entries) == 1 {
		return "string"
	}

	// 默认返回 Hash
	return "hash"
}

// importString 导入 String 类型
func (imp *Importer) importString(ctx context.Context, client *redis.Client, key string, entries []zipFileEntry) error {
	if len(entries) != 1 {
		return fmt.Errorf("String 类型应该只有一个条目，但找到 %d 个", len(entries))
	}

	value := entries[0].content // []byte
	// Set() 直接接受 []byte，会正确处理二进制数据
	return client.Set(ctx, key, value, 0).Err()
}

// importHash 导入 Hash 类型
func (imp *Importer) importHash(ctx context.Context, client *redis.Client, key string, entries []zipFileEntry) error {
	for _, entry := range entries {
		// 提取字段名（最后一部分是字段名）
		fieldName := filepath.Base(entry.path)
		// HSet() 接受 []byte，会正确处理二进制字段值
		if err := client.HSet(ctx, key, fieldName, entry.content).Err(); err != nil {
			return err
		}
	}
	return nil
}

// importZSet 导入 ZSET 类型
func (imp *Importer) importZSet(ctx context.Context, client *redis.Client, key string, entries []zipFileEntry) error {
	for _, entry := range entries {
		// 从路径中提取文件名
		fileName := filepath.Base(entry.path)
		_, score := parseZSetFileName(fileName)

		// 内容是 member 名称（可能是二进制）
		memberName := entry.content

		// 使用 ZAdd 添加成员
		if err := client.ZAdd(ctx, key, redis.Z{
			Score:  float64(score),
			Member: memberName,
		}).Err(); err != nil {
			return err
		}
	}
	return nil
}

// importSet 导入 Set 类型
func (imp *Importer) importSet(ctx context.Context, client *redis.Client, key string, entries []zipFileEntry) error {
	for _, entry := range entries {
		// 内容是 member 名称（可能是二进制）
		member := entry.content
		// SAdd() 接受 []byte，会正确处理二进制成员
		if err := client.SAdd(ctx, key, member).Err(); err != nil {
			return err
		}
	}
	return nil
}

// importList 导入 List 类型
func (imp *Importer) importList(ctx context.Context, client *redis.Client, key string, entries []zipFileEntry) error {
	// List 文件名格式：000000 或 000000_content
	// 需要按索引排序后，使用 RPUSH 按顺序导入
	// RPUSH 保证：RPUSH a,b,c -> List [a,b,c]

	// 按索引排序条目
	sortedEntries := make([]zipFileEntry, len(entries))
	copy(sortedEntries, entries)

	// 简单的冒泡排序，按文件名的索引部分排序
	for i := 0; i < len(sortedEntries)-1; i++ {
		for j := i + 1; j < len(sortedEntries); j++ {
			fileNameI := filepath.Base(sortedEntries[i].path)
			fileNameJ := filepath.Base(sortedEntries[j].path)
			// 提取索引（前 6 位）
			indexI := fileNameI[:6]
			indexJ := fileNameJ[:6]
			if indexI > indexJ {
				sortedEntries[i], sortedEntries[j] = sortedEntries[j], sortedEntries[i]
			}
		}
	}

	// 按顺序 RPUSH
	for _, entry := range sortedEntries {
		// 内容是 List 元素（可能是二进制）
		if err := client.RPush(ctx, key, entry.content).Err(); err != nil {
			return fmt.Errorf("导入 List 元素失败: %w", err)
		}
	}
	return nil
}
