/**
 * Excel 打开操作 composable
 *
 * 提供统一的 Excel 打开功能和错误处理
 */
import {OpenExcel} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/activity-wiki-check/activitywikicheckservice.js"
import {type MessageApi} from "naive-ui"

/**
 * 统一的 Excel 错误处理
 * 区分"成功但伴警告"和真正的失败
 */
function handleExcelError(message: MessageApi, err: unknown, context: string) {
  if (err && typeof err === 'object' && 'message' in err) {
    const errorMsg = (err as { message: unknown }).message as string
    if (errorMsg.includes('undefined') || errorMsg.includes('Cannot read properties')) {
      message.warning(`${context}成功，但有警告: ${String(err)}`)
    } else {
      message.error(`${context}失败: ${errorMsg}`)
    }
  } else {
    message.error(`${context}失败: ${String(err)}`)
  }
}

/**
 * 通过 Sheet 名称打开 Excel
 *
 * @param message - Naive UI 的 message API 实例，由调用方在 setup() 中通过 useMessage() 获取后传入
 * @param sheetName - Sheet 名称（如"活动表|Activity"）
 * @param excelDir - Excel 配置目录路径
 */
export async function openExcelBySheet(message: MessageApi, sheetName: string, excelDir: string) {
  try {
    await OpenExcel(sheetName, excelDir)
    message.success(`正在打开: ${sheetName}`)
  } catch (err) {
    handleExcelError(message, err, `打开Excel(${sheetName})`)
  }
}

/**
 * 直接打开 Excel 文件路径
 *
 * @param message - Naive UI 的 message API 实例，由调用方在 setup() 中通过 useMessage() 获取后传入
 * @param filePath - Excel 文件完整路径
 */
export async function openExcelFile(message: MessageApi, filePath: string) {
  try {
    await OpenExcel('', filePath)
    message.success('正在打开Excel文件')
  } catch (err) {
    handleExcelError(message, err, '打开Excel文件')
  }
}
