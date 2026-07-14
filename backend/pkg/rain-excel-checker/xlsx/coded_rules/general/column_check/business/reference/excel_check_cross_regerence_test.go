// Package reference 提供引用关系相关的校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package reference

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestCrossReferenceCheckRule(t *testing.T) {
	// 检查 Excel 文件路径是否存在
	excelPath := "../../config/excel"
	if _, err := os.Stat(excelPath); os.IsNotExist(err) {
		t.Skipf("Excel 文件路径不存在，跳过测试: %s", excelPath)
	}

	cols, cIdx, c1idx, params, sheetMap := fakeDataCrossReferenceCheckRule(t)

	bc := new(CrossReferenceCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataCrossReferenceCheckRule(t *testing.T) (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)

	params["refSheet"] = "活动表|Activity"
	params["refCol"] = "Id"

	// 读取文件
	excels, err := excelio.ReadFileOrDir("../../" + "config/excel")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, excel := range excels {
			excel.Close()
		}
	}()
	// 因为文件名并不完全是表名，所以接下来访问每个文件里的sheet名
	if len(excels) == 0 {
		t.Fatal("no excel")
	}
	fmt.Println("加载了", len(excels), "个excel文件")
	filter, err := excelio.ExcelFilter(excels)
	if err != nil {
		t.Fatal(err)
	}
	// 构建sheet对文件映射
	sheetMap = make(map[string]*excelize.File)
	for file, sheets := range filter {
		for _, sheet := range sheets {
			sheetMap[sheet.Name] = file
		}
	}

	return [][]string{
		{"1", "2", "999"},
	}, 0, 0, params, sheetMap
}

func TestCrossReferenceCheckRule2(t *testing.T) {
	// 检查 Excel 文件路径是否存在
	excelPath := "../../" + "config/excel"
	if _, err := os.Stat(excelPath); os.IsNotExist(err) {
		t.Skipf("Excel 文件路径不存在，跳过测试: %s", excelPath)
	}

	cols, cIdx, c1idx, params, sheetMap := fakeDataCrossReferenceCheckRule2(t)

	bc := new(CrossReferenceCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataCrossReferenceCheckRule2(t *testing.T) (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)

	params["refSheet"] = "活动表|Activity"
	params["refCol"] = "Id"

	params["pattern"] = "{(\\d+);\\d+}"
	params["captureGroups"] = "1"

	// 读取文件
	sheetMap, err := excelio.GetSheetMap("../../" + "config/excel")
	if err != nil {
		t.Fatal(err)
	}

	return [][]string{
		{"{1;2}", "{2;3}{56;12}", "{999;4}"},
	}, 0, 0, params, sheetMap
}

func TestRegex(t *testing.T) {

	var extractPattern *regexp.Regexp

	extractPattern, err := regexp.Compile(`\{(\d+);\d+}`)
	if err != nil {
		t.Fatal(err)
	}

	var extractedValues []string

	matches := extractPattern.FindAllStringSubmatch("{213;345}{245;123}", -1)
	for _, match := range matches {
		for _, groupIdx := range []int{1} {
			if groupIdx < len(match) && match[groupIdx] != "" {
				extractedValues = append(extractedValues, match[groupIdx])
			}
		}
	}

	t.Log(extractedValues)
}

func TestCrossReferenceCheckRule_ArrayMode(t *testing.T) {
	// 创建模拟的参照表数据
	refFile := excelize.NewFile()
	_, err := refFile.NewSheet("道具表|Item")
	if err != nil {
		t.Fatal(err)
	}

	// 设置列名行（第3行，索引2）
	refFile.SetCellValue("道具表|Item", "A3", "Id")
	refFile.SetCellValue("道具表|Item", "B3", "Name")

	// 设置数据行（从第5行开始，索引4）
	refFile.SetCellValue("道具表|Item", "A5", "2000001")
	refFile.SetCellValue("道具表|Item", "A6", "4000001")
	refFile.SetCellValue("道具表|Item", "A7", "7000001")
	refFile.SetCellValue("道具表|Item", "A8", "1000040")

	sheetMap := map[string]*excelize.File{
		"道具表|Item": refFile,
	}

	// 测试数据：逗号分隔的数组格式
	cols := [][]string{
		{"{2000001;1},{4000001;5}", "{7000001;3},{9999999;1}", "{2000001;1}"},
	}

	bc := new(CrossReferenceCheckRule)

	// 数组模式 + 正则提取
	params := map[string]string{
		"refSheet":   "道具表|Item",
		"refCol":     "Id",
		"pattern":    `\{(\d+);\d+}`,
		"groups":     "1",
		"isArray":    "true",
		"matchType":  "exists",
		"exactMatch": "true",
	}

	res := bc.Check("", cols, 0, 0, params, sheetMap)

	// 第0行: {2000001;1},{4000001;5} - 2000001和4000001都在参照表中，应该通过
	// 第1行: {7000001;3},{9999999;1} - 9999999不在参照表中，应该失败
	// 第2行: {2000001;1} - 2000001在参照表中，应该通过（单元素数组）
	assert.Len(t, res, 1, "应该只有1个错误")
	assert.Equal(t, 1, res[0].Index, "第1行应该报错（9999999不存在）")
}

func TestCrossReferenceCheckRule_SingleModeWithPattern(t *testing.T) {
	// 创建模拟的参照表数据
	refFile := excelize.NewFile()
	_, err := refFile.NewSheet("道具表|Item")
	if err != nil {
		t.Fatal(err)
	}

	refFile.SetCellValue("道具表|Item", "A3", "Id")
	refFile.SetCellValue("道具表|Item", "A5", "2000001")
	refFile.SetCellValue("道具表|Item", "A6", "4000001")

	sheetMap := map[string]*excelize.File{
		"道具表|Item": refFile,
	}

	// 测试数据：多组花括号（非数组，是连续的多组）
	cols := [][]string{
		{"{2000001;1}{4000001;5}", "{9999999;1}"},
	}

	bc := new(CrossReferenceCheckRule)

	// 单值模式 + 正则提取（不使用isArray）
	params := map[string]string{
		"refSheet":   "道具表|Item",
		"refCol":     "Id",
		"pattern":    `\{(\d+);\d+}`,
		"groups":     "1",
		"matchType":  "exists",
		"exactMatch": "true",
	}

	res := bc.Check("", cols, 0, 0, params, sheetMap)

	// 第0行: {2000001;1}{4000001;5} - 通过正则提取2000001和4000001，都在参照表中
	// 第1行: {9999999;1} - 9999999不在参照表中，应该失败
	assert.Len(t, res, 1, "应该只有1个错误")
	assert.Equal(t, 1, res[0].Index, "第1行应该报错")
}

func TestSplitArrayElements(t *testing.T) {
	// 测试数组拆分函数
	result := splitArrayElements("{1;2},{3;4},{5;6}", ",")
	assert.Equal(t, []string{"{1;2}", "{3;4}", "{5;6}"}, result)

	// 测试无逗号的情况
	result = splitArrayElements("{1;2}", ",")
	assert.Equal(t, []string{"{1;2}"}, result)

	// 测试空字符串
	result = splitArrayElements("", ",")
	assert.Equal(t, 0, len(result))

	// 测试花括号内有逗号的情况（不应该拆分）
	result = splitArrayElements("{1;2,3},{4;5}", ",")
	assert.Equal(t, []string{"{1;2,3}", "{4;5}"}, result)
}
