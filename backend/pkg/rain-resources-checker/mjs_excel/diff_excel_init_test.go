package mjs_excel

import (
	"encoding/json"
	"os"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki"
)

func TestDiffExcel(t *testing.T) {
	sheetMap, err := excelio.GetSheetMap("../config/excel")
	if err != nil {
		t.Fatal(err)
	}
	excels, err := InitDiffRefExcel(sheetMap)
	if err != nil {
		t.Fatal(err)
	}
	//excels.HeroWikiDiff = herowiki.BuildHeroWikiDiff(excels)
	var wikiData diff.DataContainer
	wikiData.HeroWikiDiff = herowiki.BuildHeroWikiDiff(excels)
	if jd, err := json.MarshalIndent(wikiData, "", "  "); err != nil {
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
