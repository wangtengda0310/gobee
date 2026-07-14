package drop_item

const (
	Id            = iota // 掉落ID
	Name                 // 掉落名称
	DropGroup            // 掉落组ID
	Item                 // 掉落道具
	Weight               // 掉落权重
	WeightInc            // 权重递增
	Deduplication        // 是否去重
	CheckExist           // 检查已有
	ExcludeExist         // 排除已有
	MustHave             // 必须拥有才会加入掉落
	ReplaceGroup         // 替代掉落组ID
	ValidDate            // 加入掉落时间
	ExpireDate           // 移出掉落时间
)
