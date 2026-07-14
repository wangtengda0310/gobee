package diff

import (
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/activitywiki_def"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/accumulated_recharge"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/achieve"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/achieve_hero"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/activity"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/arena_score_rewards"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/buff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/country"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/draw_pet"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/draw_skin"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_group"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_item"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_rule"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_lines"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_collition"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_item"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_spine"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_ui"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/item_hero_skin"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/limit_skin_times_reward"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/pet"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/pet_audio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/recommend_bd"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/robot_action"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/season_pass"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/season_pass_bag"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/season_pass_reward"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/season_pass_task"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/shop"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/shop_goods"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_lines"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_melt"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_tag"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_ui"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/task_complete_cond"
)

// Diffable 可对比的接口
type Diffable interface {
	GetType() string // 获取数据类型（用于区分不同的数组）
}

// DiffResult 总体对比结果
type DiffResult struct {
	Timestamp     time.Time
	DataTypeStats map[string]*TypeDiffResult // 各类型数据的对比结果
}

// TypeDiffResult 单种数据类型的对比结果
type TypeDiffResult struct {
	DataType      string
	AddedIds      []int
	RemovedIds    []int
	ChangedItems  map[int]*FieldChanges
	ItemCount     int // 当前总数
	PreviousCount int // 上次总数
}

// FieldChanges 字段变化
type FieldChanges struct {
	Id            int
	Name          string
	FieldPath     string // 字段完整路径
	Changes       []*FieldChange
	NestedStructs map[string]*NestedStructChange // 嵌套项摘要
}

// FieldChange 单个字段变化
type FieldChange struct {
	FieldPath   string // 完整字段路径，如 "MainSkill.SkillName"
	FieldName   string
	StructName  string // 所属结构体名
	NestedLevel int    // 嵌套层级
	OldValue    interface{}
	NewValue    interface{}
	ValueType   string // 值类型
	ChangeType  ChangeType
}

// NestedStructChange 嵌套结构体变化
type NestedStructChange struct {
	StructPath string                         // 结构体路径
	StructType string                         // 结构体类型名
	FieldCount int                            // 变化的字段数
	Changes    []*FieldChange                 // 该结构体的字段变化
	Children   map[string]*NestedStructChange // 子结构体变化
}

// ChangeType 变化类型
type ChangeType string

const (
	ChangeTypeModified ChangeType = "modified"
	ChangeTypeAdded    ChangeType = "added"
	ChangeTypeRemoved  ChangeType = "removed"
)

// DataContainer 数据容器，包含所有要对比的数据类型
type DataContainer struct {
	HeroWikiDiff           *herowiki_def.HeroWikiDiff
	HeroWikiDiffResult     *HeroWikiDiffResult // 新增：HeroWikiDiff的diff结果
	HeroDiff               *[]hero.HeroDiff
	CountryDiff            *[]country.CountryDiff
	HeroLinesDiff          *[]hero_lines.HeroLinesDiff
	HeroUIDiff             *[]hero_ui.HeroUIDiff
	HeroSkinCollectionDiff *[]hero_skin_collition.HeroSkinCollectionDiff
	HeroSkinItemDiff       *[]hero_skin_item.HeroSkinItemDiff
	HeroSkinSpineDiff      *[]hero_skin_spine.HeroSkinSpineDiff
	ItemHeroSkinDiff       *[]item_hero_skin.HeroSkinDiff
	DropGroupDiff          *[]drop_group.DropGroupDiff
	DropItemDiff           *[]drop_item.DropItemDiff
	DropRuleDiff           *[]drop_rule.DropRuleDiff
	SeasonPassDiff         *[]season_pass.SeasonPassDiff
	SeasonPassBagDiff      *[]season_pass_bag.SeasonPassBagDiff
	SeasonPassRewardDiff   *[]season_pass_reward.SeasonPassRewardDiff
	SeasonPassTaskDiff     *[]season_pass_task.SeasonPassTaskDiff
	ArenaScoreRewardsDiff  *[]arena_score_rewards.ArenaScoreRewardDiff
	SkillDiff              *[]skill.SkillDiff
	SkillLineDiff          *[]skill_lines.SkillLinesDiff
	SkillMeltDiff          *[]skill_melt.SkillMeltDiff
	SkillTagDiff           *[]skill_tag.SkillTagDiff
	SkillUIDiff            *[]skill_ui.SkillUIDiff
	BuffDiff               *[]buff.BuffDiff
	RecommendBdDiff        *[]recommend_bd.RecommendBdDiff
	RobotActionDiff        *[]robot_action.RobotActionDiff
	TaskCompleteCondDiff   *[]task_complete_cond.TaskCompleteConditonDiff
	AchieveHeroDiff        *[]achieve_hero.HeroAchieveDiff
	AchieveDiff            *[]achieve.AchieveDiff
	// 活动Wiki数据
	ActivityWikiDiff         *activitywiki_def.ActivityWikiDiff
	ActivityDiff             *[]activity.ActivityDiff
	DrawSkinDiff             *[]draw_skin.DrawSkinDiff
	LimitSkinTimesRewardDiff *[]limit_skin_times_reward.LimitSkinTimesRewardDiff
	ShopDiff                 *[]shop.ShopDiff
	ShopGoodsDiff            *[]shop_goods.ShopGoodsDiff
	// 结缘庭相关数据
	DrawPetDiff  *[]draw_pet.DrawPetDiff
	PetDiff      *[]pet.PetDiff
	PetAudioDiff *[]pet_audio.PetAudioDiff
	// activity-wiki-dev: 新增字段 - 累充活动奖励数据
	AccumulatedRechargeDiff *[]accumulated_recharge.AccumulatedRechargeDiff
}

// HeroWikiDiffResult HeroWikiDiff的diff结果
type HeroWikiDiffResult struct {
	Timestamp         time.Time
	HeroesDiff        map[string]*HeroDiffDetail // 按EHeroId索引的diff详情
	Summary           *HeroWikiDiffSummary
	RemovedHeroesData map[string]*herowiki_def.HeroCompleteData // 删除武将的完整数据（用于前端显示）
}

// HeroDiffDetail 单个武将的diff详情
type HeroDiffDetail struct {
	EHeroId       string
	Name          string
	FieldChanges  []*FieldChange
	NestedChanges map[string]*NestedStructChange
	ChangeType    ChangeType // added, removed, modified
	ChangeCount   int        // 变化字段总数
}

// HeroWikiDiffSummary 汇总信息
type HeroWikiDiffSummary struct {
	TotalHeroes    int
	AddedHeroes    []string // 新增的武将ID
	RemovedHeroes  []string // 删除的武将ID
	ModifiedHeroes []string // 修改的武将ID
	TotalChanges   int      // 总变化数
}
