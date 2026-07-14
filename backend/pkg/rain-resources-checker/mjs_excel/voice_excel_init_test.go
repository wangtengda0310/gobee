package mjs_excel

import (
	"encoding/json"
	"os"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
)

func TestExcel(t *testing.T) {
	sheetMap, err := excelio.GetSheetMap("../config/excel")
	if err != nil {
		t.Fatal(err)
	}
	excels, err := InitAudioRefExcel(sheetMap)
	if err != nil {
		t.Fatal(err)
	}
	if jd, err := json.MarshalIndent(excels, "", "  "); err != nil {
		t.Fatal(err)
	} else {
		err = os.MkdirAll("tmp", 0777)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile("tmp/test.json", jd, 0777)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestExcelSearch(t *testing.T) {
	sheetMap, err := excelio.GetSheetMap("../config/excel")
	if err != nil {
		t.Fatal(err)
	}
	for sheetName, file := range sheetMap {
		cols, err := file.GetCols(sheetName)
		if err != nil {
			t.Fatal(err)
		}
		for _, col := range cols {
			if len(col) <= 4 {
				continue
			}
			//if strings.Contains(col[excelio.MJS_FIXED_ROWS_TYPE], "EAudioId") {
			//	fmt.Println(sheetName)
			//}
			//if col[excelio.MJS_FIXED_ROWS_TYPE] == "ESkillId" {
			//	fmt.Println(sheetName, path.Base(file.Path))
			//}
			//if col[excelio.MJS_FIXED_ROWS_NAME] == "SkinItemId" {
			//	fmt.Println(sheetName, path.Base(file.Path))
			//}
			//if col[excelio.MJS_FIXED_ROWS_NAME] == "SkinPinYin" {
			//	fmt.Println(sheetName, path.Base(file.Path))
			//}
		}
	}
}
