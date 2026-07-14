/**
 * 错误码配置数据
 *
 * @module config/ErrorCode
 * @description
 * 从后端服务获取错误码映射
 */
import {GameExcelService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game";

let ids: { [p: `${number}`]: string } = {}

await GameExcelService.GetErrorCodeMap().then(res => {
    ids = res as { [p: `${number}`]: string }
}).catch(err => {
    console.log("获取proto失败")
})

export const errCodeMap = ids
