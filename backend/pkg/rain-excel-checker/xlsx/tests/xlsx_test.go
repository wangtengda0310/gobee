package xlsx

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/engine"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/ruleconfig"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

func TestReadXlsx(t *testing.T) {
	// 读取文件
	excels, err := excelio.ReadFileOrDir("resources")
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
	filter, err := excelio.ExcelFilter(excels)
	if err != nil {
		t.Fatal(err)
	}
	//for _, fail := range failures {
	//	t.Log("失败", fail)
	//}
	for f, s := range filter {
		t.Log("成功", f.Path, s)
	}
}

func TestCheckXlsx(t *testing.T) {
	// 文件path
	xlsxPaths := []string{"resources/ArenaScore.xlsx", "resources/Activity_活动表.xlsx"}

	checkList := []*json_rule.SheetRule{
		{
			Sheet: "竞技场积分表|ArenaScore",
			Rules: map[string]*json_rule.SheetColRule{
				"Dan": {
					PropName: "Dan",
					PropType: "int",
					PropRules: []*json_rule.ColRule{
						{
							Type: json_rule.INCREASE_ID,
							Params: map[string]string{
								"test1": "0",
								"test2": "1",
								"test3": "false",
							},
						},
					},
				},
				"DanName": {
					PropName: "DanName",
					PropType: "string",
					PropRules: []*json_rule.ColRule{
						{
							Type: json_rule.CHS_ONLY,
							Params: map[string]string{
								"test1": "1",
								"test2": "0",
								"test3": "false",
							},
						},
					},
				},
			},
		},
		{
			Sheet: "活动表|Activity",
			Rules: map[string]*json_rule.SheetColRule{
				"StartTime": {
					PropName: "StartTime",
					PropType: "int",
					PropRules: []*json_rule.ColRule{
						{
							Type: json_rule.INCREASE_ID,
							Params: map[string]string{
								"test1": "3",
								"test2": "0",
								"test3": "false",
							},
						},
					},
				},
			},
		},
	}

	res := make([]*json_rule.ColCheckResult, 0, len(checkList))
	for i, check := range checkList {
		checkRes, err := engine.ReadAndCheckXlsx(xlsxPaths[i], check, nil)
		if err != nil {
			t.Fatal(err)
		}
		res = append(res, checkRes...)
	}

	// json表示
	jsonData, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(jsonData))
}

func TestSaveCheck(t *testing.T) {
	checkList := []*json_rule.SheetRule{
		{
			Sheet: "竞技场积分表|ArenaScore",
			Rules: map[string]*json_rule.SheetColRule{
				"Dan": {
					PropName: "Dan",
					PropType: "int",
					PropRules: []*json_rule.ColRule{
						{
							Type: json_rule.INCREASE_ID,
							Uuid: uuid.New().String(),
							Params: map[string]string{
								"test1": "2",
								"test2": "0",
								"test3": "false",
							},
						},
					},
				},
				"DanName": {
					PropName: "DanName",
					PropType: "string",
					PropRules: []*json_rule.ColRule{
						{
							Type: json_rule.CHS_ONLY,
							Uuid: uuid.New().String(),
							Params: map[string]string{
								"test1": "06",
								"test2": "0",
								"test3": "false",
							},
						},
					},
				},
			},
		},
		{
			Sheet: "活动表|Activity",
			Rules: map[string]*json_rule.SheetColRule{
				"StartTime": {
					PropName: "StartTime",
					PropType: "int",
					PropRules: []*json_rule.ColRule{
						{
							Type: json_rule.INCREASE_ID,
							Uuid: uuid.New().String(),
							Params: map[string]string{
								"test1": "456",
								"test2": "0",
								"test3": "false",
							},
						},
					},
				},
			},
		},
	}

	if err := ruleconfig.SaveCheck("cases/excel_cases/check_rule.json", checkList); err != nil {
		t.Fatal(err)
	}
}

func TestLoadJsonRules(t *testing.T) {
	if list, err := ruleconfig.LoadJsonRules("cases/excel_cases/check_rule.json"); err != nil {
		t.Fatal(err)
	} else {
		t.Log(list)
	}
}

// 可选：一个简单的只读取不验证的函数
func TestSimpleLoadJsonRules(t *testing.T) {
	// 获取所有JSON文件
	files, err := filepath.Glob("../rain-qa-func/cases/excel_cases/*.json")
	if err != nil {
		t.Fatalf("无法读取test目录: %v", err)
	}

	if len(files) == 0 {
		t.Skip("没有找到JSON文件，跳过测试")
	}

	var loadedRules []*json_rule.SheetRule

	// 串行读取（更简单）
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("读取文件 %s 失败: %v", file, err)
		}

		var rule *json_rule.SheetRule
		if err := json.Unmarshal(data, &rule); err != nil {
			t.Fatalf("解析JSON文件 %s 失败: %v", file, err)
		}

		loadedRules = append(loadedRules, rule)
		t.Logf("成功加载Sheet: %s, 包含 %d 个列规则", rule.Sheet, len(rule.Rules))
	}

	t.Logf("总共加载了 %d 个规则文件", len(loadedRules))
}

// 清理测试文件的函数（可选，用于测试后清理）
func TestCleanup(t *testing.T) {
	if err := os.RemoveAll("cases/excel_cases/check_rule.json"); err != nil {
		t.Logf("清理测试目录失败: %v", err)
	} else {
		t.Log("成功清理测试目录")
	}
}

func TestCheckAll(t *testing.T) {
	// 读取文件
	excels, err := excelio.ReadFileOrDir("resources")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, excel := range excels {
			excel.Close()
		}
	}()
	// 因为文件名并不完全是表名, 所以接下来访问每个文件里的sheet名
	if len(excels) == 0 {
		t.Fatal("no excel")
	}
	filter, err := excelio.ExcelFilter(excels)
	if err != nil {
		t.Fatal(err)
	}
	// 构建sheet对文件映射
	sheetMap := make(map[string]*excelize.File)
	for file, sheets := range filter {
		for _, sheet := range sheets {
			sheetMap[sheet.Name] = file
		}
	}
	// 加载规则
	if list, err := ruleconfig.LoadJsonRules("cases/excel_cases/check_rule.json"); err != nil {
		t.Fatal(err)
	} else {
		t.Log("加载规则:", len(list))
		for _, rule := range list {
			if xlsx, exist := sheetMap[rule.Sheet]; !exist {
				t.Fatal("文件不存在")
			} else {
				t.Log("检查表:", rule.Sheet)
				res, _ := engine.CheckSheetCols(xlsx, rule, sheetMap)
				okRes := make([]*json_rule.ColCheckResult, 0, len(res))
				for _, re := range res {
					if !re.Ok {
						okRes = append(okRes, re)
					}
				}
				jsonData, _ := json.MarshalIndent(okRes, "", " ")
				t.Log(string(jsonData))
			}
		}
	}
}

func TestFindEnum(t *testing.T) {
	// 读取文件
	excels, err := excelio.ReadFileOrDir("resources")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, excel := range excels {
			excel.Close()
		}
	}()
	// 因为文件名并不完全是表名, 所以接下来访问每个文件里的sheet名
	if len(excels) == 0 {
		t.Fatal("no excel")
	}
	filter, err := excelio.ExcelFilter(excels)
	if err != nil {
		t.Fatal(err)
	}
	// 构建sheet对文件映射
	sheetMap := make(map[string]*excelize.File)
	for file, sheets := range filter {
		for _, sheet := range sheets {
			sheetMap[sheet.Name] = file
		}
	}
	for sheetName, file := range sheetMap {
		cols, err := file.GetCols(sheetName)
		if err != nil {
			t.Fatal(err)
		}
		for _, col := range cols {
			if len(col) > excelio.MJS_FIXED_ROWS_NUM && strings.HasPrefix(col[excelio.MJS_FIXED_ROWS_TYPE], "E#") {
				t.Log(col[excelio.MJS_FIXED_ROWS_TYPE])
			}
		}
	}
}
