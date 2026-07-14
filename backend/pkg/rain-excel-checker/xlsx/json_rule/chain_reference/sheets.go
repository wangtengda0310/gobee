// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件提供从关系链配置中提取引用表名的功能
package chain_reference

// ExtractChainStepSheets 从 CHAIN_REFERENCE 规则的 chainSteps 参数中提取所有引用的表名
// 后端自主解析，不依赖前端传递 chainRequiredSheets 参数
func ExtractChainStepSheets(params map[string]string) []string {
	chainStepsJSON, ok := params["chainSteps"]
	if !ok || chainStepsJSON == "" {
		return nil
	}

	pairConfig, err := ParseChainPairConfig(chainStepsJSON)
	if err != nil {
		return nil
	}

	seen := make(map[string]bool)
	var sheets []string

	for _, step := range pairConfig.Left.Steps {
		if step.Sheet != "" && !seen[step.Sheet] {
			seen[step.Sheet] = true
			sheets = append(sheets, step.Sheet)
		}
	}
	for _, step := range pairConfig.Right.Steps {
		if step.Sheet != "" && !seen[step.Sheet] {
			seen[step.Sheet] = true
			sheets = append(sheets, step.Sheet)
		}
	}

	return sheets
}
