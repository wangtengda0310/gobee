package shop_goods

const (
	Id                     = iota // 商品id
	Name                          // 商品名称
	Desc                          // 商品说明
	Item                          // 商品道具
	ExtraItem                     // 额外道具
	CostId                        // 消耗货币ID(=0代表RMB-元)
	ShopType                      // 商店类型(EShopType)
	RechargeAndriod               // 安卓充值ID
	RechargeIOS                   // IOS充值ID
	RechargeMulti                 // 充值倍数
	RechargeMultiGroup            // 充值倍数组
	Price                         // 价格
	OldPrice                      // 原价
	Discount                      // 折扣
	LimitType                     // 限购类型
	LimitCount                    // 限购数量
	MaxBuyCount                   // 最大购买数量
	Pos                           // 商品顺序
	IsNew                         // 是否新上架
	Countdown                     // 倒计时(秒)
	IsHomePage                    // 是否首页推荐
	Icon                          // 商品图标
	OnShelfTime                   // 上架时间
	OffShelfTime                  // 下架时间
	RechargeMultiBeginTime        // 充值倍数开始时间
	RechargeMultiEndTime          // 充值倍数结束时间
	PreSelOutShopId               // 预选售罄商店ID
	HideWhenSelOut                // 售罄时是否隐藏
	LimitZeroWhenHasItem          // 拥有时限购为0
	IconDisplayType               // 商品icon显示类型
	ShowItemCount                 // 是否显示商品数量
	ShowQuality                   // 是否显示品质
	IconInBuyWindow               // 商品在购买窗口图标
	RechargeMultiGoodsDes         // 充值倍数商品说明
	RewardID                      // 奖励ID
)
