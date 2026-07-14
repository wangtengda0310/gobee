package buff

const (
	Id                     = iota // BUffBuff
	EBuffId                       // E#EBuffId
	Name                          // BUff名称
	NeedRecord                    // 战报记录
	BuffType                      // Buff类型
	IsDeleteByCasterDead          // 施法者死亡是否移除
	IsDeleteByExecutorDead        // 执行者死亡是否移除
	IsValidByFengJin              // 封禁技能后buff是否仍然有效
	IsReserveByRemoveSkill        // 删除技能时是否保留
	TransferSkillBuffType         // 转移技能时buff转移类型
	Round                         // 生效回合
	Value                         // 生效次数
	EndByEffect                   // 结束类型（生效次数为0是否清除）
	EndType                       // 结束类型
	IsServerOnly                  // 是否服务器专用
	IsCasterOnly                  // 是否仅施法者可见
	OwnerType                     // BuffOwner的类型
	ShowArea                      // 客户端显示区域
	Icon                          // Buff图标
	Effect                        // BuffLoop特效
	FlashEffect                   // Buff闪动特效
	BuffPriority                  // Buff显示优先级
	OverlyingType                 // 叠加类型
	IsTrigger                     // 是否一定触发
	BuffState                     // Buff携带的状态
	BuffPro                       // Buff修改的属性
	ProValue                      // Buff修改属性的值
	CostEffectValue               // 消耗的生效次数
	TriggerTiming                 // 触发时机
	TriggerPriority               // 触发优先级
	TriggerCondition              // 触发条件
	TriggerAction                 // 触发行为
	BuffDot                       // 是否是DOT
	EffectDescribe                // 效果描述
)
