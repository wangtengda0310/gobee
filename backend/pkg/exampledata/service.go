package exampledata

import (
	"fmt"
	"path/filepath"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/appconfig"
	internal "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg"
)

// ExampleDataService 内置示例数据服务(独立,方便推翻删除)。
// 仅 Android 端用(桌面端用户有真实数据目录)。
type ExampleDataService struct{}

// New 创建示例数据服务
func New() *ExampleDataService { return &ExampleDataService{} }

// LoadResult 加载示例结果
type LoadResult struct {
	JsonsDir       string `json:"jsonsDir"`       // 战斗用例目录(配置 JsonsDir 指向)
	ResourcesDir   string `json:"resourcesDir"`   // game .bytes 目录(配置 ExcelResourcesDir 指向)
	FightCaseCount int    `json:"fightCaseCount"` // 释放的战斗用例数
	FileCount      int    `json:"fileCount"`      // 释放的总文件数(.bytes + 用例)
}

// LoadExampleData 释放内置示例数据到私有目录,并配置 function-test 指向。
// 系统干净:先清旧 example-data/ 再释放(多次加载不累积)。
// 调用后用户导航到"战斗测试"页,页面 mount 读新配置加载用例树 + game resources。
// @frontend
func (s *ExampleDataService) LoadExampleData() (*LoadResult, error) {
	// 释放目录:配置文件同级的 example-data/(Android 即 files/example-data/)
	destDir := filepath.Join(filepath.Dir(appconfig.FilePath()), "example-data")
	count, err := ReleaseTo(destDir)
	if err != nil {
		return nil, fmt.Errorf("释放示例数据失败: %w", err)
	}
	jsonsDir := filepath.Join(destDir, "fight_cases")
	resourcesDir := filepath.Join(destDir, "resources")

	// 改 function_test 配置(保留其他字段,仅覆盖两路径)
	section := appconfig.New("function_test")
	cfg := &internal.FuncCaseConfig{}
	if section.Exists() {
		_ = section.Load(cfg) // 读现有配置,失败用零值(不阻断)
	}
	cfg.JsonsDir = jsonsDir
	cfg.ExcelResourcesDir = resourcesDir
	if err := section.Save(cfg); err != nil {
		return nil, fmt.Errorf("保存配置失败: %w", err)
	}

	return &LoadResult{
		JsonsDir:       jsonsDir,
		ResourcesDir:   resourcesDir,
		FightCaseCount: countFightCases(),
		FileCount:      count,
	}, nil
}

// countFightCases 统计 embed 内战斗用例数(固定 10 个,从 embed 实算避免硬编码漂移)
func countFightCases() int {
	entries, err := embedded.ReadDir("embed/fight_cases")
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() {
			n++
		}
	}
	return n
}
