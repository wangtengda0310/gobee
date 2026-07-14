// Package check_manager 提供缓存补充逻辑的单元测试
package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestSupplementFromCache_Basic 测试从缓存补充缺失的 sheet
// 场景：sheetMap 有 2 个 sheet，cache 有 3 个 sheet，其中 1 个是新增的
func TestSupplementFromCache_Basic(t *testing.T) {
	// 准备数据
	sheetMap := make(map[string]*excelize.File)
	sheetMap["SheetA"] = &excelize.File{}
	sheetMap["SheetB"] = &excelize.File{}

	cache := make(map[string]*excelize.File)
	cache["SheetA"] = &excelize.File{} // 已存在，不应重复补充
	cache["SheetB"] = &excelize.File{} // 已存在，不应重复补充
	cache["SheetC"] = &excelize.File{} // 新增，应被补充

	count := supplementFromCache(sheetMap, cache, nil)

	assert.Equal(t, 1, count, "应补充 1 个缺失的 sheet")
	assert.Len(t, sheetMap, 3, "sheetMap 应有 3 个 sheet")
	_, hasC := sheetMap["SheetC"]
	assert.True(t, hasC, "SheetC 应被补充到 sheetMap 中")
}

// TestSupplementFromCache_EmptyCache 测试空缓存不会补充任何内容
func TestSupplementFromCache_EmptyCache(t *testing.T) {
	sheetMap := make(map[string]*excelize.File)
	sheetMap["SheetA"] = &excelize.File{}

	cache := make(map[string]*excelize.File)

	count := supplementFromCache(sheetMap, cache, nil)

	assert.Equal(t, 0, count, "空缓存不应补充任何内容")
	assert.Len(t, sheetMap, 1, "sheetMap 数量不变")
}

// TestSupplementFromCache_AllExisting 测试缓存中的所有 sheet 已存在于 sheetMap
func TestSupplementFromCache_AllExisting(t *testing.T) {
	sheetMap := make(map[string]*excelize.File)
	sheetMap["SheetA"] = &excelize.File{}
	sheetMap["SheetB"] = &excelize.File{}

	cache := make(map[string]*excelize.File)
	cache["SheetA"] = &excelize.File{}
	cache["SheetB"] = &excelize.File{}

	count := supplementFromCache(sheetMap, cache, nil)

	assert.Equal(t, 0, count, "所有 sheet 已存在，不应补充")
	assert.Len(t, sheetMap, 2, "sheetMap 数量不变")
}

// TestSupplementFromCache_EmptySheetMap 测试空 sheetMap 时从缓存全部补充
func TestSupplementFromCache_EmptySheetMap(t *testing.T) {
	sheetMap := make(map[string]*excelize.File)

	cache := make(map[string]*excelize.File)
	cache["SheetA"] = &excelize.File{}
	cache["SheetB"] = &excelize.File{}
	cache["SheetC"] = &excelize.File{}

	count := supplementFromCache(sheetMap, cache, nil)

	assert.Equal(t, 3, count, "空 sheetMap 应补充所有缓存中的 sheet")
	assert.Len(t, sheetMap, 3, "sheetMap 应有 3 个 sheet")
}

// TestSupplementSheetMap_WithCache 测试 supplementSheetMap 优先使用缓存路径
func TestSupplementSheetMap_WithCache(t *testing.T) {
	sheetMap := make(map[string]*excelize.File)
	sheetMap["SheetA"] = &excelize.File{}

	cache := make(map[string]*excelize.File)
	cache["SheetB"] = &excelize.File{}
	cache["SheetC"] = &excelize.File{}

	// cache 不为 nil，应走缓存路径
	supplementSheetMap(sheetMap, "", "", cache, nil)

	assert.Len(t, sheetMap, 3, "应通过缓存补充到 3 个 sheet")
}

// TestSupplementSheetMap_NilCache 测试 supplementSheetMap 缓存为 nil 时走 commit 路径
// 注意：commit 路径需要真实 git 仓库，这里只验证不会 panic
func TestSupplementSheetMap_NilCache(t *testing.T) {
	sheetMap := make(map[string]*excelize.File)
	sheetMap["SheetA"] = &excelize.File{}

	// cache 为 nil，会走 commit 路径，repoPath 和 commitHash 无效时会打印警告但不 panic
	supplementSheetMap(sheetMap, "/nonexistent", "0000000000000000000000000000000000000000", nil, nil)

	// sheetMap 不应有变化（commit 路径会因路径无效而跳过）
	assert.Len(t, sheetMap, 1, "无效路径不应补充任何 sheet")
}

// TestWithFallbackSheetMap 测试 CheckOption 正确设置 fallbackSheetMap
func TestWithFallbackSheetMap(t *testing.T) {
	cache := make(map[string]*excelize.File)
	cache["SheetA"] = &excelize.File{}

	options := &checkOptions{}
	opt := WithFallbackSheetMap(cache)
	opt(options)

	assert.NotNil(t, options.fallbackSheetMap, "fallbackSheetMap 应被设置")
	assert.Len(t, options.fallbackSheetMap, 1, "fallbackSheetMap 应有 1 个 sheet")
	assert.Contains(t, options.fallbackSheetMap, "SheetA", "fallbackSheetMap 应包含 SheetA")
}

// TestWithFallbackSheetMap_Nil 测试传入 nil 的 fallbackSheetMap
func TestWithFallbackSheetMap_Nil(t *testing.T) {
	options := &checkOptions{}
	opt := WithFallbackSheetMap(nil)
	opt(options)

	// map 的零值是 nil，所以 fallbackSheetMap 为 nil
	assert.Nil(t, options.fallbackSheetMap, "nil map 应保持 nil")
}

// TestCheckOptions_Default 测试默认 checkOptions 状态
func TestCheckOptions_Default(t *testing.T) {
	options := &checkOptions{}

	assert.Nil(t, options.fallbackSheetMap, "默认 fallbackSheetMap 应为 nil")
}

// ==================== supplementSheetMap 降级路径测试 ====================

// TestSupplementSheetMap_CacheMissFallthrough 缓存完全未命中时降级到 git 加载
// 场景：sheetMap 有 DropItem，缓存有 PetAudio（不含 Item），需要 Item
// 预期：缓存补充 0 个，降级到 supplementFromCommit（因无效路径实际不补充，但不 panic）
func TestSupplementSheetMap_CacheMissFallthrough(t *testing.T) {
	sheetMap := map[string]*excelize.File{
		"掉落道具表|DropItem": {},
	}
	cache := map[string]*excelize.File{
		"灵宠音效|PetAudio": {},
	}
	requiredSheets := map[string]bool{"Item": true}

	// 传入无效 git 路径，supplementFromCommit 会因路径无效而跳过，但不 panic
	supplementSheetMap(sheetMap, "/nonexistent", "0000000000000000000000000000000000000000", cache, requiredSheets)

	// 缓存中没有 Item，git 路径无效也不会加载，sheetMap 不变
	assert.Len(t, sheetMap, 1, "缓存未命中 + git 无效路径，sheetMap 不应变化")
	assert.Contains(t, sheetMap, "掉落道具表|DropItem", "原有 sheet 不应丢失")
}

// TestSupplementSheetMap_CachePartialHitFallthrough 缓存部分命中时降级加载剩余
// 场景：需要 DropItem 和 Item，缓存只有 DropItem
// 预期：DropItem 从缓存补充，Item 降级到 supplementFromCommit
func TestSupplementSheetMap_CachePartialHitFallthrough(t *testing.T) {
	sheetMap := map[string]*excelize.File{
		"武将|Hero": {},
	}
	cache := map[string]*excelize.File{
		"掉落道具表|DropItem": {},
	}
	requiredSheets := map[string]bool{"DropItem": true, "Item": true}

	supplementSheetMap(sheetMap, "/nonexistent", "0000000000000000000000000000000000000000", cache, requiredSheets)

	// DropItem 从缓存补充成功，Item 因 git 无效路径未能加载
	assert.Len(t, sheetMap, 2, "应从缓存补充 DropItem")
	assert.Contains(t, sheetMap, "掉落道具表|DropItem", "DropItem 应从缓存补充")
	assert.Contains(t, sheetMap, "武将|Hero", "原有 sheet 不应丢失")
}

// TestSupplementSheetMap_CacheFullHitNoFallback 缓存完全命中时不触发降级
// 场景：需要 Item 和 DropItem，缓存两者都有
// 预期：全部从缓存补充，不调用 supplementFromCommit
func TestSupplementSheetMap_CacheFullHitNoFallback(t *testing.T) {
	sheetMap := map[string]*excelize.File{
		"武将|Hero": {},
	}
	cache := map[string]*excelize.File{
		"道具表|Item":       {},
		"掉落道具表|DropItem": {},
	}
	requiredSheets := map[string]bool{"Item": true, "DropItem": true}

	supplementSheetMap(sheetMap, "/nonexistent", "0000000000000000000000000000000000000000", cache, requiredSheets)

	assert.Len(t, sheetMap, 3, "应从缓存补充 Item 和 DropItem")
	assert.Contains(t, sheetMap, "道具表|Item", "Item 应从缓存补充")
	assert.Contains(t, sheetMap, "掉落道具表|DropItem", "DropItem 应从缓存补充")
}

// TestSupplementSheetMap_EmptyCacheFallthrough 空缓存（非 nil）时降级到 git 加载
// 场景：缓存是空 map（非 nil），需要 Item
// 预期：supplementFromCache 返回 0，降级到 supplementFromCommit
func TestSupplementSheetMap_EmptyCacheFallthrough(t *testing.T) {
	sheetMap := map[string]*excelize.File{
		"掉落道具表|DropItem": {},
	}
	cache := map[string]*excelize.File{} // 空 map，非 nil
	requiredSheets := map[string]bool{"Item": true}

	supplementSheetMap(sheetMap, "/nonexistent", "0000000000000000000000000000000000000000", cache, requiredSheets)

	assert.Len(t, sheetMap, 1, "空缓存 + git 无效路径，sheetMap 不应变化")
}

// TestSupplementSheetMap_AllNeededExist_NoSupplement 所有需要的表已存在时直接返回
// 注意：needed 计算使用精确匹配 sheetMap[sheet]，所以 requiredSheets 的 key 必须与 sheetMap 的 key 完全一致
// 场景：requiredSheets 要求 Item 和 Hero，sheetMap 中以精确 key 存在
// 预期：needed 为空，不触发缓存补充也不降级
func TestSupplementSheetMap_AllNeededExist_NoSupplement(t *testing.T) {
	sheetMap := map[string]*excelize.File{
		"Item": {},
		"Hero": {},
	}
	cache := map[string]*excelize.File{
		"PetAudio": {}, // 不在 requiredSheets 中，不应被补充
	}
	requiredSheets := map[string]bool{"Item": true, "Hero": true}

	supplementSheetMap(sheetMap, "/nonexistent", "0000000000000000000000000000000000000000", cache, requiredSheets)

	// needed 为空直接返回（第476行），缓存中的 PetAudio 不应被补充
	assert.Len(t, sheetMap, 2, "所有 needed 已存在，sheetMap 不应变化")
	assert.NotContains(t, sheetMap, "PetAudio", "PetAudio 不在 needed 中，不应被补充")
}
