package helpers

import (
	"path/filepath"
)

// PathNormalizer 路径规范化工具
// 统一处理相对路径、绝对路径、不同分隔符的路径比较问题
type PathNormalizer struct {
	baseDir string // 基准目录，用于解析相对路径
}

// NewPathNormalizer 创建路径规范化器
func NewPathNormalizer(baseDir string) *PathNormalizer {
	// 确保 baseDir 是绝对路径
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		absBase = baseDir
	}
	return &PathNormalizer{baseDir: absBase}
}

// Normalize 将路径统一转换为：绝对路径 + 正斜杠格式
// 例如：../../config/excel/Hero.xlsx -> D:/work/config/excel/Hero.xlsx
func (n *PathNormalizer) Normalize(path string) string {
	var absPath string

	// 如果是相对路径，先转换为绝对路径
	if !filepath.IsAbs(path) {
		// 尝试基于 baseDir 解析相对路径
		absPath = filepath.Join(n.baseDir, path)
		// 如果结果仍然不是绝对路径，使用 filepath.Abs
		if !filepath.IsAbs(absPath) {
			var err error
			absPath, err = filepath.Abs(path)
			if err != nil {
				absPath = path
			}
		}
	} else {
		absPath = path
	}

	// 统一分隔符为正斜杠，并清理路径
	normalized := filepath.ToSlash(filepath.Clean(absPath))
	return normalized
}

// Equal 比较两个路径是否等价（规范化后比较）
func (n *PathNormalizer) Equal(a, b string) bool {
	return n.Normalize(a) == n.Normalize(b)
}

// NormalizeMap 批量规范化 map 中的所有路径
func (n *PathNormalizer) NormalizeMap(paths map[string]bool) map[string]bool {
	result := make(map[string]bool, len(paths))
	for path, value := range paths {
		result[n.Normalize(path)] = value
	}
	return result
}
