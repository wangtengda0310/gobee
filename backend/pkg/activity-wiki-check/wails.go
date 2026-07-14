package activitywikicheck

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/appconfig"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/fileutils"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/engine"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/ruleconfig"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/activitywiki"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel"
)

// ActivityWikiCheckService 活动Wiki检查服务
// @frontend
type ActivityWikiCheckService struct{}

// ActivityWikiConfigService 活动 Wiki 配置管理服务
type ActivityWikiConfigService struct {
	section *appconfig.Section
}

// ActivityWikiConfig 应用配置结构
type ActivityWikiConfig struct {
	ExcelDir   string `json:"excel_dir"`
	ClientPath string `json:"client_path"`
}

// NewActivityWikiConfigService 创建配置服务实例
func NewActivityWikiConfigService() *ActivityWikiConfigService {
	return &ActivityWikiConfigService{
		section: appconfig.New("activity_wiki"),
	}
}

// getDefaultConfig 获取默认配置
func (cs *ActivityWikiConfigService) getDefaultConfig() *ActivityWikiConfig {
	return &ActivityWikiConfig{
		ExcelDir: "../../config/excel",
	}
}

// GetConfig 获取当前配置
// @frontend @mcp
func (cs *ActivityWikiConfigService) GetConfig() (*ActivityWikiConfig, error) {
	var config ActivityWikiConfig
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
func (cs *ActivityWikiConfigService) SaveConfig(config *ActivityWikiConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}
	return cs.section.Save(config)
}

// UpdateConfig 更新配置（部分更新）
func (cs *ActivityWikiConfigService) UpdateConfig(updates map[string]any) (*ActivityWikiConfig, error) {
	// 获取当前配置
	currentConfig, err := cs.GetConfig()
	if err != nil {
		return nil, err
	}

	// 临时转换为 map 以便更新
	configMap := make(map[string]any)
	data, _ := json.Marshal(currentConfig)
	json.Unmarshal(data, &configMap)

	// 应用更新
	for key, value := range updates {
		if _, exists := configMap[key]; exists {
			configMap[key] = value
		}
	}

	// 转换回 ActivityWikiConfig 结构体
	updatedData, _ := json.Marshal(configMap)
	var updatedConfig ActivityWikiConfig
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
func (cs *ActivityWikiConfigService) ResetToDefault() (*ActivityWikiConfig, error) {
	defaultConfig := cs.getDefaultConfig()
	if err := cs.SaveConfig(defaultConfig); err != nil {
		return nil, err
	}
	return defaultConfig, nil
}

// GetConfigFilePath 获取配置文件路径
func (cs *ActivityWikiConfigService) GetConfigFilePath() string {
	return appconfig.FilePath()
}

// ConfigExists 检查配置是否存在
func (cs *ActivityWikiConfigService) ConfigExists() bool {
	return cs.section.Exists()
}

// DeleteConfig 删除配置
func (cs *ActivityWikiConfigService) DeleteConfig() error {
	return cs.section.Delete()
}

// NewActivityWikiCheckService 创建活动Wiki检查服务
func NewActivityWikiCheckService() *ActivityWikiCheckService {
	return &ActivityWikiCheckService{}
}

// Check 检查活动Wiki数据
// excelDir: Excel配置目录路径
// @frontend @mcp
func (a *ActivityWikiCheckService) Check(excelDir string) (*diff.DataContainer, error) {
	sheetMap, err := excelio.GetSheetMap(excelDir)
	if err != nil {
		return nil, err
	}
	excels, err := mjs_excel.InitDiffRefExcel(sheetMap)
	if err != nil {
		return nil, err
	}

	excels.ActivityWikiDiff = activitywiki.BuildActivityWikiDiff(excels)

	return excels, nil
}

// OpenExcel 打开指定Sheet对应的Excel文件
// sheetName: Sheet名称（如"活动表|Activity"），为空时直接打开 excelDir 作为文件路径
// excelDir: Excel配置目录路径，或当 sheetName 为空时为文件完整路径
// @frontend
func (a *ActivityWikiCheckService) OpenExcel(sheetName string, excelDir string) error {
	return fileutils.OpenExcel(sheetName, excelDir)
}

// FieldRuleStat 字段级规则统计
type FieldRuleStat struct {
	ColRuleCount int            `json:"colRuleCount"` // 列级规则数量（仅Enabled）
	RowRuleCount int            `json:"rowRuleCount"` // 行级规则数量（预留，当前为0）
	TotalCount   int            `json:"totalCount"`   // = ColRuleCount + RowRuleCount
	ErrorCount   int            `json:"errorCount"`   // 校验错误数量（执行检查后填充）
	RuleNames    []RuleNameInfo `json:"ruleNames"`    // 规则名称列表
}

// RuleNameInfo 规则名称信息（用于前端Popover展示）
type RuleNameInfo struct {
	Type        string `json:"type"`        // 规则类型标识（如 DRAWSKIN_DATA_VALIDITY_CHECK）
	DisplayName string `json:"displayName"` // 规则显示名称
}

// SheetRuleStats 单个Sheet的规则统计
type SheetRuleStats struct {
	SheetName       string                    `json:"sheetName"`
	TableRuleCount  int                       `json:"tableRuleCount"`  // 表级规则数（仅Enabled）
	TableRuleNames  []RuleNameInfo            `json:"tableRuleNames"`  // 表级规则名称列表
	TableErrorCount int                       `json:"tableErrorCount"` // 表级规则校验错误数
	FieldRuleStats  map[string]*FieldRuleStat `json:"fieldRuleStats"`  // key=字段名
}

// RuleCoverageData 规则覆盖数据
type RuleCoverageData struct {
	Sheets map[string]*SheetRuleStats `json:"sheets"` // key=Sheet名
}

// GetRuleCoverage 获取指定目录下所有Sheet的规则覆盖统计
// 用于前端角标显示，统计每个Sheet启用了多少表级规则和列级规则
// caseDir: cases/excel_cases 目录路径
// @frontend
func (a *ActivityWikiCheckService) GetRuleCoverage(caseDir string) (*RuleCoverageData, error) {
	// 加载JSON规则文件
	sheetRules, err := ruleconfig.LoadJsonRules(caseDir)
	if err != nil {
		return nil, fmt.Errorf("加载规则文件失败: %v", err)
	}

	result := &RuleCoverageData{
		Sheets: make(map[string]*SheetRuleStats),
	}

	// 用于通过短名快速查找Sheet统计
	shortNameIndex := make(map[string]string)

	// 遍历每个SheetRule统计规则
	for _, sr := range sheetRules {
		if sr == nil {
			continue
		}

		stats := &SheetRuleStats{
			SheetName:      sr.Sheet,
			FieldRuleStats: make(map[string]*FieldRuleStat),
		}

		// 统计表级规则（仅Enabled），同时收集规则名称
		enabledTableCount := 0
		var tableRuleNames []RuleNameInfo
		for _, tr := range sr.TableRules {
			if tr != nil && tr.Enabled {
				enabledTableCount++
				tableRuleNames = append(tableRuleNames, RuleNameInfo{
					Type:        string(tr.Type),
					DisplayName: tr.DisplayName,
				})
			}
		}
		stats.TableRuleCount = enabledTableCount
		stats.TableRuleNames = tableRuleNames

		// 统计列级规则（仅Enabled），同时收集规则名称
		for fieldName, colRule := range sr.Rules {
			if colRule == nil {
				continue
			}
			enabledColCount := 0
			var fieldRuleNames []RuleNameInfo
			for _, cr := range colRule.PropRules {
				if cr != nil && cr.Enabled {
					enabledColCount++
					fieldRuleNames = append(fieldRuleNames, RuleNameInfo{
						Type:        string(cr.Type),
						DisplayName: cr.DisplayName,
					})
				}
			}
			stats.FieldRuleStats[fieldName] = &FieldRuleStat{
				ColRuleCount: enabledColCount,
				RowRuleCount: 0, // 预留
				TotalCount:   enabledColCount,
				RuleNames:    fieldRuleNames,
			}
		}

		result.Sheets[sr.Sheet] = stats

		// 建立短名索引，用于后续匹配DefaultTableRules
		parts := strings.Split(sr.Sheet, "|")
		if len(parts) >= 2 {
			shortNameIndex[parts[len(parts)-1]] = sr.Sheet
		} else {
			shortNameIndex[sr.Sheet] = sr.Sheet
		}
	}

	// 补充默认表级规则
	for shortName, defaultRules := range json_rule.DefaultTableRules {
		fullName, ok := shortNameIndex[shortName]
		if !ok {
			// 没有对应的SheetRule，创建新的统计条目
			var names []RuleNameInfo
			for _, rt := range defaultRules {
				displayName := string(rt)
				if meta, ok2 := json_rule.TableRuleMetas[rt]; ok2 && meta.DisplayName != "" {
					displayName = meta.DisplayName
				}
				names = append(names, RuleNameInfo{
					Type:        string(rt),
					DisplayName: displayName,
				})
			}
			result.Sheets[shortName] = &SheetRuleStats{
				SheetName:      shortName,
				TableRuleCount: len(defaultRules),
				TableRuleNames: names,
				FieldRuleStats: make(map[string]*FieldRuleStat),
			}
			continue
		}
		// 已有SheetRule，累加默认规则数量
		existing := result.Sheets[fullName]
		existing.TableRuleCount += len(defaultRules)
		for _, rt := range defaultRules {
			displayName := string(rt)
			if meta, ok2 := json_rule.TableRuleMetas[rt]; ok2 && meta.DisplayName != "" {
				displayName = meta.DisplayName
			}
			existing.TableRuleNames = append(existing.TableRuleNames, RuleNameInfo{
				Type:        string(rt),
				DisplayName: displayName,
			})
		}
	}

	return result, nil
}

// GetRuleCoverageWithErrors 获取规则覆盖统计并合并校验错误计数
// 先获取基础覆盖数据，再执行 Excel 规则检查，将错误数按 sheet+field 粒度合并
// caseDir: cases/excel_cases 目录路径
// excelDir: Excel 配置目录路径
// @frontend
func (a *ActivityWikiCheckService) GetRuleCoverageWithErrors(caseDir string, excelDir string) (*RuleCoverageData, error) {
	result, err := a.GetRuleCoverage(caseDir)
	if err != nil {
		return nil, err
	}

	// 从缓存读取配表检查结果（由配表测试页面执行检查后写入）
	colResults, tableResults := engine.GetCachedCheckResults()

	// [DEBUG] 日志
	fmt.Printf("[GetRuleCoverageWithErrors] 规则覆盖sheets=%d, 缓存ColResults=%d, TableResults=%d\n",
		len(result.Sheets), len(colResults), len(tableResults))

	sheetKeys := make([]string, 0, len(result.Sheets))
	for k := range result.Sheets {
		sheetKeys = append(sheetKeys, k)
	}
	fmt.Printf("[GetRuleCoverageWithErrors] 规则覆盖Sheet列表: %v\n", sheetKeys)

	matchedCount := 0
	unmatchedCount := 0

	// 列级检查结果：按 sheet+field 统计错误数
	for _, cr := range colResults {
		if cr == nil || len(cr.ErrCells) == 0 {
			continue
		}
		sheetName := ""
		if cr.SheetName != nil {
			sheetName = *cr.SheetName
		}
		fieldName := ""
		if cr.ColName != nil {
			fieldName = *cr.ColName
		}

		stats, ok := result.Sheets[sheetName]
		if !ok {
			unmatchedCount++
			fmt.Printf("[GetRuleCoverageWithErrors] 列级错误未匹配Sheet: sheet=%q, field=%q, errCells=%d\n",
				sheetName, fieldName, len(cr.ErrCells))
			continue
		}

		errCount := len(cr.ErrCells)
		if fieldName != "" {
			if fs, ok := stats.FieldRuleStats[fieldName]; ok {
				fs.ErrorCount += errCount
				matchedCount++
				fmt.Printf("[GetRuleCoverageWithErrors] 字段错误匹配: sheet=%q, field=%q, +err=%d -> total=%d\n",
					sheetName, fieldName, errCount, fs.ErrorCount)
			} else {
				stats.TableErrorCount += errCount
				fmt.Printf("[GetRuleCoverageWithErrors] 字段无规则: sheet=%q, field=%q, +tableErr=%d\n",
					sheetName, fieldName, errCount)
			}
		} else {
			stats.TableErrorCount += errCount
		}
	}

	// 表级检查结果：按 sheet 统计错误数
	for _, tr := range tableResults {
		if tr == nil {
			continue
		}
		sheetName := ""
		if tr.SheetName != nil {
			sheetName = *tr.SheetName
		}
		errCount := len(tr.ErrCells)
		if errCount == 0 {
			continue
		}
		fmt.Printf("[GetRuleCoverageWithErrors] 表级错误: sheet=%q, errCells=%d\n", sheetName, errCount)
		if stats, ok := result.Sheets[sheetName]; ok {
			stats.TableErrorCount += errCount
		}
	}

	fmt.Printf("[GetRuleCoverageWithErrors] 合并完成: matched=%d, unmatched=%d\n", matchedCount, unmatchedCount)

	return result, nil
}

// ResourceCheckResult 单个资源路径的检查结果
type ResourceCheckResult struct {
	Path       string `json:"path"`       // 原始路径
	Exists     bool   `json:"exists"`     // 文件是否存在
	PreviewUrl string `json:"previewUrl"` // base64 data URL（仅图片文件存在时）
}

// CheckResourcePaths 批量检查资源文件是否存在，返回检查结果
// paths: 资源路径列表（如 "Assets/Art/Icons/xxx.png"）
// clientPath: 客户端项目根目录路径（如 "D:/work/client/Master/Card"）
// @frontend
func (a *ActivityWikiCheckService) CheckResourcePaths(paths []string, clientPath string) ([]*ResourceCheckResult, error) {
	if clientPath == "" {
		return nil, fmt.Errorf("clientPath is required")
	}

	// 支持的图片扩展名
	imageExts := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
		".webp": true,
		".bmp":  true,
	}

	results := make([]*ResourceCheckResult, 0, len(paths))

	for _, p := range paths {
		absPath := filepath.Join(clientPath, p)
		result := &ResourceCheckResult{
			Path: p,
		}

		info, err := os.Stat(absPath)
		if err != nil || info.IsDir() {
			result.Exists = false
			results = append(results, result)
			continue
		}

		result.Exists = true

		// 检查是否为图片文件，生成 base64 预览（限制 100KB 以内）
		ext := strings.ToLower(filepath.Ext(absPath))
		if imageExts[ext] && info.Size() <= 100*1024 {
			data, err := os.ReadFile(absPath)
			if err == nil {
				// jpg/jpeg 使用 image/jpeg，其他使用 image/{ext}
				mimeExt := ext
				if ext == ".jpg" || ext == ".jpeg" {
					mimeExt = ".jpeg"
				}
				mimeType := "image/" + mimeExt[1:] // 去掉点号
				result.PreviewUrl = "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data)
			}
		}

		results = append(results, result)
	}

	return results, nil
}
