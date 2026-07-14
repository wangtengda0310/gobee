package feishu

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultFeishuMessageTemplate 默认消息模板
const DefaultFeishuMessageTemplate = `【来自claude code】战斗测试完成
英雄：{heroName}
总数：{total}
通过：{passed}
失败：{failed}
通过率：{passRate}%`

// FeishuNotifyConfig 飞书通知配置
type FeishuNotifyConfig struct {
	Enabled         bool   `json:"enabled"`
	RobotGUID       string `json:"robotGuid"`
	MessageTemplate string `json:"messageTemplate"`
}

// FeishuNotifyConfigService 飞书通知配置服务
// @frontend
type FeishuNotifyConfigService struct {
	configFile string
}

// NewFeishuNotifyConfigService 创建飞书通知配置服务实例
func NewFeishuNotifyConfigService() *FeishuNotifyConfigService {
	// 配置文件放在当前目录下的 feishu_notify_config.json
	configFile := "feishu_notify_config.json"

	// 如果是相对路径，转换为绝对路径
	if !filepath.IsAbs(configFile) {
		abs, err := filepath.Abs(configFile)
		if err == nil {
			configFile = abs
		}
	}

	return &FeishuNotifyConfigService{
		configFile: configFile,
	}
}

// GetConfig 获取飞书通知配置
// @frontend
func (s *FeishuNotifyConfigService) GetConfig() (*FeishuNotifyConfig, error) {
	data, err := os.ReadFile(s.configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 返回默认配置
			return &FeishuNotifyConfig{
				Enabled:         false,
				RobotGUID:       "",
				MessageTemplate: DefaultFeishuMessageTemplate,
			}, nil
		}
		return nil, err
	}

	var config FeishuNotifyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig 保存飞书通知配置
// @frontend
func (s *FeishuNotifyConfigService) SaveConfig(config *FeishuNotifyConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.configFile, data, 0644)
}

// SendNotification 发送飞书通知
func (s *FeishuNotifyConfigService) SendNotification(heroName string, total, passed, failed int, passRate float64) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}

	if !config.Enabled {
		return nil
	}

	if config.RobotGUID == "" {
		return nil
	}

	message := buildFeishuMessage(config.MessageTemplate, heroName, total, passed, failed, passRate)
	SendFeiShuRobotText(config.RobotGUID, "%s", message)

	return nil
}

// buildFeishuMessage 构建飞书消息
func buildFeishuMessage(template, heroName string, total, passed, failed int, passRate float64) string {
	message := template
	message = strings.ReplaceAll(message, "{heroName}", heroName)
	message = strings.ReplaceAll(message, "{total}", fmt.Sprintf("%d", total))
	message = strings.ReplaceAll(message, "{passed}", fmt.Sprintf("%d", passed))
	message = strings.ReplaceAll(message, "{failed}", fmt.Sprintf("%d", failed))
	message = strings.ReplaceAll(message, "{passRate}", fmt.Sprintf("%.1f", passRate))
	return message
}

// FeishuNotifier 飞书通知器接口
// 用于避免循环依赖：function-test 和 settings 都可以使用此接口
type FeishuNotifier interface {
	GetConfig() (*FeishuNotifyConfig, error)
	SaveConfig(config *FeishuNotifyConfig) error
	SendNotification(heroName string, total, passed, failed int, passRate float64) error
}
