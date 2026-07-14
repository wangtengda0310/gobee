// Package reference 提供引用关系相关的校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package reference

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// RegexCheckRule 正则表达式检查规则
type RegexCheckRule struct{}

func (c *RegexCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 获取正则表达式模式
	patternStr, ok := params[string(json_rule.PATTERN)]
	if !ok || patternStr == "" {
		return []*json_rule.CellError{
			{
				Index:  0,
				Reason: "必须提供正则表达式参数 'pattern'",
			},
		}
	}

	// 编译正则表达式
	pattern, err := regexp.Compile(patternStr)
	if err != nil {
		return []*json_rule.CellError{
			{
				Index:  0,
				Reason: fmt.Sprintf("正则表达式编译错误: %v", err),
			},
		}
	}

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

	// 是否要求完全匹配（默认true）
	fullMatch := true
	if full, ok := params["fullMatch"]; ok {
		fullMatch = strings.ToLower(full) == "true"
	}

	// 错误信息模板
	errorTemplate := "格式不正确"
	if msg, ok := params["errorMessage"]; ok && msg != "" {
		errorTemplate = msg
	}

	// 描述信息（用于错误提示）
	description := ""
	if desc, ok := params["description"]; ok && desc != "" {
		description = desc
	}

	// 获取提取组信息（用于验证提取的内容）
	extractGroups := make([]string, 0)
	if groupsStr, ok := params[string(json_rule.GROUPS)]; ok && groupsStr != "" {
		extractGroups = strings.Split(groupsStr, ",")
		for i := range extractGroups {
			extractGroups[i] = strings.TrimSpace(extractGroups[i])
		}
	}

	// 获取验证函数（可选）
	var validator func([]string) (bool, string)
	if validatorName, ok := params["validator"]; ok && validatorName != "" {
		validator = getValidator(validatorName, params)
	}

	for i, s := range myColData {
		// 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

		// 检查字符串
		var matched bool
		if fullMatch {
			matched = pattern.MatchString(s)
		} else {
			// 部分匹配：查找字符串中是否有匹配的部分
			matched = pattern.FindString(s) != ""
		}

		if !matched {
			reason := errorTemplate
			if description != "" {
				reason = fmt.Sprintf("%s: %s (应为: %s)", errorTemplate, s, description)
			} else {
				reason = fmt.Sprintf("%s: %s", errorTemplate, s)
			}
			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   reason,
			})
			continue
		}

		// 如果匹配成功，进行进一步的验证
		if validator != nil || len(extractGroups) > 0 {
			// 提取匹配的组
			matches := pattern.FindAllStringSubmatch(s, -1)
			if matches != nil && len(matches) > 0 {
				// 对每个匹配项进行验证
				for matchIdx, match := range matches {
					// 提取需要的组
					groups := make([]string, 0)
					if len(extractGroups) > 0 {
						for _, groupIdxStr := range extractGroups {
							if idx, err := strconv.Atoi(groupIdxStr); err == nil && idx < len(match) {
								groups = append(groups, match[idx])
							}
						}
					} else {
						// 默认提取所有组（除了第0组是整个匹配）
						if len(match) > 1 {
							groups = match[1:]
						}
					}

					// 调用验证函数
					if validator != nil {
						if valid, errMsg := validator(groups); !valid {
							reason := fmt.Sprintf("格式验证失败: %s", s)
							if errMsg != "" {
								reason = fmt.Sprintf("%s: %s", errMsg, s)
							}
							if matchIdx > 0 {
								reason = fmt.Sprintf("%s (第%d个匹配项)", reason, matchIdx+1)
							}
							res = append(res, &json_rule.CellError{
								Index:    i,
								ExcelRow: startRowIdx + i + 1,
								Reason:   reason,
							})
							break
						}
					}

					// 内置验证：检查提取的组是否为数值
					if validateNumeric, ok := params["validateNumeric"]; ok && strings.ToLower(validateNumeric) == "true" {
						for _, group := range groups {
							if group != "" {
								if _, err := strconv.ParseFloat(group, 64); err != nil {
									res = append(res, &json_rule.CellError{
										Index:    i,
										ExcelRow: startRowIdx + i + 1,
										Reason:   fmt.Sprintf("提取的值不是有效数值: %s (在: %s)", group, s),
									})
									break
								}
							}
						}
					}

					// 内置验证：检查提取的组是否为整数
					if validateInteger, ok := params["validateInteger"]; ok && strings.ToLower(validateInteger) == "true" {
						for _, group := range groups {
							if group != "" {
								if _, err := strconv.ParseInt(group, 10, 64); err != nil {
									res = append(res, &json_rule.CellError{
										Index:    i,
										ExcelRow: startRowIdx + i + 1,
										Reason:   fmt.Sprintf("提取的值不是有效整数: %s (在: %s)", group, s),
									})
									break
								}
							}
						}
					}

					// 内置验证：检查提取的组是否在范围内
					if minStr, ok := params["minValue"]; ok {
						if min_, err := strconv.ParseFloat(minStr, 64); err == nil {
							for _, group := range groups {
								if group != "" {
									if val, err := strconv.ParseFloat(group, 64); err == nil && val < min_ {
										res = append(res, &json_rule.CellError{
											Index:    i,
											ExcelRow: startRowIdx + i + 1,
											Reason:   fmt.Sprintf("值%s小于最小值%s: %s", group, minStr, s),
										})
										break
									}
								}
							}
						}
					}

					if maxStr, ok := params["maxValue"]; ok {
						if max_, err := strconv.ParseFloat(maxStr, 64); err == nil {
							for _, group := range groups {
								if group != "" {
									if val, err := strconv.ParseFloat(group, 64); err == nil && val > max_ {
										res = append(res, &json_rule.CellError{
											Index:    i,
											ExcelRow: startRowIdx + i + 1,
											Reason:   fmt.Sprintf("值%s大于最大值%s: %s", group, maxStr, s),
										})
										break
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return res
}

// 获取验证函数
func getValidator(name string, params map[string]string) func([]string) (bool, string) {
	switch name {
	case "itemFormat":
		return func(groups []string) (bool, string) {
			if len(groups) < 2 {
				return false, "道具格式需要ID和数量两个参数"
			}

			// 检查ID是否为有效数值
			if _, err := strconv.ParseInt(groups[0], 10, 64); err != nil {
				return false, fmt.Sprintf("道具ID不是有效数值: %s", groups[0])
			}

			// 检查数量是否为有效数值且大于0
			if count, err := strconv.ParseInt(groups[1], 10, 64); err != nil {
				return false, fmt.Sprintf("道具数量不是有效数值: %s", groups[1])
			} else if count <= 0 {
				return false, fmt.Sprintf("道具数量必须大于0: %d", count)
			}

			return true, ""
		}

	case "email":
		return func(groups []string) (bool, string) {
			if len(groups) == 0 {
				return false, "没有提供邮箱地址"
			}
			// 简单的邮箱格式检查
			emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
			if matched, _ := regexp.MatchString(emailPattern, groups[0]); !matched {
				return false, "不是有效的邮箱格式"
			}
			return true, ""
		}

	case "phone":
		return func(groups []string) (bool, string) {
			if len(groups) == 0 {
				return false, "没有提供电话号码"
			}
			// 简单的手机号格式检查（中国）
			phonePattern := `^1[3-9]\d{9}$`
			if matched, _ := regexp.MatchString(phonePattern, groups[0]); !matched {
				return false, "不是有效的手机号码"
			}
			return true, ""
		}

	case "url":
		return func(groups []string) (bool, string) {
			if len(groups) == 0 {
				return false, "没有提供URL"
			}
			// 简单的URL格式检查
			urlPattern := `^(https?|ftp)://[^\s/$.?#].[^\s]*$`
			if matched, _ := regexp.MatchString(urlPattern, groups[0]); !matched {
				return false, "不是有效的URL格式"
			}
			return true, ""
		}

	default:
		return nil
	}
}
