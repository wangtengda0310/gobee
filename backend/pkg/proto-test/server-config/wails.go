package serverconfig

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings"
)

// ServerConfigService 为 Server Config 前端页面提供扩展能力。
// 当前主要封装 Unity 服务器列表注入与客户端配置导出。
// @frontend
type ServerConfigService struct {
	excelConfigSvc *settings.ExcelConfigService
}

// NewServerConfigService 创建 ServerConfigService 实例
func NewServerConfigService(excelConfigSvc *settings.ExcelConfigService) *ServerConfigService {
	return &ServerConfigService{
		excelConfigSvc: excelConfigSvc,
	}
}

// resolveExcelDir 解析策划配表目录：优先使用传入值，其次读取统一配置
func (s *ServerConfigService) resolveExcelDir(excelDir string) (string, error) {
	if excelDir != "" {
		return excelDir, nil
	}
	if s.excelConfigSvc == nil {
		return "", ErrExcelDirRequired
	}
	cfg, err := s.excelConfigSvc.GetConfig()
	if err != nil {
		return "", err
	}
	if cfg == nil || cfg.ExcelDir == "" {
		return "", ErrExcelDirRequired
	}
	return cfg.ExcelDir, nil
}

// InjectUnityServer 向策划配表的服务器配置表.xlsx 中写入或更新 Unity 服务器条目。
// 当 cfg.ExcelDir 为空时，自动从统一策划配表目录配置读取。
// 当 cfg.IpPort 为空时，自动使用本机IP和 cfg.HTTPListenPort 构造。
// @frontend @mcp
func (s *ServerConfigService) InjectUnityServer(cfg ServerXlsxConfig) error {
	excelDir, err := s.resolveExcelDir(cfg.ExcelDir)
	if err != nil {
		return err
	}
	cfg.ExcelDir = excelDir
	return InjectUnityServer(cfg)
}

// ExportClientConfig 在策划配表目录下执行客户端配置导出批处理 export_client.bat。
// 当 excelDir 为空时，自动从统一策划配表目录配置读取。
// @frontend @mcp
func (s *ServerConfigService) ExportClientConfig(excelDir string) error {
	excelDir, err := s.resolveExcelDir(excelDir)
	if err != nil {
		return err
	}
	return ExportClientConfig(excelDir)
}
