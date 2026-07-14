package src

type checkResult interface {
	String() string
}

// 飞书消息中追加生效规则统计
// 汇总、摘要
// merge
// - []commit
//   - []file
//   - []rule
//
// 前端组件中展示生效的否则统计
type summary struct {
	sheetRules []sheetRule
	colRules   []colRule
	rowRules   []rowRule
	multiRules []multiSheetRule
}

// 应该在feishu-lib下
type feishu struct {
}

// 应该在当前位置
type console struct {
}

// 应该在 rain-qa-func下
type intercepter struct {
}

// 前端组件交互数据
// - 列出规则
// - 增加json规则
// - 校验结果
type panelDTO interface {
}
