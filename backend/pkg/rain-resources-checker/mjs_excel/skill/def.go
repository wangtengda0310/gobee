package skill

const (
	Id                = iota // 技能ID
	SkillName                // 技能名称
	ShortSkillName           // 技能缩写名称
	ESkillId                 // 技能枚举定义
	SkillType                // 技能类型
	IsFromOther              // 其他人的技能
	CounterFormula           // 计数器公式
	IsFromAura               // 来自光环
	TransCard                // 转换卡牌类型
	SkillFromType            // 技能来源类型
	ResetCounterType         // 技能重置计数类型
	ResetTimesType           // 技能重置使用次数类型
	SkillLimitTimes          // 技能限定次数
	TotalLimitTimes          // 每局游戏限制次数
	TriggerCondition         // 触发时机
	SkillEffect              // 技能效果
	EmptyWaitTime            // 空节点等待时间
	Buff                     // Buff ID
	ShowPointAndAttr         // 响应时显示点数与花色
	MutexSkill               // 互斥技能ID列表
	BattleCardClass          // 战斗标记牌类型
	AIJudgeArea              // 大模型区域判定参数
	DeadType                 // 阵亡类型
	InitPro                  // 初始属性
	MagicSkillID             // 光环技能ID
	IsAutoSelAllOther        // 是否自动选中所有其他玩家
	IsForbidCopy             // 是否禁止复制（刻写）
	IsForbidTrans            // 是否禁止转移
	IsForbidDestroy          // 是否禁止废除
)
