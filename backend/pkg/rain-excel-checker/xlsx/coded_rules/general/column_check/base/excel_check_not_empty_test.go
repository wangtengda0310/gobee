// Package base 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package base

import (
	"encoding/json"
	"os"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/xuri/excelize/v2"
)

func TestNotEmptyCheckRule(t *testing.T) {
	// 检查 Excel 文件路径是否存在
	excelPath := "../../config/excel"
	if _, err := os.Stat(excelPath); os.IsNotExist(err) {
		t.Skipf("Excel 文件路径不存在，跳过测试: %s", excelPath)
	}

	cols, cIdx, c1idx, params, sheetMap := fakeDataNotEmptyCheckRule(t)

	bc := new(NotEmptyCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataNotEmptyCheckRule(t *testing.T) (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.ALLOW_COMMIT)] = "false"

	// 读取文件
	sheetMap, err := excelio.GetSheetMap("../../" + "config/excel")
	if err != nil {
		t.Fatal(err)
	}
	cols, err = sheetMap["武将台词|HeroLines"].GetCols("武将台词|HeroLines")
	if err != nil {
		t.Fatal(err)
	}
	for i, col := range cols {
		if col[excelio.MJS_FIXED_ROWS_NAME] == "AudioId" {
			cIdx = i
			break
		}
	}

	return cols, cIdx, excelio.MJS_FIXED_ROWS_NUM, params, sheetMap
}
