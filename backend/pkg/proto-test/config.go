package prototest

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/appconfig"
)

// protoTestSection 统一配置文件中的 proto_test section 读写器
var protoTestSection = appconfig.New("proto_test")

// ProtoTestConfig proto-test 监听配置
// @frontend @mcp
type ProtoTestConfig struct {
	// TCPListenPort 本地 TCP 监听端口（代理游戏客户端 TCP 连接）
	TCPListenPort int `json:"tcp_listen_port"`
	// HTTPListenPort 本地 HTTP 监听端口（代理游戏客户端 HTTP 请求）
	HTTPListenPort int `json:"http_listen_port"`
	// TargetServerAddr 目标 TCP 服务器地址（如 10.254.114.204:18000）
	TargetServerAddr string `json:"target_server_addr"`
	// TargetHTTPAddr 目标 HTTP 服务器地址（如 10.254.114.204:20144）
	TargetHTTPAddr string `json:"target_http_addr"`
}

// DefaultProtoTestConfig 返回默认 proto-test 配置
func DefaultProtoTestConfig() *ProtoTestConfig {
	return &ProtoTestConfig{
		TCPListenPort:    18000,
		HTTPListenPort:   20144,
		TargetServerAddr: "10.254.114.204:18000",
		TargetHTTPAddr:   "10.254.114.204:20144",
	}
}

// ProtoTestConfigService proto-test 配置管理服务
// @frontend @mcp
type ProtoTestConfigService struct{}

// NewProtoTestConfigService 创建 proto-test 配置服务实例
func NewProtoTestConfigService() *ProtoTestConfigService {
	return &ProtoTestConfigService{}
}

// GetConfig 获取当前配置
// @frontend @mcp
func (s *ProtoTestConfigService) GetConfig() (*ProtoTestConfig, error) {
	var config ProtoTestConfig
	if err := protoTestSection.Load(&config); err != nil {
		return nil, fmt.Errorf("读取配置失败: %v", err)
	}

	// section 不存在时返回默认值（不写盘，避免读操作产生副作用）
	if !protoTestSection.Exists() {
		return DefaultProtoTestConfig(), nil
	}

	// 旧配置可能缺少新字段，回退到默认值
	defaultConfig := DefaultProtoTestConfig()
	if config.TCPListenPort <= 0 || config.TCPListenPort > 65535 {
		config.TCPListenPort = defaultConfig.TCPListenPort
	}
	if config.HTTPListenPort <= 0 || config.HTTPListenPort > 65535 {
		config.HTTPListenPort = defaultConfig.HTTPListenPort
	}
	if config.TargetServerAddr == "" {
		config.TargetServerAddr = defaultConfig.TargetServerAddr
	}
	if config.TargetHTTPAddr == "" {
		config.TargetHTTPAddr = defaultConfig.TargetHTTPAddr
	}

	return &config, nil
}

// SaveConfig 保存配置
// @frontend @mcp
func (s *ProtoTestConfigService) SaveConfig(config *ProtoTestConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}
	if config.TCPListenPort <= 0 || config.TCPListenPort > 65535 {
		return fmt.Errorf("TCP 监听端口必须在 1-65535 之间")
	}
	if config.HTTPListenPort <= 0 || config.HTTPListenPort > 65535 {
		return fmt.Errorf("HTTP 监听端口必须在 1-65535 之间")
	}
	return protoTestSection.Save(config)
}

// UpdateConfig 部分更新配置
// @frontend @mcp
func (s *ProtoTestConfigService) UpdateConfig(updates map[string]any) (*ProtoTestConfig, error) {
	currentConfig, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	if v, ok := updates["tcp_listen_port"]; ok {
		currentConfig.TCPListenPort = int(toNumber(v))
	}
	if v, ok := updates["http_listen_port"]; ok {
		currentConfig.HTTPListenPort = int(toNumber(v))
	}
	if v, ok := updates["target_server_addr"]; ok {
		currentConfig.TargetServerAddr = toString(v)
	}
	if v, ok := updates["target_http_addr"]; ok {
		currentConfig.TargetHTTPAddr = toString(v)
	}

	if err := s.SaveConfig(currentConfig); err != nil {
		return nil, err
	}
	return currentConfig, nil
}

// toNumber 将 int/float64 等转为 float64
func toNumber(v any) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case float32:
		return float64(n)
	case float64:
		return n
	default:
		return 0
	}
}

// toString 将 string 或 fmt.Stringer 转为 string
func toString(v any) string {
	switch s := v.(type) {
	case string:
		return s
	default:
		return fmt.Sprintf("%v", s)
	}
}
