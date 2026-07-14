package shop

const (
	ShopType        = iota // 商店类型(EShopType枚举)
	Name                   // 商店名称
	ShopName               // 商店显示名称
	UseCurrency            // 商店使用的货币
	OpenTime               // 开启时间
	CloseTime              // 关闭时间
	IsLimitedShop          // 是否为限时商店
	IsDynamicsShop         // 是否为动态商店
	InMainShopOrder        // 主界面中商店顺序
)
