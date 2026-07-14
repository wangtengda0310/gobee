package hero

const (
	Id                  = iota // 武将ID
	Name                       // 武将名称
	EHeroId                    // E#HeroId
	IsOpen                     // 武将是否开放
	OpenDate                   // 开放时间
	Gender                     // 性别
	Point                      // 初始体力
	HpLimit                    // 体力上限
	HandLimit                  // 手牌上限
	EquipLimit                 // 装备上限
	Country                    // 国号
	IsAlwaysZhuGong            // 是否常备主公
	Skill                      // 技能 ID
	ExcludeIdentity            // 排除的身份枚举
	NotUseModeType             // 不可以使用的房间模式
	HeroType                   // 武将类型
	EHeroType                  // E#HeroType
	CanMelt                    // 是否能熔炼
	MeltName                   // 熔炼名称
	IsNewHero                  // 是否为新增武将
	IsGacha                    // 是否招募产出
	BelongExpansionPack        // 所属扩展包
)
