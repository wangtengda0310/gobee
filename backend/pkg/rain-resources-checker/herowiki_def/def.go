package herowiki_def

// HeroWikiDiff 整合所有武将相关数据
type HeroWikiDiff struct {
	// 索引映射，用于快速查找
	Indexes *HeroIndexes

	// 按武将ID组织的完整数据
	Heroes map[string]*HeroCompleteData
}

func (t HeroWikiDiff) GetType() string {
	return "HeroWikiDiffDiff"
}

// HeroIndexes 索引映射
type HeroIndexes struct {
	// ID索引
	HeroByID      map[int]string // HeroDiff.Id -> EHeroId
	HeroByEHeroId map[string]int // EHeroId -> HeroDiff.Id

	// 皮肤索引
	SkinsByHeroId map[string][]int // HeroId -> []SkinItemId
	LinesBySkinId map[int][]int    // SkinItemId -> []LineId

	// 技能索引
	ESkillBySkillId map[int]string
	SkillsByHeroId  map[string][]string // HeroId -> []SkillId
	LinesBySkillId  map[string][]int    // SkillId -> []LineId

	// buff索引
	BuffIdByEBuffId map[string]string

	// 国家索引
	HeroesByCountry map[string][]string // Country -> []HeroId

	// 成就索引
	AchieveHeroesByHeroId map[string][]string // HeroId -> []AchieveId
	AchieveByHeroId       map[string][]int    // HeroId -> 去掉开头10为HeroItemId -> []AchieveId
}

func (t HeroIndexes) GetType() string {
	return "HeroIndexesDiff"
}

// HeroCompleteData 单个武将的完整数据
type HeroCompleteData struct {
	// 基础信息
	Basic *HeroBasicInfo

	// UI相关信息
	UI *HeroUIInfo

	// 皮肤集合
	Skins []*HeroSkinInfo

	// 技能集合
	Skills []*HeroSkillInfo

	// 所属国家
	Country *CountryInfo

	// 推荐布阵
	RecommendBd *RecommendBdInfo

	// 成就
	Achievements []*HeroAchievementInfo

	// 机器人行为
	RobotActions []*RobotActionInfo

	// 掉落相关信息
	DropInfo *HeroDropInfo
}

func (t HeroCompleteData) GetType() string {
	return "HeroCompleteDataDiff"
}

// HeroBasicInfo 武将基础信息
type HeroBasicInfo struct {
	Id                  int
	EHeroId             string
	Name                string
	IsOpen              bool
	OpenDate            string
	Gender              int
	Point               int
	HpLimit             int
	HandLimit           int
	EquipLimit          int
	Country             string
	IsAlwaysZhuGong     bool
	ExcludeIdentity     []int
	NotUseModeType      []int
	HeroType            int
	EHeroType           string
	CanMelt             bool
	MeltName            []string
	IsNewHero           bool
	IsGacha             bool
	BelongExpansionPack string
}

func (t HeroBasicInfo) GetType() string {
	return "HeroBasicInfoDiff"
}

// HeroUIInfo 武将UI信息
type HeroUIInfo struct {
	HeroDiffId          int
	Name                string
	VoiceType           int
	BelongExpansionPack string
	AwardDes            string
	LongIntroduction    string
	ShortIntroduction   string
	Evidence            string
	Evaluation          string
	IsNew               bool
	CopyWriter          string
	SkillDesigner       string
	GetWay              string
	Position            string
	ExclusiveCard       []int
	NewbieShowSkillTag  []int
	WinningRateIn2v2    int
	WinRateShowPriority int
}

func (t HeroUIInfo) GetType() string {
	return "HeroUIInfoDiff"
}

// HeroSkinInfo 武将皮肤信息
type HeroSkinInfo struct {
	// 皮肤基础
	ItemId     int
	HeroId     string
	Name       string
	GetWay     string
	SkinPinYin string
	SkinType   string
	RailType   string

	// 皮肤UI
	SeatSpecialImg      string
	HeroUIExtraIcons    []string
	OriginalArtDesigner string
	CollectionType      string
	IsOpen              bool
	OpenDate            string
	CollectionTagImg    string
	KillShowTime        int
	BodyOffset          []int
	HasOutFrameIcon     bool

	// 皮肤Spine
	Spine *HeroSkinSpineInfo

	// 皮肤资源
	Resource *HeroSkinResourceInfo

	// 皮肤台词
	LinesDubbed string
	Lines       []*HeroLineInfo

	// 收藏册
	Collection *HeroSkinCollectionInfo
}

func (t HeroSkinInfo) GetType() string {
	return "HeroSkinInfoDiff"
}

// HeroSkinSpineInfo 皮肤Spine信息
type HeroSkinSpineInfo struct {
	SkinItemId                          int
	IsHasSeatSpine                      bool
	IsHasBookSpine                      bool
	IsHasMainBgSpine                    bool
	MainBgFx                            string
	IsHasSeatKillSpine                  bool
	KillFxId                            int
	IsHasCollectionBgSpine              bool
	IsHasCollectionCardBgSpine          bool
	SpineAnimAudio                      string
	IsHasCollectionCardBgSpineDuplicate bool
	KillAudio                           string
}

func (t HeroSkinSpineInfo) GetType() string {
	return "HeroSkinSpineInfoDiff"
}

// HeroSkinResourceInfo 皮肤资源信息
type HeroSkinResourceInfo struct {
	SkinItemId int
	Path       string
}

func (t HeroSkinResourceInfo) GetType() string {
	return "HeroSkinResourceInfoDiff"
}

// HeroSkinCollectionInfo 皮肤收藏册信息
type HeroSkinCollectionInfo struct {
	Type     string
	Name     string
	NameImg  string
	NameBg   string
	Desc     string
	Weight   int
	OpenDate string
}

func (t HeroSkinCollectionInfo) GetType() string {
	return "HeroSkinCollectionInfoDiff"
}

// HeroLineInfo 台词信息
type HeroLineInfo struct {
	Id           int
	Type         string
	TabName      string
	Text         string
	AudioId      string
	Achievements []int
	GroupId      string
}

func (t HeroLineInfo) GetType() string {
	return "HeroLineInfoDiff"
}

// HeroSkillInfo 技能信息
type HeroSkillInfo struct {
	// 技能基础
	Basic *SkillBasicInfo

	// 技能UI
	UI *SkillUIInfo

	// 技能熔炼
	Melt *SkillMeltInfo

	// 技能台词
	Lines []*SkillLineInfo

	// 技能标签
	Tags []*SkillTagInfo
}

func (t HeroSkillInfo) GetType() string {
	return "HeroSkillInfoDiff"
}

// SkillBasicInfo 技能基础信息
type SkillBasicInfo struct {
	Id                string
	SkillName         string
	ShortSkillName    string
	ESkillId          string
	SkillType         string
	IsFromOther       bool
	CounterFormula    string
	IsFromAura        bool
	TransCard         string
	SkillFromType     string
	ResetCounterType  string
	ResetTimesType    string
	SkillLimitTimes   int
	TotalLimitTimes   int
	TriggerCondition  []int
	SkillEffect       []string
	EmptyWaitTime     string
	Buff              []*BuffInfo
	ShowPointAndAttr  bool
	MutexSkill        []string
	BattleCardClass   string
	AIJudgeArea       []string
	DeadType          int
	InitPro           []string
	MagicSkillID      string
	IsAutoSelAllOther bool
	IsForbidCopy      bool
	IsForbidTrans     bool
	IsForbidDestroy   bool
}

func (t SkillBasicInfo) GetType() string {
	return "SkillBasicInfoDiff"
}

// BuffInfo 技能基础信息
type BuffInfo struct {
	Id                     string
	EBuffId                string
	Name                   string
	NeedRecord             bool
	BuffType               int
	IsDeleteByCasterDead   bool
	IsDeleteByExecutorDead bool
	IsValidByFengJin       bool
	IsReserveByRemoveSkill bool
	TransferSkillBuffType  int
	Round                  int
	Value                  int
	EndByEffect            bool
	EndType                string
	IsServerOnly           bool
	IsCasterOnly           bool
	OwnerType              int
	ShowArea               string
	Icon                   string
	Effect                 int
	FlashEffect            int
	BuffPriority           int
	OverlyingType          int
	IsTrigger              bool
	BuffState              []int
	BuffPro                []int
	ProValue               []int
	CostEffectValue        []int
	TriggerTiming          []int
	TriggerPriority        []int
	TriggerCondition       []string
	TriggerAction          []int
	BuffDot                int
	EffectDescribe         string
}

func (t BuffInfo) GetType() string {
	return "BuffInfoDiff"
}

// SkillUIInfo 技能UI信息
type SkillUIInfo struct {
	Id              string
	SkillName       string
	PlayCardAudio   string
	IdentityLine    []string
	Audio           []int
	SkillText       string
	ShortSkillText  string
	KeyWords        []int
	SettlementDes   string
	Allusion        string
	DesignThought   string
	SkillTag        []int
	BattleSkillStep []string
	HasRelation     bool
	RelatedSkill    []string
	SpecialAudio    string
	ESkillId        string
	EAudio          string
	EAudioId        string
	ESkillTag       string
}

func (t SkillUIInfo) GetType() string {
	return "SkillUIInfoDiff"
}

// SkillMeltInfo 技能熔炼信息
type SkillMeltInfo struct {
	Id        string
	MeltPower int
	CanMelt   bool
}

func (t SkillMeltInfo) GetType() string {
	return "SkillMeltInfoDiff"
}

// SkillLineInfo 技能台词信息
type SkillLineInfo struct {
	Id              int
	SkillId         string
	SkinId          int
	SkillFirstLine  []int
	SkillSecondLine []int
	SkillThirdLine  []int
	SkillForthLine  []int
	SpecialAudio    string
}

func (t SkillLineInfo) GetType() string {
	return "SkillLineInfoDiff"
}

// SkillTagInfo 技能标签信息
type SkillTagInfo struct {
	SkillTag string
	TagName  string
	TagColor string
	TagDes   string
}

func (t SkillTagInfo) GetType() string {
	return "SkillTagInfoDiff"
}

// CountryInfo 国家信息
type CountryInfo struct {
	ECountry                  string
	Name                      string
	KingName                  string
	SeatFlagImagePath         string
	PlayerInfoFlagImagePath   string
	PlayerInfoFlagTextPath    string
	IsWhiteBg                 bool
	IsOpen                    bool
	NameArtSelectedFontPath   string
	NameArtUnselectedFontPath string
	ShiLiSmallTopIcon         string
	ShiLiSmallBottomIcon      string
	SortPriority              int
	GuildIcon                 string

	// 该国家的武将
	Heroes []string
}

func (t CountryInfo) GetType() string {
	return "CountryInfoDiff"
}

// RecommendBdInfo 推荐布阵信息
type RecommendBdInfo struct {
	HeroId string
	Level1 []string
	Level2 []string
	Level3 []string
	Level4 []string
	Level5 []string
	Level6 []string
}

func (t RecommendBdInfo) GetType() string {
	return "RecommendBdInfoDiff"
}

// HeroAchievementInfo 武将成就信息
type HeroAchievementInfo struct {
	HeroAchieve   *HeroAchieveInfo
	AchieveDetail *AchieveDetailInfo
}

func (t HeroAchievementInfo) GetType() string {
	return "HeroAchievementInfoDiff"
}

type HeroAchieveInfo struct {
	Id           string   // 局内成就类型 (E#HeroAchieve)
	Name         string   // 成就名称
	IsMult       bool     // 是否复用
	Mode         []int    // 房间模式
	MinPlayerNum int      // 房间人数
	UseHero      []string // 使用英雄 (E#HeroId列表)
	Class        int      // 身份类型
	Camp         int      // 阵营
	Identity     int      // 身份
	Hooker       []int    // 对应钩子
	HookerTarget []int    // 对应钩子
	CondParam    []int    // 条件一
}

func (t HeroAchieveInfo) GetType() string {
	return "HeroAchieveInfoDiff"
}

type AchieveDetailInfo struct {
	Id             int                        // 成就id
	Name           string                     // 成就名称
	IsHide         bool                       // 是否隐藏
	CompleteCondId int                        // 成就完成条件
	Reward         []string                   // 成就奖励
	None           string                     // 空列
	Des            string                     // 成就描述
	Condition      string                     // 成就完成条件
	ConditionInfo  *TaskCompleteConditionInfo // 成就完成条件
	HeroItemId     []int                      // 关联武将ID
	OpenDate       string                     // 开放时间
}

func (t AchieveDetailInfo) GetType() string {
	return "AchieveDetailInfoDiff"
}

type TaskCompleteConditionInfo struct {
	Id                int
	CondDes           string
	CompleteCond      string
	CompleteCondParam []int
	JumpCond          string
	JumpParm          []int
}

func (t TaskCompleteConditionInfo) GetType() string {
	return "TaskCompleteConditionInfoDiff"
}

// RobotActionInfo 机器人行为信息
type RobotActionInfo struct {
	Id               string
	Action           []string
	TargetNum        []int
	TargetType       []string
	CardNum          []int
	CardFromType     []string
	TransCardSkill   []int
	DefaultCardSkill []int
}

func (t RobotActionInfo) GetType() string {
	return "RobotActionInfoDiff"
}

// HeroDropInfo 武将掉落相关信息
type HeroDropInfo struct {
	// 武将直接掉落的规则
	DirectDropRules []*DropRuleSummary

	// 武将作为保底/特殊掉落的规则
	GuaranteeDropRules []*DropRuleSummary

	// 武将所在的所有掉落组
	DropGroups []*DropGroupSummary

	// 按掉落类型分类
	ByDropType map[string]*DropTypeInfo
}

func (t HeroDropInfo) GetType() string {
	return "HeroDropInfoDiff"
}

// DropRuleSummary 掉落规则摘要
type DropRuleSummary struct {
	RuleId      int
	RuleName    string
	DropCount   int
	DropGroups  []int // 关联的掉落组ID列表
	IsGuarantee bool  // 是否是保底掉落
}

func (t DropRuleSummary) GetType() string {
	return "DropRuleSummaryDiff"
}

// DropGroupSummary 掉落组摘要
type DropGroupSummary struct {
	GroupId     int
	GroupName   string
	Weight      int
	WeightInc   int
	Deduplicate bool
	ValidDate   string
	ExpireDate  string

	// 该掉落组中涉及此武将的具体掉落项
	DropItems []*HeroDropItem
}

func (t DropGroupSummary) GetType() string {
	return "DropGroupSummaryDiff"
}

// HeroDropItem 武将掉落项
type HeroDropItem struct {
	ItemId      int
	ItemName    string
	DropGroupId int

	// 掉落物品信息
	ItemConfigs []*ItemConfig

	Weight       int
	WeightInc    int
	Deduplicate  bool
	CheckExist   bool
	ExcludeExist bool
	MustHave     bool
	ReplaceGroup int
	ValidDate    string
	ExpireDate   string
}

func (t HeroDropItem) GetType() string {
	return "HeroDropItemDiff"
}

// ItemConfig 物品配置
type ItemConfig struct {
	ItemId int    // 物品ID
	Count  int    // 数量
	IsHero bool   // 是否是武将
	HeroId string // 如果是武将，对应的EHeroId
}

func (t ItemConfig) GetType() string {
	return "ItemConfigDiff"
}

// DropTypeInfo 按类型分类的掉落信息
type DropTypeInfo struct {
	TypeName   string
	DropRules  []*DropRuleSummary
	TotalCount int
}

func (t DropTypeInfo) GetType() string {
	return "DropTypeInfoDiff"
}
