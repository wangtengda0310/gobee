package drop_rule

const (
	Id               = iota // 掉落id
	Name                    // 掉落名称
	Count                   // 掉落次数
	DropGroup               // 掉落组ID
	EnsureSmall             // 组保底次数-小
	EnsureSmallGroup        // 保底组ID-小
	EnsureBig               // 组保底次数-大
	EnsureBigGroup          // 保底组ID-大
	EnsureItemCount         // 道具保底次数
	EnsureItemID            // 保底道具ID
	ItemCheckExist          // 道具保底排重
)
