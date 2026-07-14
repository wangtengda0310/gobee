package functiontest

import (
	coregame "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// JsonCaseServiceProvieder Wails 服务提供者 - JSON 用例服务
// @frontend
type JsonCaseServiceProvieder struct {
	service *JsonCaseService
}

// NewJsonCaseServiceProvieder 创建 JSON 用例服务提供者
func NewJsonCaseServiceProvieder(app *application.App) *JsonCaseServiceProvieder {
	return &JsonCaseServiceProvieder{
		service: NewJsonCaseService(app),
	}
}

// GetService 获取服务实例
func (p *JsonCaseServiceProvieder) GetService() *JsonCaseService {
	return p.service
}

// FuncCaseConfigServiceProvieder Wails 服务提供者 - 配置服务
// @frontend
type FuncCaseConfigServiceProvieder struct {
	service *FuncCaseConfigService
}

// NewFuncCaseConfigServiceProvieder 创建配置服务提供者
func NewFuncCaseConfigServiceProvieder() *FuncCaseConfigServiceProvieder {
	return &FuncCaseConfigServiceProvieder{
		service: NewFuncCaseConfigService(),
	}
}

// GetService 获取服务实例
func (p *FuncCaseConfigServiceProvieder) GetService() *FuncCaseConfigService {
	return p.service
}

// FunctionTestGameService Wails 服务提供者 - 功能测试页面专用的 Game 服务
// 直接嵌入 core.GameExcelService，暴露所有方法
// @frontend @mcp
type FunctionTestGameService struct {
	*coregame.GameExcelService
}

// NewFunctionTestGameService 创建功能测试页面的 Game 服务实例
func NewFunctionTestGameService() *FunctionTestGameService {
	return &FunctionTestGameService{
		GameExcelService: coregame.NewGameExcelService(),
	}
}
