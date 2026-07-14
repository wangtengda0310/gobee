/**
 * 技能描述文案数据（来自「技能表现配置表|SkillUI」）
 *
 * @module config/SkillUI
 * @description
 * SkillsTemplate（GetAllSkillCfg）只有技能名 SkillName，没有描述字段。
 * 本模块从后端 SkillUIDescService 加载 skillId -> SkillText 映射，
 * 供座位卡片「删除技能」下拉框的已选标签做悬浮提示。
 *
 * 路径由后端统一策划配表目录配置决定（无需前端传参），
 * 文案缺失时退化为不显示提示，不阻塞用例加载。
 */
import { shallowRef } from 'vue'
import {SkillUIDescService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game";

export const excelSkillDescMap = shallowRef<{ [skillId: number]: string | undefined }>({})

/**
 * 刷新技能描述映射（路径由后端统一配置决定）
 */
export const refreshSkillDescConfig = async () => {
    try {
        const res = await SkillUIDescService.LoadSkillUIDesc()
        excelSkillDescMap.value = res ?? {}
    } catch (err) {
        console.warn('[SkillUI.ts] 加载技能描述失败, tooltip 将不显示', err)
        excelSkillDescMap.value = {}
    }
}
