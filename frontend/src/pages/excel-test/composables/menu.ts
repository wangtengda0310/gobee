/**
 * Excel 测试页面菜单配置
 *
 * 定义顶部菜单栏选项
 */
import {DialogApiInjection} from "naive-ui/es/dialog/src/DialogProvider";
import {MessageApiInjection} from "naive-ui/es/message/src/MessageProvider";
import {MenuOption, NButton} from "naive-ui";
import {h, ref} from "vue";
import {loadExcelsAndExcelRules, saveExcelRules, startExcelRules} from "./func";

/**
 * 创建菜单选项
 * @param dialog 对话框 API
 * @param message 消息 API
 */
export const menuOptions = (dialog: DialogApiInjection, message: MessageApiInjection): MenuOption[] => {
    return [
        {
            label: () => h(NButton, {
                text: true,
                onClick: (e: MouseEvent) => {
                    // e.stopPropagation() // 阻止菜单关闭
                    // loadCases()
                    dialog.warning({
                        title: '加载配表',
                        content: '加载配表会覆盖当前内存中已加载的配表, 如有需要请先保存',
                        positiveText: '确定',
                        negativeText: '取消',
                        draggable: true,
                        onPositiveClick: () => {
                            loadExcelsAndExcelRules()
                        },
                        onNegativeClick: () => {
                        }
                    })
                }
            }, {
                default: () => '加载配表'
            }),
            key: 'load-excels',
        },
        {
            label: () => h(NButton, {
                text: true,
                onClick: (e: MouseEvent) => {
                    // e.stopPropagation() // 阻止菜单关闭
                    // openSaveCaseModal()
                    dialog.warning({
                        title: '保存配表',
                        content: '会保存当前所有配表到工作目录',
                        positiveText: '确定',
                        negativeText: '取消',
                        draggable: true,
                        onPositiveClick: () => {
                            saveExcelRules()
                            // console.error("还没实现")
                        },
                        onNegativeClick: () => {
                        }
                    })
                }
            }, {
                default: () => '保存用例'
            }),
            key: 'save-excels',
        },
        {
            label: () => h(NButton, {
                text: true,
                onClick: (e: MouseEvent) => {
                    // e.stopPropagation() // 阻止菜单关闭
                    // openSaveCaseModal()
                    dialog.warning({
                        title: '执行检查',
                        content: '执行所有规则到表',
                        positiveText: '确定',
                        negativeText: '取消',
                        draggable: true,
                        onPositiveClick: () => {
                            startExcelRules()
                            // console.error("还没实现")
                        },
                        onNegativeClick: () => {
                        }
                    })
                }
            }, {
                default: () => '执行检查'
            }),
            key: 'check-excels',
        },
    ]
}

/** 当前激活的菜单项 */
export const activeKey = ref<string | null>(null)
