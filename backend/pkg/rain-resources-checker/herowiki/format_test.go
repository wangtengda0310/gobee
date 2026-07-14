package herowiki

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel"
)

func TestFormat(t *testing.T) {

	sheetMap, err := excelio.GetSheetMap("../config/excel")
	if err != nil {
		t.Fatal(err)
	}
	excels, err := mjs_excel.InitDiffRefExcel(sheetMap)
	if err != nil {
		t.Fatal(err)
	}
	excels.HeroWikiDiff = BuildHeroWikiDiff(excels)

	println("ok")
}
