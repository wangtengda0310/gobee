// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 关系链用于跨表多跳转查找，支持动态 N 步跳转和多种比较模式
package chain_reference

// ==================== 关系链数据结构 ====================

// ChainStep 关系链的单个跳转步骤
// 描述从当前值出发，在目标表中查找匹配行并提取新值的过程
type ChainStep struct {
	Sheet         string `json:"sheet"`         // 目标表名（如 "道具表|Item" 或 "Item"）
	PreCol        string `json:"preCol"`        // 在目标表中用哪列匹配
	FindVal       string `json:"findVal"`       // 匹配值来源："self"=上一步结果, "col"=当前行某列
	NextCol       string `json:"nextCol"`       // 匹配成功后提取哪列的值作为下一步输入（左链第一步时可指定当前表的列名来取值）
	Pattern       string `json:"pattern"`       // 正则提取模式（可选，用于解析 {123;456} 等格式）
	Groups        string `json:"groups"`        // 正则捕获组（可选，逗号分隔，如 "1" 或 "1,2"）
	FilterCol     string `json:"filterCol"`     // 仅当目标表某列值等于 FilterVal 时才匹配（可选）
	FilterVal     string `json:"filterVal"`     // 过滤条件值（可选）
	IsArray       string `json:"isArray"`       // 是否按逗号拆分提取值（"true"/"false"，保留花括号内容）
	FilterIsArray string `json:"filterIsArray"` // 过滤值是否为多值（"true" 时按逗号拆分 filterVal）
	FilterMode    string `json:"filterMode"`    // 过滤模式: ""(单值) | "multi"(多值) | "withinDays"(距今<N天)
	FilterDays    string `json:"filterDays"`    // 距今天数（withinDays 模式使用）
}

// ChainConfig 关系链完整配置
// 一条关系链由多个跳转步骤组成，最终提取指定列的值用于比较
type ChainConfig struct {
	Steps      []ChainStep `json:"steps"`      // 跳转步骤列表
	CompareCol string      `json:"compareCol"` // 比较列名（可选，用于从最终匹配行提取比较值）
}

// ChainValue 关系链最终提取的单个值对
// Match 是通过关系链提取的值，Compare 是可选的比较值
type ChainValue struct {
	Match   string // 通过关系链提取的匹配值
	Compare string // 可选的比较值（当 CompareCol 非空时提取）
}

// ChainResult 关系链执行结果
// 包含所有提取的值对，提供便捷的查询方法
type ChainResult struct {
	Values               []ChainValue // 提取的值对列表
	StepValues           [][]string   // 每一步的 nextValues（StepValues[0]=第一步, StepValues[last]=最后一步）
	FirstStepInputValues []string     // 第一步的查找值（正则提取后、在目标表中查找的值），用于两阶段比较的 Phase 2 右侧值
}

// MatchValues 获取所有 Match 值
// 返回所有通过关系链提取的匹配值列表
func (r *ChainResult) MatchValues() []string {
	if r == nil {
		return nil
	}
	result := make([]string, len(r.Values))
	for i, v := range r.Values {
		result[i] = v.Match
	}
	return result
}

// FirstStepValues 获取第一步 nextValues
// nil 安全：接收者为 nil 时返回 nil
func (r *ChainResult) FirstStepValues() []string {
	if r == nil || len(r.StepValues) == 0 {
		return nil
	}
	return r.StepValues[0]
}

// LastStepValues 获取最后一步 nextValues
// nil 安全：接收者为 nil 时返回 nil
// 等价于 MatchValues()（由 ExecuteChain 实现保证）
func (r *ChainResult) LastStepValues() []string {
	if r == nil || len(r.StepValues) == 0 {
		return nil
	}
	return r.StepValues[len(r.StepValues)-1]
}

// GetFirstStepInputValues 获取第一步的查找值（正则提取后、在目标表中查找的值）
// 用于两阶段比较的 Phase 2 右侧值
// nil 安全：接收者为 nil 时返回 nil
func (r *ChainResult) GetFirstStepInputValues() []string {
	if r == nil {
		return nil
	}
	return r.FirstStepInputValues
}

// GetCompareValues 获取指定匹配键对应的比较值列表
// 支持模糊匹配：当 matchKey 为空或包含通配符时，返回所有匹配的 Compare 值
func (r *ChainResult) GetCompareValues(matchKey string) []string {
	if r == nil {
		return nil
	}

	// 如果 matchKey 为空，返回所有非空的 Compare 值
	if matchKey == "" {
		var result []string
		for _, v := range r.Values {
			if v.Compare != "" {
				result = append(result, v.Compare)
			}
		}
		return result
	}

	// 精确匹配：返回指定 matchKey 对应的 Compare 值
	for _, v := range r.Values {
		if v.Match == matchKey {
			if v.Compare == "" {
				return nil
			}
			return []string{v.Compare}
		}
	}

	return nil
}

// ChainPairConfig 完整的两链配置（从 JSON 参数解析）
type ChainPairConfig struct {
	Left  ChainConfig `json:"left"`  // 来源链（参考数据链）
	Right ChainConfig `json:"right"` // 目标链（被检查数据链）
}
