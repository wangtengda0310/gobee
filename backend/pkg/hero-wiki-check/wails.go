package herowikicheck

import (
	"encoding/json"
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/appconfig"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/hero_res_check"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel"
)

// HeroWikiResCheckService 武将 Wiki 资源检查服务
// @frontend
type HeroWikiResCheckService struct{}

// HeroWikiConfigService 武将 Wiki 配置管理服务
type HeroWikiConfigService struct {
	section *appconfig.Section
}

// HeroWikiConfig 应用配置结构
type HeroWikiConfig struct {
	ExcelDir   string `json:"excel_dir"`
	OldJsonDir string `json:"old_json_dir"`
}

// NewHeroWikiConfigService 创建配置服务实例
func NewHeroWikiConfigService() *HeroWikiConfigService {
	return &HeroWikiConfigService{
		section: appconfig.New("hero_wiki"),
	}
}

// getDefaultConfig 获取默认配置
func (cs *HeroWikiConfigService) getDefaultConfig() *HeroWikiConfig {
	return &HeroWikiConfig{
		ExcelDir:   "../../config/excel",
		OldJsonDir: "tmp",
	}
}

// GetConfig 获取当前配置
// @frontend @mcp
func (cs *HeroWikiConfigService) GetConfig() (*HeroWikiConfig, error) {
	var config HeroWikiConfig
	if err := cs.section.Load(&config); err != nil {
		return nil, fmt.Errorf("读取配置失败: %v", err)
	}

	if !cs.section.Exists() {
		return cs.getDefaultConfig(), nil
	}

	return &config, nil
}

// SaveConfig 保存配置
// @frontend @mcp
func (cs *HeroWikiConfigService) SaveConfig(config *HeroWikiConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}
	return cs.section.Save(config)
}

// UpdateConfig 更新配置（部分更新）
func (cs *HeroWikiConfigService) UpdateConfig(updates map[string]interface{}) (*HeroWikiConfig, error) {
	// 获取当前配置
	currentConfig, err := cs.GetConfig()
	if err != nil {
		return nil, err
	}

	// 临时转换为 map 以便更新
	configMap := make(map[string]interface{})
	data, _ := json.Marshal(currentConfig)
	json.Unmarshal(data, &configMap)

	// 应用更新
	for key, value := range updates {
		if _, exists := configMap[key]; exists {
			configMap[key] = value
		}
	}

	// 转换回 HeroWikiConfig 结构体
	updatedData, _ := json.Marshal(configMap)
	var updatedConfig HeroWikiConfig
	if err := json.Unmarshal(updatedData, &updatedConfig); err != nil {
		return nil, fmt.Errorf("更新配置失败: %v", err)
	}

	// 保存更新后的配置
	if err := cs.SaveConfig(&updatedConfig); err != nil {
		return nil, err
	}

	return &updatedConfig, nil
}

// ResetToDefault 重置为默认配置
func (cs *HeroWikiConfigService) ResetToDefault() (*HeroWikiConfig, error) {
	defaultConfig := cs.getDefaultConfig()
	if err := cs.SaveConfig(defaultConfig); err != nil {
		return nil, err
	}
	return defaultConfig, nil
}

// GetConfigFilePath 获取配置文件路径
func (cs *HeroWikiConfigService) GetConfigFilePath() string {
	return appconfig.FilePath()
}

// ConfigExists 检查配置是否存在
func (cs *HeroWikiConfigService) ConfigExists() bool {
	return cs.section.Exists()
}

// DeleteConfig 删除配置
func (cs *HeroWikiConfigService) DeleteConfig() error {
	return cs.section.Delete()
}

// NewHeroWikiResCheckService 创建武将 Wiki 资源检查服务
func NewHeroWikiResCheckService() *HeroWikiResCheckService {
	return &HeroWikiResCheckService{}
}

// Check 检查武将Wiki资源
// cardDir 表示同级目录的../../client/Master/Card/位置
// oldJsonPath "tmp/herowiki_test.json"
// @frontend @mcp
func (h *HeroWikiResCheckService) Check(excelDir, oldJsonPath string) (*diff.DataContainer, error) {
	sheetMap, err := excelio.GetSheetMap(excelDir)
	if err != nil {
		return nil, err
	}
	excels, err := mjs_excel.InitDiffRefExcel(sheetMap)
	if err != nil {
		return nil, err
	}

	excels.HeroWikiDiff = herowiki.BuildHeroWikiDiff(excels)

	comparator := hero_res_check.NewComparator(oldJsonPath)
	if oldData, err := comparator.LoadPreviousData(); err == nil {
		result := comparator.CompareHeroWikiDiff(oldData.HeroWikiDiff, excels.HeroWikiDiff)
		excels.HeroWikiDiffResult = result
	}

	return excels, nil
}

// Save 保存当前检查结果到指定路径
// savePath: 保存路径，由前端 oldJsonPath 传入
// container: 当前检查结果数据
// @frontend @mcp
func (h *HeroWikiResCheckService) Save(savePath string, container *diff.DataContainer) error {
	comparator := hero_res_check.NewComparator(savePath)
	return comparator.SaveCurrentData(container)
}

// HeroWikiGameService 武将 Wiki 检查页面专用的 Game 服务包装器
// 只暴露武将 Wiki 检查页面需要的方法，遵循接口隔离原则
// @frontend @mcp
type HeroWikiGameService struct {
	*game.GameExcelService
}

// NewHeroWikiGameService 创建武将 Wiki 检查页面的 Game 服务实例
func NewHeroWikiGameService() *HeroWikiGameService {
	return &HeroWikiGameService{
		GameExcelService: game.NewGameExcelService(),
	}
}
