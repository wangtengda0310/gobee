package params

// OpenIDShortName 当前账号变量短名，由发送循环预置到 variableStore
const OpenIDShortName = "openid"

// VariableDef 变量定义，描述一个可从服务端 Ntf 中自动提取的变量
//
// 契约约束: WatchMsgIDs 必须是"请求触发型"消息 ID —— 即由本会话发出的 Req 触发服务端推送的 Ntf，
// 而非登录时自发推送型。原因: FrameMux.DrainAndStart 在池复用连接路径下会丢弃积压帧(上次会话残留)，
// 若变量依赖登录自发推送 Ntf, drain 会永久吃掉它。readLoop 启动后由本会话 Req 触发的新帧才能被缓存。
//
// AvailableReqs 限制该变量在前端卡片模式下对哪些 Req 可见（存 proto 消息名 msg_name）。
// nil 或空切片表示对所有 Req 可用（如账号级变量 openid）。
type VariableDef struct {
	ShortName     string                       // 短名，如 "cityId"
	DisplayName   string                       // UI 显示名，如 "🏰 工会城战 - 城池ID"
	WatchMsgIDs   []uint16                     // 关注的帧 MsgID 列表（必须是请求触发型 Ntf）
	ExtractFunc   func(frame any) (any, error) // 提取函数，frame 实际类型为 *DecodedFrame（使用 any 避免循环依赖）
	AvailableReqs []string                     // 对哪些 Req 可见（msg_name），nil/空=对所有 Req 可用
}

// VariableInfo 前端用的变量信息（序列化为 JSON 传给前端）
type VariableInfo struct {
	ShortName     string   `json:"short_name"`
	DisplayName   string   `json:"display_name"`
	AvailableReqs []string `json:"available_reqs,omitempty"` // 对哪些 Req 可见（msg_name），nil/空=对所有 Req 可用
}

// builtinVariables 内置变量列表
// 由 streamproto 层通过 SetBuiltinVariables 注入（因为 ExtractFunc 依赖 DecodedFrame）
var builtinVariables []VariableDef

// SetBuiltinVariables 设置内置变量列表（由 variables 包 init 调用）
func SetBuiltinVariables(vars []VariableDef) {
	builtinVariables = vars
}

// AppendBuiltinVariables 追加内置变量（用于测试补充注册，不覆盖已有变量）
func AppendBuiltinVariables(vars []VariableDef) {
	builtinVariables = append(builtinVariables, vars...)
}

// GetVariableDefs 返回所有内置变量定义
func GetVariableDefs() []VariableDef {
	return builtinVariables
}

// GetAvailableVariables 返回前端可用的变量列表
func GetAvailableVariables() []VariableInfo {
	infos := make([]VariableInfo, 0, len(builtinVariables))
	for _, def := range builtinVariables {
		infos = append(infos, VariableInfo{
			ShortName:     def.ShortName,
			DisplayName:   def.DisplayName,
			AvailableReqs: def.AvailableReqs,
		})
	}
	return infos
}

// FindVariableByShortName 根据短名查找变量定义，未找到返回 nil
func FindVariableByShortName(name string) *VariableDef {
	for i := range builtinVariables {
		if builtinVariables[i].ShortName == name {
			return &builtinVariables[i]
		}
	}
	return nil
}
