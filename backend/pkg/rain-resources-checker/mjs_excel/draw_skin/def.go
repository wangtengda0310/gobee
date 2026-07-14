package draw_skin

// DrawSkin 表列索引常量
// 注意：实际 Excel 表在 Name(1) 和 OnceDropRule(5) 之间有 3 个额外列，
// 不能使用 iota 连续递增，必须用明确的列索引
const (
	Id               = 0  // 抽奖ID
	Name             = 1  // 抽奖名称
	EnsureNewer      = 2  // 新手保底次数(M0)
	EnsureCount      = 3  // 正常保底次数(M)
	FirstEnsureNewer = 4  // 新手保底触发次数(X)
	OnceDropRule     = 5  // 单抽掉落规则ID
	TenDropRule      = 6  // 10连抽掉落规则ID
	OnceItemCost     = 7  // 单抽消耗道具
	TenItemCost      = 8  // 10连抽消耗道具
	DailyLimit       = 9  // 每日抽卡限制
	ActivityId       = 10 // 活动ID
	BigAwardCount    = 11 // 大奖保底次数
	BigAwardItemId   = 12 // 大奖道具ID
	StartTime        = 13 // 开始时间
	EndTime          = 14 // 结束时间
)
