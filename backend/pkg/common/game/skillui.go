package game

import (
	"strconv"
	"sync"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_ui"
)

// ExcelDirProvider 提供统一策划配表目录路径。
// 用接口而非直接依赖 settings 包，避免 game ↔ settings 循环依赖。
// 由 cmd/rain-qa-func/wails.go 在注册服务时注入 settings.ExcelConfigService。
type ExcelDirProvider interface {
	GetExcelDir() string
}

// SkillUIDescService 技能表现配置表（技能表现配置表|SkillUI）解析服务。
// 独立于 GameExcelService：GameExcelService 读取 rain-robot 编译后的 .bytes 资源，
// 本服务读取策划配表目录下的 .xlsx 文件，用于提供 SkillsTemplate 中缺失的技能描述文案。
//
// 路径来源：统一策划配表目录配置（ExcelDirProvider 注入），
// 与 hero-wiki-check / activity-wiki-check 等模块共用同一份配置，避免前端传参导致的路径不一致。
type SkillUIDescService struct {
	mu          sync.Mutex
	descMap     map[int32]string // skillId -> SkillText
	initialized bool
	dirProvider ExcelDirProvider
}

// NewSkillUIDescService 创建技能描述服务实例
func NewSkillUIDescService() *SkillUIDescService {
	return &SkillUIDescService{
		descMap: map[int32]string{},
	}
}

// SetExcelDirProvider 注入统一策划配表目录提供者（需在注册服务前调用）
func (s *SkillUIDescService) SetExcelDirProvider(p ExcelDirProvider) {
	s.dirProvider = p
}

// LoadSkillUIDesc 从统一策划配表目录解析「技能表现配置表|SkillUI」xlsx，
// 构建 skillId -> SkillText 映射并缓存。无参，路径由后端统一配置决定。
// @frontend @mcp
func (s *SkillUIDescService) LoadSkillUIDesc() (map[int32]string, error) {
	// 从注入的统一配置提供者读取策划配表目录
	excelDir := ""
	if s.dirProvider != nil {
		excelDir = s.dirProvider.GetExcelDir()
	}
	if excelDir == "" {
		s.mu.Lock()
		s.descMap = map[int32]string{}
		s.initialized = true
		s.mu.Unlock()
		return map[int32]string{}, nil
	}

	sheetMap, err := excelio.GetSheetMap(excelDir)
	if err != nil {
		// 目录未配置或为空时返回空 map，不阻塞前端用例加载
		s.mu.Lock()
		s.descMap = map[int32]string{}
		s.initialized = true
		s.mu.Unlock()
		return map[int32]string{}, nil
	}
	defer func() {
		// sheetMap 中的 *excelize.File 仅用于本次解析，用完即关
		seen := map[interface{}]struct{}{}
		for _, f := range sheetMap {
			if _, ok := seen[f]; ok {
				continue
			}
			seen[f] = struct{}{}
			_ = f.Close()
		}
	}()

	skillUIDiffs, err := skill_ui.GetSkillUIDiffMap(sheetMap)
	if err != nil {
		return map[int32]string{}, nil
	}

	// 解析技能表，建立 拼音(ESkillId) -> 数字Id 映射。
	// SkillUI 表的 Id 列存的是拼音（如 "Tao"），不是数字；需要借技能表把拼音转成数字技能ID。
	skillDiffs, err := skill.GetSkillDiffMap(sheetMap)
	if err != nil {
		return map[int32]string{}, nil
	}
	pinyinToId := make(map[string]int32, len(*skillDiffs))
	for _, sd := range *skillDiffs {
		if id, err := strconv.Atoi(sd.Id); err == nil && sd.ESkillId != "" {
			pinyinToId[sd.ESkillId] = int32(id)
		}
	}

	// SkillUI.Id 是拼音，通过 pinyinToId 转成数字技能ID；ESkillId 兜底匹配
	descMap := make(map[int32]string, len(*skillUIDiffs))
	for _, ui := range *skillUIDiffs {
		if ui.SkillText == "" {
			continue
		}
		key := ui.Id
		if key == "" {
			key = ui.ESkillId
		}
		numericId, ok := pinyinToId[key]
		if !ok {
			continue
		}
		descMap[numericId] = ui.SkillText
	}

	s.mu.Lock()
	s.descMap = descMap
	s.initialized = true
	s.mu.Unlock()

	return descMap, nil
}

// GetSkillUIDescMap 获取已加载的技能描述映射（调用前需先调用 LoadSkillUIDesc）
// @frontend @mcp
func (s *SkillUIDescService) GetSkillUIDescMap() map[int32]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.descMap == nil {
		return map[int32]string{}
	}
	return s.descMap
}

// IsInitialized 是否已加载
func (s *SkillUIDescService) IsInitialized() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.initialized
}
