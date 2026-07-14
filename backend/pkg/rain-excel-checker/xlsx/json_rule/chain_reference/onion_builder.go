// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件实现洋葱模型的链构建器
// BuildOnionChain 根据两链配置和参数，构建从 Validate → Compare → 左链 → Match → 右链(反向) 的执行链
package chain_reference

// ==================== 洋葱模型构建器 ====================

// BuildOnionChain 从配置构建洋葱链，返回入口执行函数
// 构建顺序（从内到外包裹）：
//  1. terminal（最内层 no-op）
//  2. 右链步骤反向（StepN, ..., Step0）— 从内到外包裹
//  3. Match handler（两阶段门控）
//  4. 左链步骤正向（Step0, ..., StepM）— 从内到外包裹
//  5. Compare handler
//  6. Validate handler（最外层）
//
// 执行顺序（从外到内）：
//
//	Validate → Compare → LeftStep0 → ... → LeftStepM → Match → RightStep0 → ... → RightStepN → terminal
//	   ↑                                                                                              ↓
//	   └────────────────────── 结果回传（Violation/Reason）←────────────────────────────────────────────┘
//
// 参数：
//   - config: 两链配置（ChainPairConfig）
//   - params: 规则参数（chainCompare、chainMatchCompare 等）
//
// 返回：
//   - 入口执行函数，接收 ChainContext 并执行整条洋葱链
func BuildOnionChain(config *ChainPairConfig, params map[string]string) func(ctx *ChainContext) error {
	// 步骤 1: 解析参数
	_ = params["chainCompare"]
	_ = params["chainMatchCompare"]

	// 步骤 2: 构建终端函数（最内层 no-op）
	// 当执行到终端时，洋葱模型已从外到内穿透所有层，准备从内到外回传结果
	terminal := func(ctx *ChainContext) error {
		return nil
	}

	// 步骤 3: 从内到外构建 handler 链
	// Go 中实现洋葱模型的惯用方式：每个 handler 通过闭包捕获 next 函数
	// wrapHandler 将一个 ChainHandler 接口包装为 NextFunc，形成调用链
	current := terminal

	// 步骤 3a: 右链步骤反向（从内到外：StepN, ..., Step0）
	// 反向遍历右链步骤，每步包装为 handler
	// 执行顺序：Match → RightStep0 → RightStep1 → ... → RightStepN → terminal
	// 数据流：Match 设置 RightCurrentValues → StepN 反向查找 → ... → Step0 反向查找
	for i := len(config.Right.Steps) - 1; i >= 0; i-- {
		step := config.Right.Steps[i]
		handler := &RightStepHandler{
			Step:     step,
			StepIdx:  i,
			IsFirst:  i == 0,
			ChainCfg: config.Right,
		}
		current = wrapHandler(handler, current)
	}

	// 步骤 3b: Match handler（两阶段门控）
	// 比较左链最后一步与右链最终表的值，匹配成功时将交集键传递给右链反向步骤
	matchCompareType := params["chainMatchCompare"]
	if matchCompareType != "" && len(config.Right.Steps) > 0 {
		rightLastStep := config.Right.Steps[len(config.Right.Steps)-1]
		matchHandler := &MatchHandler{
			MatchType:     matchCompareType,
			RightLastStep: rightLastStep,
			RightConfig:   config.Right,
		}
		current = wrapHandler(matchHandler, current)
	}

	// 步骤 3c: 左链步骤正向（从外到内：Step0 先执行）
	// 反向遍历（从 M 到 0），使 Step0 成为最外层（最先执行 Handle）
	// 执行顺序：Step0.Handle → Step1.Handle → ... → StepM.Handle → Match
	for i := len(config.Left.Steps) - 1; i >= 0; i-- {
		step := config.Left.Steps[i]
		handler := &LeftStepHandler{
			Step:     step,
			StepIdx:  i,
			IsLast:   i == len(config.Left.Steps)-1,
			ChainCfg: config.Left,
		}
		current = wrapHandler(handler, current)
	}

	// 步骤 3d: Compare handler
	compareType := params["chainCompare"]
	hasLeftCompare := config.Left.CompareCol != ""
	hasRightCompare := config.Right.CompareCol != ""
	useTimeOverlap := compareType == "time_overlap" && hasLeftCompare && hasRightCompare
	compareHandler := &CompareHandler{
		CompareType:    compareType,
		LeftKey:        params["chainLeftKey"],
		RightKey:       params["chainRightKey"],
		UseTimeOverlap: useTimeOverlap,
	}
	current = wrapHandler(compareHandler, current)

	// 步骤 3e: Validate handler（最外层，最先执行）
	// 校验配置参数的合法性
	validate := &ValidateHandler{}
	current = wrapHandler(validate, current)

	return current
}

// wrapHandler 将 ChainHandler 接口包装为 NextFunc
// 这是洋葱模型的核心串联机制：
//   - 返回的 NextFunc 在被调用时，先执行 handler 的自身逻辑
//   - handler 在适当位置调用 next（内层函数），将控制权传递给内层
//   - 内层返回后，handler 可以处理内层的结果
//
// 参数：
//   - handler: 当前层的处理器
//   - next: 内层执行函数
func wrapHandler(handler ChainHandler, next NextFunc) NextFunc {
	return func(ctx *ChainContext) error {
		return handler.Handle(ctx, next)
	}
}
