// activity-wiki-dev: 活动Wiki开发技能生成
// DrawPet 结缘亭表列索引常量
package draw_pet

const (
	Id                  = iota // 抽奖ID
	Name                       // 名称
	EnsureNewer                // 新手保底次数(M0)
	EnsureCount                // 正常保底次数(M)
	FirstEnsureNewer           // 新手保底触发次数(X)
	OnceDropRule               // 单抽掉落规则ID
	TenDropRule                // 10连抽掉落规则ID
	OnceItemCost               // 单抽消耗道具
	TenItemCost                // 10连抽消耗道具
	DailyLimit                 // 每日抽卡限制
	ActivityId                 // 活动ID
	BigAwardCount              // 大奖保底次数
	BigAwardItemId             // 大奖道具ID
	StartTime                  // 开始时间
	EndTime                    // 结束时间
	BigAwards                  // 大奖道具
	PartnerItem                // 伙伴道具
	Byproducts                 // 副产物道具ID
	ProbabilityContents        // 概率公示内容
	ProbabilityPercents        // 概率公示内容概率
	DrawPetTitleIcon           // 结缘亭界面标题图片
	DrawPetTitleDescBg         // 结缘亭标题描述底图
	DrawRuleContent            // 抽奖规则内容
)
