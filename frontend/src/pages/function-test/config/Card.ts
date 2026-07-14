/**
 * 卡牌配置数据
 *
 * @module config/Card
 * @description
 * 从后端服务获取卡牌配置数据，提供卡牌映射和列表
 */
import { shallowRef } from 'vue'
import {CardsTemplate} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel";
import {GameExcelService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game";

export const excelCardsMap = shallowRef<{ [p: `${number}`]: CardsTemplate }>({})
export const excelCardsList = shallowRef<[number, string, number, number][]>([])

const initCardConfig = async () => {
    console.log('[Card.ts] initCardConfig 开始调用')
    try {
        const res = await GameExcelService.GetAllCardCfg()
        console.log('[Card.ts] GetAllCardCfg 返回, 键数量=', Object.keys(res).length)
        const ids: { [p: `${number}`]: CardsTemplate } = {}
        for (let cardId in res) {
            if (res[cardId] != null) {
                ids[cardId] = res[cardId]
            }
        }
        console.log('[Card.ts] 过滤后卡牌数量=', Object.keys(ids).length)
        excelCardsMap.value = ids
        excelCardsList.value = Object.keys(ids).map(k => {
            return [Number(k), ids[k]?.Name, ids[k]?.Point, ids[k]?.AttrType]
        })
        console.log('[Card.ts] excelCardsList.value 长度=', excelCardsList.value.length, '第一条=', excelCardsList.value[0])
    } catch (err) {
        console.log("[Card.ts] 获取Card失败", err)
    }
}

// 模块加载时自动初始化一次
await initCardConfig()

// 导出刷新函数供 loadCases 调用
export const refreshCardConfig = initCardConfig

export const cardTypeMap = {
    0: "空",
    1: "金",
    2: "木",
    3: "水",
    4: "火",
    5: "土",
    6: "大",
}
