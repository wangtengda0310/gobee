package json_rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGeneralRuleOverrides_SheetMatch(t *testing.T) {
	// 精确匹配
	overrides := GetGeneralRuleOverrides("PetAudio")
	assert.Equal(t, "AnimationState,ItemCfgId", overrides[NEW_ROW_NOTIFY][string(ID_COL_NAMES)],
		"应该返回 PetAudio 的复合主键配置")
	assert.Equal(t, "AnimationState,ItemCfgId", overrides[ROW_CHANGE_NOTIFY][string(ID_COL_NAMES)],
		"ROW_CHANGE_NOTIFY 也应该有相同配置")

	// 后缀匹配
	overrides2 := GetGeneralRuleOverrides("灵宠音效|PetAudio")
	assert.Equal(t, "AnimationState,ItemCfgId", overrides2[NEW_ROW_NOTIFY][string(ID_COL_NAMES)],
		"后缀匹配应该生效")
}

func TestGetGeneralRuleOverrides_NoMatch(t *testing.T) {
	// 不匹配的表应该返回空 map
	overrides := GetGeneralRuleOverrides("SomeOtherTable")
	assert.Equal(t, 0, len(overrides), "不匹配的表应该返回空 map")
}

func TestGetGeneralRuleOverrides_PetTriggerWeight(t *testing.T) {
	// PetTriggerWeight 也需要复合主键
	overrides := GetGeneralRuleOverrides("PetTriggerWeight")
	assert.Equal(t, "Id,ItemCfgId", overrides[NEW_ROW_NOTIFY][string(ID_COL_NAMES)])
}
