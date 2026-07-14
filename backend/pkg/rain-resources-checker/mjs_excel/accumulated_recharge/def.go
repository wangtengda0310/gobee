// activity-wiki-dev: 活动Wiki开发技能生成
// 功能: 累充奖励表列索引常量定义
// 关联活动类型: ActTypeAccumulatedRecharge
// 生成时间: 2026-05-05

package accumulated_recharge

const (
	Id          = iota // 奖励ID
	ActId              // 活动ID (E#EActivityId 枚举)
	RechargeNum        // 累充数
	Reward             // 充值项奖励 (ItemCfg[])
)
