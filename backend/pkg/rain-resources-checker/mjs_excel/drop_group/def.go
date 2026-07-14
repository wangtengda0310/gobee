package drop_group

const (
	Id            = iota // 掉落组id
	Name                 // 掉落组名称
	Weight               // 掉落组权重
	WeightInc            // 权重递增
	Deduplication        // 是否去重
	ValidDate            // 加入掉落时间
	ExpireDate           // 移出掉落时间
)
