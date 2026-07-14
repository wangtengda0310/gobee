// Package helpers 提供校验规则的内部辅助工具
// 本包包含列检查、参数解析、表查找等通用辅助函数
package helpers

import (
	"fmt"
	"strconv"
	"strings"
)

// ==================== 参数解析工具 ====================

// ParseIntParamWithError 解析整数参数字符串，返回 error
func ParseIntParamWithError(s string) (int, error) {
	return ParseIntWithError(s)
}

// ParseIntParam 解析整数参数，如果为空则返回默认值
func ParseIntParam(params map[string]string, key string, defaultValue int) int {
	if val, ok := params[key]; ok && val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// ParseIntWithError 解析整数，支持十六进制（0x前缀），返回 error
func ParseIntWithError(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("空字符串")
	}
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		val := 0
		n, _ := fmt.Sscanf(s, "0x%x", &val)
		if n != 1 {
			return 0, fmt.Errorf("无法解析十六进制: %s", s)
		}
		return val, nil
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ParseInt 解析整数，支持十六进制（0x前缀），忽略错误
func ParseInt(s string) int {
	val, _ := ParseIntWithError(s)
	return val
}

// ParseBool 解析布尔值
func ParseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes"
}

// ParseSkillList 解析技能列表（逗号分隔）
func ParseSkillList(s string) []int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		if id := ParseInt(part); id > 0 {
			result = append(result, id)
		}
	}
	return result
}
