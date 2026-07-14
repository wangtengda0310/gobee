/**
 * 技能配置数据
 *
 * @module config/Skill
 * @description
 * 从后端服务获取技能配置数据，提供技能映射和列表
 */
import { shallowRef } from 'vue'
import {GameExcelService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game";
import {SkillsTemplate} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel";

export const excelSkillMap = shallowRef<{ [p: `${number}`]: SkillsTemplate }>({})
export const excelSkillList = shallowRef<[number, string][]>([])

const initSkillConfig = async () => {
    console.log('[Skill.ts] initSkillConfig 开始调用')
    try {
        const res = await GameExcelService.GetAllSkillCfg()
        console.log('[Skill.ts] GetAllSkillCfg 返回, 键数量=', Object.keys(res).length)
        const ids: { [p: `${number}`]: SkillsTemplate } = {}
        for (let skillId in res) {
            if (res[skillId] != null) {
                ids[skillId] = res[skillId]
            }
        }
        console.log('[Skill.ts] 过滤后技能数量=', Object.keys(ids).length)
        excelSkillMap.value = ids
        excelSkillList.value = Object.keys(ids).map(k => [Number(k), ids[k]?.SkillName])
        console.log('[Skill.ts] excelSkillList.value 长度=', excelSkillList.value.length, '第一条=', excelSkillList.value[0])
    } catch (err) {
        console.log("[Skill.ts] 获取Skill失败", err)
    }
}

// 模块加载时自动初始化一次
await initSkillConfig()

// 导出刷新函数供 loadCases 调用
export const refreshSkillConfig = initSkillConfig
