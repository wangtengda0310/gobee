package internal

import (
	"path/filepath"
	"testing"
)

func TestParseExcel(t *testing.T) {
	excelPath := filepath.Join("..", "testdata", "demo.xlsx")
	structs, err := ParseExcel(excelPath)
	if err != nil {
		t.Fatalf("ParseExcel 解析失败: %v", err)
	}
	if len(structs) == 0 {
		t.Fatalf("ParseExcel 结果为空")
	}
	// 简单断言第一个 struct 的部分字段
	st := structs[0]
	if st.Name == "" || len(st.Entries) == 0 {
		t.Errorf("第一个 struct 解析异常: %+v", st)
	}
	// 可根据实际 demo.xlsx 增加更多断言
}
