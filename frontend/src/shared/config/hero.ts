/**
 * 英雄配置数据
 *
 * @module config/Hero
 * @description
 * 从后端服务获取英雄配置数据，提供英雄映射和列表
 */
import { shallowRef } from 'vue'
import {GameExcelService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game";
import {HeroConfig} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel_config";

export const excelHeroMap = shallowRef<{ [p: `${number}`]: HeroConfig }>({})
export const excelHeroList = shallowRef<[number, string][]>([])

const initHeroConfig = async () => {
    console.log('[Hero.ts] initHeroConfig 开始调用')
    try {
        const res = await GameExcelService.GetAllHeroCfg()
        console.log('[Hero.ts] GetAllHeroCfg 返回, 键数量=', Object.keys(res).length)
        const ids: { [p: `${number}`]: HeroConfig } = {}
        for (let heroId in res) {
            if (res[heroId] == null || res[heroId].HeroType != 1) {
                continue
            }
            ids[heroId] = res[heroId]
        }
        console.log('[Hero.ts] 过滤后英雄数量=', Object.keys(ids).length)
        excelHeroMap.value = ids
        excelHeroList.value = Object.keys(ids).map(k => [Number(k), ids[k]?.Name])
        console.log('[Hero.ts] excelHeroList.value 长度=', excelHeroList.value.length, '第一条=', excelHeroList.value[0])
    } catch (err) {
        console.log("[Hero.ts] 获取Hero失败", err)
    }
}

// 模块加载时自动初始化一次
await initHeroConfig()

// 导出刷新函数供 loadCases 调用
export const refreshHeroConfig = initHeroConfig
