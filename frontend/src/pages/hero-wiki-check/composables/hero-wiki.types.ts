/**
 * HeroWikiCheck 页面的类型定义
 *
 * 包含武将 Wiki 检查页面所需的 diff 数据映射类型
 */
import {SkillDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill";
import {SkillLinesDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_lines";
import {HeroLinesDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_lines";
import {CountryDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/country";
import {HeroSkinItemDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_item";

/**
 * Diff 数据索引映射
 *
 * 用于在 HeroWikiCheck 页面中快速查找各类 diff 数据
 */
export type DiffIndexMap = {
    skillsDiffMap: Map<number, SkillDiff>
    skillLinesDiffMap: Map<number, SkillLinesDiff>
    skillLinesDiffMap_Enum: Map<string, SkillLinesDiff>
    heroLinesDiffMap: Map<number, HeroLinesDiff>
    countryDiffMap: Map<string, CountryDiff>
    heroSkinItemDiffMap: Map<number, HeroSkinItemDiff[]>
}
