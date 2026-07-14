package variables

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/params"
	"github.com/stretchr/testify/assert"
)

// ========== 变量注册测试（验证 registry.go 的 init()） ==========

// TestGetAvailableVariables 验证返回非空列表，含 cityId 与 openid
func TestGetAvailableVariables(t *testing.T) {
	vars := params.GetAvailableVariables()
	assert.NotEmpty(t, vars, "变量列表不应为空")

	foundCity := false
	foundOpenID := false
	for _, v := range vars {
		if v.ShortName == "cityId" {
			foundCity = true
			assert.Equal(t, "🏰 工会城战 - 城池ID", v.DisplayName)
		}
		if v.ShortName == "openid" {
			foundOpenID = true
			assert.Equal(t, "👤 当前账号", v.DisplayName)
		}
	}
	assert.True(t, foundCity, "应包含 cityId 变量")
	assert.True(t, foundOpenID, "应包含 openid 变量")
}

// TestFindVariableByShortName 找到 cityId/openid，找不到返回 nil
func TestFindVariableByShortName(t *testing.T) {
	def := params.FindVariableByShortName("cityId")
	assert.NotNil(t, def, "应找到 cityId")
	assert.Equal(t, "cityId", def.ShortName)
	assert.NotNil(t, def.ExtractFunc)

	openIDDef := params.FindVariableByShortName("openid")
	assert.NotNil(t, openIDDef, "应找到 openid")
	assert.Equal(t, "openid", openIDDef.ShortName)
	assert.Empty(t, openIDDef.WatchMsgIDs, "openid 为账号级变量，不依赖 Ntf 提取")

	notFound := params.FindVariableByShortName("nonexistent")
	assert.Nil(t, notFound, "不存在的变量应返回 nil")
}

// TestVariableAvailableReqs 验证各变量的 AvailableReqs 注册值
// AvailableReqs 限制变量在前端卡片模式对哪些 Req 可见（msg_name）；
// nil/空表示对所有 Req 可用（如 openid）
func TestVariableAvailableReqs(t *testing.T) {
	cases := []struct {
		shortName     string
		availableReqs []string
	}{
		{"cityId", []string{"TeamSelectGuildCityReq"}},
		{"roomCreator", []string{"RoomLookOnReq"}},
		{"roomID", []string{"RoomLookOnReq"}},
		{"openid", nil}, // 账号级变量，对所有 Req 可用
	}
	for _, c := range cases {
		def := params.FindVariableByShortName(c.shortName)
		if !assert.NotNil(t, def, "应找到 %s", c.shortName) {
			continue
		}
		assert.Equal(t, c.availableReqs, def.AvailableReqs, "%s 的 AvailableReqs 不匹配", c.shortName)
	}
}

// TestGetAvailableVariables_ContainsAvailableReqs 验证 VariableInfo 携带 AvailableReqs 下发前端
func TestGetAvailableVariables_ContainsAvailableReqs(t *testing.T) {
	vars := params.GetAvailableVariables()
	cityID := findInfo(vars, "cityId")
	if assert.NotNil(t, cityID) {
		assert.Equal(t, []string{"TeamSelectGuildCityReq"}, cityID.AvailableReqs)
	}
	roomCreator := findInfo(vars, "roomCreator")
	if assert.NotNil(t, roomCreator) {
		assert.Equal(t, []string{"RoomLookOnReq"}, roomCreator.AvailableReqs)
	}
	openID := findInfo(vars, "openid")
	if assert.NotNil(t, openID) {
		assert.Empty(t, openID.AvailableReqs, "openid 应对所有 Req 可用")
	}
}

func findInfo(vars []params.VariableInfo, shortName string) *params.VariableInfo {
	for i := range vars {
		if vars[i].ShortName == shortName {
			return &vars[i]
		}
	}
	return nil
}
