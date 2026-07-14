package settings

import (
	"encoding/json"
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/appconfig"
)

// ExcelConfig 统一策划配表目录配置
// 作为各模块读取策划配表的唯一入口，逐步替代 hero_wiki/activity_wiki/excel_test 等
// 模块各自维护的 excel_dir 配置。
// @frontend @mcp
type ExcelConfig struct {
	ExcelDir string `json:"excel_dir"`
}

// ExcelConfigService 统一策划配表目录配置服务
type ExcelConfigService struct {
	section *appconfig.Section
}

// NewExcelConfigService 创建统一策划配表目录配置服务实例
func NewExcelConfigService() *ExcelConfigService {
	return &ExcelConfigService{
		section: appconfig.New("excel_config"),
	}
}

// getDefaultConfig 获取默认配置
func (s *ExcelConfigService) getDefaultConfig() *ExcelConfig {
	return &ExcelConfig{
		ExcelDir: "../../config",
	}
}

// GetConfig 获取当前统一策划配表目录配置
// @frontend @mcp
func (s *ExcelConfigService) GetConfig() (*ExcelConfig, error) {
	var config ExcelConfig
	if err := s.section.Load(&config); err != nil {
		return nil, fmt.Errorf("读取统一策划配表目录配置失败: %v", err)
	}

	if !s.section.Exists() {
		return s.getDefaultConfig(), nil
	}

	if config.ExcelDir == "" {
		config.ExcelDir = s.getDefaultConfig().ExcelDir
	}

	return &config, nil
}

// SaveConfig 保存统一策划配表目录配置
// @frontend @mcp
func (s *ExcelConfigService) SaveConfig(config *ExcelConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}
	if config.ExcelDir == "" {
		return fmt.Errorf("excel_dir 不能为空")
	}
	return s.section.Save(config)
}

// UpdateConfig 部分更新统一策划配表目录配置
// @frontend @mcp
func (s *ExcelConfigService) UpdateConfig(updates map[string]any) (*ExcelConfig, error) {
	currentConfig, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	configMap := make(map[string]any)
	data, _ := json.Marshal(currentConfig)
	json.Unmarshal(data, &configMap)

	for key, value := range updates {
		if _, exists := configMap[key]; exists {
			configMap[key] = value
		}
	}

	updatedData, _ := json.Marshal(configMap)
	var updatedConfig ExcelConfig
	if err := json.Unmarshal(updatedData, &updatedConfig); err != nil {
		return nil, fmt.Errorf("更新统一策划配表目录配置失败: %v", err)
	}

	if err := s.SaveConfig(&updatedConfig); err != nil {
		return nil, err
	}

	return &updatedConfig, nil
}

// ResetToDefault 重置为默认配置
// @frontend @mcp
func (s *ExcelConfigService) ResetToDefault() (*ExcelConfig, error) {
	defaultConfig := s.getDefaultConfig()
	if err := s.SaveConfig(defaultConfig); err != nil {
		return nil, err
	}
	return defaultConfig, nil
}

// GetConfigFilePath 获取配置文件路径
// @frontend
func (s *ExcelConfigService) GetConfigFilePath() string {
	return appconfig.FilePath()
}
