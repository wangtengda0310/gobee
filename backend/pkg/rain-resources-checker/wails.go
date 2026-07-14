package resourcechecker

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/hero_res_check"
)

// HeroResCheckService 武将语音资源检查服务
type HeroResCheckService struct{}

// NewHeroResCheckService 创建武将语音资源检查服务
func NewHeroResCheckService() *HeroResCheckService {
	return &HeroResCheckService{}
}

// Check 检查武将语音资源
// cardDir 表示同级目录的../../client/Master/Card/位置
// @frontend
func (h *HeroResCheckService) Check(excelDir, cardDir string) (*hero_res_check.VoiceCheckReport, error) {
	rep, err := hero_res_check.CheckGitlabAndLocalVoicesInExcel(excelDir, cardDir)
	if err != nil {
		return nil, err
	}
	return rep, nil
}
