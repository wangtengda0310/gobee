package engine

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestCollectRelevantSheetsForColumn_CrossReference(t *testing.T) {
	colRule := &json_rule.SheetColRule{
		PropName: "RewardId",
		PropRules: []*json_rule.ColRule{
			{
				Type: json_rule.FOREIGN_KEY,
				Params: map[string]string{
					"targetSheet": "道具|Item",
				},
			},
		},
	}

	sheets := CollectRelevantSheetsForColumn("活动任务|ActivityTask", colRule)

	assert.True(t, sheets["活动任务|ActivityTask"])
	assert.True(t, sheets["道具|Item"])
	assert.Len(t, sheets, 2)
}

func TestFilterXlsxFileNamesByRequiredSheets(t *testing.T) {
	allNames := []string{
		"Hero.xlsx",
		"Item.xlsx",
		"Activity.xlsx",
		"Survey_问卷.xlsx",
	}

	required := map[string]bool{
		"Hero": true,
		"Item": true,
	}

	filtered := filterXlsxFileNamesByRequiredSheets(allNames, required)

	assert.Contains(t, filtered, "Hero.xlsx")
	assert.Contains(t, filtered, "Item.xlsx")
	assert.NotContains(t, filtered, "Survey_问卷.xlsx")
}

func TestFilterXlsxFileNames_PipeSheetName(t *testing.T) {
	allNames := []string{
		"SeasonPass_赛季战令表.xlsx",
		"SeasonPassReward_赛季战令奖励表.xlsx",
		"Item.xlsx",
	}

	required := map[string]bool{
		"赛季战令表|SeasonPass":         true,
		"赛季战令奖励表|SeasonPassReward": true,
	}

	filtered := filterXlsxFileNamesByRequiredSheets(allNames, required)

	assert.Contains(t, filtered, "SeasonPass_赛季战令表.xlsx")
	assert.Contains(t, filtered, "SeasonPassReward_赛季战令奖励表.xlsx")
	assert.NotContains(t, filtered, "Item.xlsx")
}

func TestCollectRelevantSheetsForColumn_CommaSeparatedChainRequired(t *testing.T) {
	colRule := &json_rule.SheetColRule{
		PropRules: []*json_rule.ColRule{
			{
				Type: json_rule.CHAIN_REFERENCE,
				Params: map[string]string{
					"chainRequiredSheets": "赛季战令表|SeasonPass,赛季战令奖励表|SeasonPassReward",
				},
			},
		},
	}

	sheets := CollectRelevantSheetsForColumn("道具|Item", colRule)

	assert.True(t, sheets["道具|Item"])
	assert.True(t, sheets["赛季战令表|SeasonPass"])
	assert.True(t, sheets["赛季战令奖励表|SeasonPassReward"])
}

func TestFindMissingRequiredSheets(t *testing.T) {
	sheetMap := map[string]*excelize.File{
		"武将|Hero": nil,
		"道具|Item": nil,
	}

	required := map[string]bool{
		"Hero":     true,
		"Item":     true,
		"DropItem": true,
	}

	missing := findMissingRequiredSheets(sheetMap, required)
	assert.Contains(t, missing, "DropItem")
	assert.NotContains(t, missing, "Hero")
	assert.NotContains(t, missing, "Item")
}

func TestCollectRelevantSheetsForColumn_ChainSteps(t *testing.T) {
	chainSteps := `{"left":{"steps":[{"sheet":"Item","preCol":"Id","findVal":"col","nextCol":"ItemParam"}]},"right":{"steps":[{"sheet":"Hero","preCol":"Id","findVal":"col","nextCol":"Name"}]}}`
	colRule := &json_rule.SheetColRule{
		PropRules: []*json_rule.ColRule{
			{
				Type: json_rule.CHAIN_REFERENCE,
				Params: map[string]string{
					"chainSteps": chainSteps,
				},
			},
		},
	}

	sheets := CollectRelevantSheetsForColumn("活动|Activity", colRule)

	assert.True(t, sheets["活动|Activity"])
	assert.True(t, sheets["Item"])
	assert.True(t, sheets["Hero"])
}

func TestCollectRelevantSheetsForColumn_NilColRule(t *testing.T) {
	sheets := CollectRelevantSheetsForColumn("武将|Hero", nil)
	assert.Len(t, sheets, 1)
	assert.True(t, sheets["武将|Hero"])
}

func TestCloseSheetMap_DedupesFiles(t *testing.T) {
	f := excelize.NewFile()

	sheetMap := map[string]*excelize.File{
		"SheetA": f,
		"SheetB": f,
	}

	assert.NotPanics(t, func() {
		closeSheetMap(sheetMap)
	})
}
