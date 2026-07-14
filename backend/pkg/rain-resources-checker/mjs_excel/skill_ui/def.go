package skill_ui

const (
	Id              = iota // 技能ID
	SkillName              // 技能名称
	PlayCardAudio          // 出牌音效
	IdentityLine           // 身份技能台词
	None                   // 空
	Audio                  // 打牌时触发技能的台词
	SkillText              // 技能文案
	ShortSkillText         // 技能文案（短）
	KeyWords               // 技能关键字
	SettlementDes          // 结算详情
	Allusion               // 技能典故
	DesignThought          // 设计思路
	SkillTag               // 技能标签
	BattleSkillStep        // 战斗内阶段显示
	HasRelation            // 是否有关联技能
	RelatedSkill           // 关联技能技能ID
	SpecialAudio           // 技能特殊音效
	ESkillId               // E#SkillId
	EAudio                 // E#AudioId
	EAudioId               // E#AudioId[] (PlayCardAudio)
	ESkillTag              // E#SkillTag[]
)
