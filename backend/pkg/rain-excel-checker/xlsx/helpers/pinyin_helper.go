package helpers

import (
	"strings"

	"github.com/mozillazg/go-pinyin"
)

// PinyinFormat 拼音输出格式
type PinyinFormat string

const (
	PinyinCamel PinyinFormat = "camel" // 驼峰：ZhangFei（默认）
	PinyinLower PinyinFormat = "lower" // 全小写：zhangfei
	PinyinSnake PinyinFormat = "snake" // 下划线分隔：zhang_fei
)

// ConvertToPinyin 将中文字符串转换为拼音（驼峰格式，向后兼容）
// 非中文字符保持不变
func ConvertToPinyin(chinese string) string {
	return ConvertToPinyinWithFormat(chinese, PinyinCamel)
}

// ConvertToPinyinWithFormat 将中文字符串转换为指定格式的拼音
// format: "camel"(ZhangFei), "lower"(zhangfei), "snake"(zhang_fei)
// 非中文字符保持不变
func ConvertToPinyinWithFormat(chinese string, format PinyinFormat) string {
	if chinese == "" {
		return ""
	}

	args := pinyin.NewArgs()
	args.Style = pinyin.Normal

	pinyinSlice := pinyin.Pinyin(chinese, args)

	var result []string
	for _, wordPinyin := range pinyinSlice {
		if len(wordPinyin) > 0 {
			p := wordPinyin[0]
			if len(p) > 0 {
				switch format {
				case PinyinLower, PinyinSnake:
					result = append(result, strings.ToLower(p))
				default:
					result = append(result, strings.Title(p))
				}
			}
		}
	}

	var converted string
	switch format {
	case PinyinSnake:
		converted = strings.Join(result, "_")
	default:
		converted = strings.Join(result, "")
	}

	if converted == "" {
		return chinese
	}
	return converted
}

// GetPinyinVariants 获取中文字符串的所有可能拼音组合（处理多音字）
// 返回所有可能的拼音变体列表，用于多音字匹配
// 例如 "长大" 返回 ["ZhangDa", "ChangDa"]（camel格式）
func GetPinyinVariants(chinese string, format PinyinFormat) []string {
	if chinese == "" {
		return []string{""}
	}

	args := pinyin.NewArgs()
	args.Style = pinyin.Normal
	args.Heteronym = true // 开启多音字模式，返回所有读音

	pinyinSlice := pinyin.Pinyin(chinese, args)

	// 每个字可能有多个拼音，生成所有组合
	// 使用递归/迭代方式生成笛卡尔积
	variants := [][]string{{}}

	for _, wordPinyins := range pinyinSlice {
		if len(wordPinyins) == 0 {
			continue
		}

		// 格式化每个拼音变体
		formatted := make([]string, 0, len(wordPinyins))
		for _, p := range wordPinyins {
			if p == "" {
				continue
			}
			switch format {
			case PinyinLower, PinyinSnake:
				formatted = append(formatted, strings.ToLower(p))
			default:
				formatted = append(formatted, strings.Title(p))
			}
		}

		if len(formatted) == 0 {
			continue
		}

		// 笛卡尔积：每个现有变体与当前字的所有拼音组合
		newVariants := make([][]string, 0, len(variants)*len(formatted))
		for _, v := range variants {
			for _, fp := range formatted {
				newV := append([]string{}, v...)
				newV = append(newV, fp)
				newVariants = append(newVariants, newV)
			}
		}
		variants = newVariants
	}

	if len(variants) == 0 || (len(variants) == 1 && len(variants[0]) == 0) {
		return []string{chinese}
	}

	// 根据格式拼接
	results := make([]string, 0, len(variants))
	for _, v := range variants {
		var s string
		switch format {
		case PinyinSnake:
			s = strings.Join(v, "_")
		default:
			s = strings.Join(v, "")
		}
		if s != "" {
			results = append(results, s)
		}
	}

	if len(results) == 0 {
		return []string{chinese}
	}
	return results
}
