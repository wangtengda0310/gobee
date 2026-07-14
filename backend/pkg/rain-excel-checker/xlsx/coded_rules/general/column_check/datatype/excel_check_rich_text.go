// Package datatype 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package datatype

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// RichTextCheckRule 富文本检查规则
type RichTextCheckRule struct{}

func (c *RichTextCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

	for i, s := range myColData {
		// 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

		// 执行富文本检查
		errors := c.checkRichText(s, i)
		res = append(res, errors...)
	}
	return res
}

// 检查富文本格式
func (c *RichTextCheckRule) checkRichText(text string, rowIdx int) []*json_rule.CellError {
	var errors []*json_rule.CellError

	// 1. 标签闭合检查
	if errs := c.checkTagClosure(text, rowIdx); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	// 2. 标签格式检查
	if errs := c.checkTagFormat(text, rowIdx); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	// 3. 占位符检查
	if errs := c.checkPlaceholders(text, rowIdx); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	// 4. 允许的字符检查
	if errs := c.checkAllowedChars(text, rowIdx); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	return errors
}

// 检查标签闭合 - 修复版本
func (c *RichTextCheckRule) checkTagClosure(text string, rowIdx int) []*json_rule.CellError {
	var errors []*json_rule.CellError

	// 只提取真正的标签，忽略占位符
	tagRegex := regexp.MustCompile(`<(\/?.+?)>`)

	// 栈用于检查标签嵌套
	tagStack := []string{}
	lineIndexes := []int{}

	// 找出所有标签的位置并按顺序处理
	allTags := tagRegex.FindAllStringSubmatch(text, -1)
	tagPositions := tagRegex.FindAllStringIndex(text, -1)

	for i, match := range allTags {
		if len(match) < 2 {
			continue
		}

		fullTag := match[0]
		tagContent := match[1]

		// 跳过占位符 {0} 等，它们不是标签
		if strings.HasPrefix(tagContent, "0") || strings.HasPrefix(tagContent, "1") ||
			strings.HasPrefix(tagContent, "2") || strings.HasPrefix(tagContent, "3") ||
			strings.HasPrefix(tagContent, "4") || strings.HasPrefix(tagContent, "5") ||
			strings.HasPrefix(tagContent, "6") || strings.HasPrefix(tagContent, "7") ||
			strings.HasPrefix(tagContent, "8") || strings.HasPrefix(tagContent, "9") {
			// 这是 {0} 这样的占位符，不是标签
			continue
		}

		pos := tagPositions[i][0]

		// 判断是开始标签还是结束标签
		if strings.HasPrefix(tagContent, "/") {
			// 结束标签
			tagName := strings.TrimPrefix(tagContent, "/")
			// 提取纯标签名（去掉属性）
			if spaceIdx := strings.IndexAny(tagName, " ="); spaceIdx > 0 {
				tagName = tagName[:spaceIdx]
			}

			if len(tagStack) == 0 {
				errors = append(errors, &json_rule.CellError{
					Index:  rowIdx,
					Reason: fmt.Sprintf("多余的结束标签: %s", fullTag),
				})
				continue
			}

			// 检查是否匹配栈顶标签
			lastTag := tagStack[len(tagStack)-1]
			if lastTag == tagName {
				// 匹配，出栈
				tagStack = tagStack[:len(tagStack)-1]
				lineIndexes = lineIndexes[:len(lineIndexes)-1]
			} else {
				// 不匹配，标签嵌套错误
				errors = append(errors, &json_rule.CellError{
					Index:  rowIdx,
					Reason: fmt.Sprintf("标签嵌套错误: 开始标签<%s>与结束标签%s不匹配", lastTag, fullTag),
				})
			}
		} else {
			// 开始标签
			// 提取标签名（去掉属性和空格）
			tagName := tagContent
			if spaceIdx := strings.IndexAny(tagName, " ="); spaceIdx > 0 {
				tagName = tagName[:spaceIdx]
			}

			// 检查是否是自闭合标签
			if strings.HasSuffix(strings.TrimSpace(tagContent), "/") {
				// 自闭合标签，不入栈
				continue
			}

			tagStack = append(tagStack, tagName)
			lineIndexes = append(lineIndexes, pos)
		}
	}

	// 检查是否有未闭合的标签
	if len(tagStack) > 0 {
		for i, tag := range tagStack {
			errors = append(errors, &json_rule.CellError{
				Index:  rowIdx,
				Reason: fmt.Sprintf("标签<%s>未闭合 (位置:%d)", tag, lineIndexes[i]),
			})
		}
	}

	return errors
}

// 检查标签格式
func (c *RichTextCheckRule) checkTagFormat(text string, rowIdx int) []*json_rule.CellError {
	var errors []*json_rule.CellError

	// 匹配所有标签（排除占位符）
	tagRegex := regexp.MustCompile(`<([^>]+)>`)
	matches := tagRegex.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		fullTag := match[0]
		tagContent := match[1]

		// 处理结束标签
		if strings.HasPrefix(tagContent, "/") {
			tagName := strings.TrimPrefix(tagContent, "/")
			if spaceIdx := strings.IndexAny(tagName, " ="); spaceIdx > 0 {
				tagName = tagName[:spaceIdx]
			}

			// 检查支持的标签名
			supportedTags := map[string]bool{
				"color":  true,
				"sprite": true,
			}

			if !supportedTags[tagName] {
				errors = append(errors, &json_rule.CellError{
					Index:  rowIdx,
					Reason: fmt.Sprintf("不支持的结束标签: %s", fullTag),
				})
			}
			continue
		}

		// 处理开始标签和自闭合标签
		isSelfClosing := strings.HasSuffix(strings.TrimSpace(tagContent), "/")

		// 提取标签名
		tagName := tagContent
		if spaceIdx := strings.IndexAny(tagName, " =/"); spaceIdx > 0 {
			tagName = tagName[:spaceIdx]
		}

		// 检查支持的标签名
		supportedTags := map[string]bool{
			"color":  true,
			"sprite": true,
		}

		if !supportedTags[tagName] {
			errors = append(errors, &json_rule.CellError{
				Index:  rowIdx,
				Reason: fmt.Sprintf("不支持的标签名: %s", fullTag),
			})
			continue
		}

		// 解析属性
		attrRegex := regexp.MustCompile(`(\w+)=([^\s>]+)`)
		attrs := attrRegex.FindAllStringSubmatch(tagContent, -1)

		// 根据标签名检查属性
		switch tagName {
		case "color":
			hasColorAttr := false
			for _, attr := range attrs {
				attrName := attr[1]
				attrValue := attr[2]

				if attrName != "color" {
					errors = append(errors, &json_rule.CellError{
						Index:  rowIdx,
						Reason: fmt.Sprintf("color标签只能有color属性，当前有: %s=%s", attrName, attrValue),
					})
				} else {
					hasColorAttr = true
					// 检查颜色格式：#RRGGBB 或 #RGB
					colorRegex := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$|^#[0-9A-Fa-f]{3}$`)
					if !colorRegex.MatchString(attrValue) {
						errors = append(errors, &json_rule.CellError{
							Index:  rowIdx,
							Reason: fmt.Sprintf("颜色格式错误: %s (应为 #RRGGBB 或 #RGB)", attrValue),
						})
					}
				}
			}
			if !hasColorAttr && !isSelfClosing {
				errors = append(errors, &json_rule.CellError{
					Index:  rowIdx,
					Reason: fmt.Sprintf("color标签缺少color属性: %s", fullTag),
				})
			}

		case "sprite":
			hasSpriteAttr := false
			for _, attr := range attrs {
				attrName := attr[1]
				attrValue := attr[2]

				if attrName != "sprite" {
					errors = append(errors, &json_rule.CellError{
						Index:  rowIdx,
						Reason: fmt.Sprintf("sprite标签只能有sprite属性，当前有: %s=%s", attrName, attrValue),
					})
				} else {
					hasSpriteAttr = true
					// 检查sprite值是否为数字
					if _, err := strconv.Atoi(attrValue); err != nil {
						errors = append(errors, &json_rule.CellError{
							Index:  rowIdx,
							Reason: fmt.Sprintf("sprite值必须是数字: %s", attrValue),
						})
					}
				}
			}
			if !hasSpriteAttr && !isSelfClosing {
				errors = append(errors, &json_rule.CellError{
					Index:  rowIdx,
					Reason: fmt.Sprintf("sprite标签缺少sprite属性: %s", fullTag),
				})
			}
		}
	}

	return errors
}

// 检查占位符
func (c *RichTextCheckRule) checkPlaceholders(text string, rowIdx int) []*json_rule.CellError {
	var errors []*json_rule.CellError

	// 匹配所有占位符 {0}, {1}, {2} 等
	placeholderRegex := regexp.MustCompile(`\{(\d+)\}`)
	matches := placeholderRegex.FindAllStringSubmatch(text, -1)

	if len(matches) == 0 {
		return errors
	}

	// 检查占位符索引是否从0开始递增
	maxIndex := -1
	indexMap := make(map[int]bool)

	for _, match := range matches {
		if len(match) > 1 {
			index, _ := strconv.Atoi(match[1])
			indexMap[index] = true
			if index > maxIndex {
				maxIndex = index
			}
		}
	}

	// 检查是否从0开始连续
	for i := 0; i <= maxIndex; i++ {
		if !indexMap[i] {
			errors = append(errors, &json_rule.CellError{
				Index:  rowIdx,
				Reason: fmt.Sprintf("占位符缺少 {%d}，必须从0开始连续递增", i),
			})
		}
	}

	// 检查占位符格式是否正确（没有多余空格等）
	for _, match := range matches {
		placeholder := match[0]
		if strings.Contains(placeholder, " ") {
			errors = append(errors, &json_rule.CellError{
				Index:  rowIdx,
				Reason: fmt.Sprintf("占位符格式错误: %s (不能包含空格)", placeholder),
			})
		}
	}

	return errors
}

// 检查允许的字符
func (c *RichTextCheckRule) checkAllowedChars(text string, rowIdx int) []*json_rule.CellError {
	var errors []*json_rule.CellError

	// 移除所有标签
	stripTagsRegex := regexp.MustCompile(`<[^>]*>`)
	plainText := stripTagsRegex.ReplaceAllString(text, "")

	// 移除占位符
	placeholderRegex := regexp.MustCompile(`\{\d+\}`)
	plainText = placeholderRegex.ReplaceAllString(plainText, "")

	// 允许的字符：中文、中文标点、数字、英文、英文标点、空格
	for i, r := range plainText {
		// 允许的字符范围
		isChinese := unicode.Is(unicode.Han, r)
		isChinesePunct := isChinesePunctuation(r)
		isDigit := unicode.IsDigit(r)
		isEnglish := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		isEnglishPunct := isEnglishPunctuation(r)
		isSpace := unicode.IsSpace(r)

		if !isChinese && !isChinesePunct && !isDigit && !isEnglish && !isEnglishPunct && !isSpace {
			errors = append(errors, &json_rule.CellError{
				Index:  rowIdx,
				Reason: fmt.Sprintf("包含不允许的字符: %q (位置:%d)", r, i),
			})
		}
	}

	return errors
}

// 检查是否是中文标点
func isChinesePunctuation(r rune) bool {
	switch {
	case r >= 0x3000 && r <= 0x303F: // CJK符号和标点
		return true
	case r >= 0xFF00 && r <= 0xFFEF: // 全角ASCII、全角标点
		return true
	case r == '，' || r == '。' || r == '！' || r == '？' || r == '；' || r == '：' ||
		r == '「' || r == '」' || r == '『' || r == '』' || r == '【' || r == '】' ||
		r == '《' || r == '》' || r == '（' || r == '）' || r == '、' || r == '…' ||
		r == '～' || r == '·' || r == '―':
		return true
	default:
		return false
	}
}

// 检查是否是英文标点
func isEnglishPunctuation(r rune) bool {
	switch r {
	case '\'', '(', ')', '[', ']', '{', '}',
		'-', '_', '+', '=', '*', '&', '^', '%', '$', '#', '@', '~', '`', '\\', '|',
		'/', '<', '>':
		return true
	case '.', ',', '!', '?', ';', ':', '"':
		return false
	default:
		return false
	}
}
