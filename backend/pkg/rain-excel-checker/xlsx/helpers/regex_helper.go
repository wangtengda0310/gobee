package helpers

import (
	"regexp"
	"strconv"
	"strings"
)

// ExtractValuesByRegex 使用正则从字符串中提取所有匹配的捕获组值
// groups 指定要提取的捕获组索引（如 []int{1} 提取第 1 组）
// groups 为 nil 时默认提取整个匹配
// 无匹配时返回 nil
func ExtractValuesByRegex(value string, pattern *regexp.Regexp, groups []int) []string {
	if value == "" || pattern == nil {
		return nil
	}

	matches := pattern.FindAllStringSubmatch(value, -1)
	if matches == nil {
		return nil
	}

	var result []string
	for _, match := range matches {
		if len(groups) > 0 {
			for _, groupIdx := range groups {
				if groupIdx < len(match) && match[groupIdx] != "" {
					result = append(result, match[groupIdx])
				}
			}
		} else {
			if len(match) > 0 && match[0] != "" {
				result = append(result, match[0])
			}
		}
	}

	return result
}

// ParseCaptureGroups 解析逗号分隔的捕获组索引字符串
// 空字符串返回默认值 []int{1}
func ParseCaptureGroups(groupsStr string) []int {
	if groupsStr == "" {
		return []int{1}
	}

	var groups []int
	for _, s := range strings.Split(groupsStr, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if idx, err := strconv.Atoi(s); err == nil && idx >= 0 {
			groups = append(groups, idx)
		}
	}

	if len(groups) == 0 {
		return []int{1}
	}
	return groups
}
