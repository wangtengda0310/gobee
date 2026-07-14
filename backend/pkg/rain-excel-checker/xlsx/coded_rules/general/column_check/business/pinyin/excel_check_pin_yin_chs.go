// Package pinyin 提供拼音相关的列级校验规则
// 本包中的规则用于检查拼音与中文的对应关系

package pinyin

import (
	"fmt"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/mozillazg/go-pinyin"
	"github.com/xuri/excelize/v2"
)

// PinYinCHSCheckRule 布尔值检查
type PinYinCHSCheckRule struct{}

func (c *PinYinCHSCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 获取中文列的索引
	chsColOffset, err := strconv.Atoi(params["chsColOffset"])
	if err != nil || chsColOffset == 0 {
		// 参数错误
		return []*json_rule.CellError{
			{
				Index:  0,
				Reason: fmt.Sprintf("中文列索引参数错误: %s", params["chsColOffset"]),
			},
		}
	}
	if len(cols) < colIdx+chsColOffset || colIdx+chsColOffset < 0 {
		// 参数错误
		return []*json_rule.CellError{
			{
				Index:  0,
				Reason: fmt.Sprintf("中文列索引参数错误: 越界[%d], 实际:(%d)", colIdx+chsColOffset, len(cols)),
			},
		}
	}
	endIdx2 := helpers.GetColEndIndex(cols, colIdx+chsColOffset, startRowIdx, breakLine, params)
	chsData := cols[colIdx+chsColOffset][startRowIdx:endIdx2]

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

	for i, s := range myColData {
		// 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

		if len(chsData) <= i {
			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   fmt.Sprintf("中文列长度不足: %d, 索引: %d", len(chsData), i),
			})
			continue
		}

		if !strings.Contains(s, c.convertToPinyin(chsData[i])) {
			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   fmt.Sprintf("拼音列不包含中文列转换后的音频字符串, %s ∉ %s", s, chsData[i]),
			})
		}
	}
	return res
}

func (c *PinYinCHSCheckRule) convertToPinyin(chinese string) string {
	// 配置拼音转换选项
	args := pinyin.NewArgs()
	//args.Style = pinyin.Tone // 带声调
	args.Style = pinyin.Normal // 不带声调

	// 获取拼音切片
	pinyinSlice := pinyin.Pinyin(chinese, args)

	// 拼接成字符串，每个字的拼音首字母大写
	var result []string
	for _, wordPinyin := range pinyinSlice {
		if len(wordPinyin) > 0 {
			// 将每个字的首字母大写
			p := wordPinyin[0]
			if len(p) > 0 {
				result = append(result, strings.Title(p))
			}
		}
	}

	return strings.Join(result, "")
}
