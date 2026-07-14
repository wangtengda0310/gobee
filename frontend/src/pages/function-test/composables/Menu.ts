import {MenuOption, NButton} from "naive-ui";
import {h, ref} from "vue";
import {JsonCaseService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";
import {showFuncCaseOptionModal} from "./Modals";
import {DialogApiInjection} from "naive-ui/es/dialog/src/DialogProvider";
import {MessageApiInjection} from "naive-ui/es/message/src/MessageProvider";
import {execNowCase, loadCases, saveModifiedCases} from "./Func";


export const menuOptions = (dialog: DialogApiInjection, message: MessageApiInjection): MenuOption[] => {
    return [
        {
            label: () => h(NButton, {
                text: true,
                onClick: (e: MouseEvent) => {
                    // e.stopPropagation() // 阻止菜单关闭
                    // loadCases()
                    dialog.warning({
                        title: '加载用例',
                        content: '加载用例会覆盖当前内存中已加载的用例, 如有需要请先保存',
                        positiveText: '确定',
                        negativeText: '取消',
                        draggable: true,
                        onPositiveClick: async () => {
                            await loadCases()
                        },
                        onNegativeClick: () => {
                        }
                    })
                }
            }, {
                default: () => '加载用例'
            }),
            key: 'load-cases',
        },
        {
            label: () => h(NButton, {
                text: true,
                onClick: (e: MouseEvent) => {
                    // e.stopPropagation() // 阻止菜单关闭
                    // openSaveCaseModal()
                    dialog.warning({
                        title: '保存用例',
                        content: '会保存当前所有用例到工作目录',
                        positiveText: '确定',
                        negativeText: '取消',
                        draggable: true,
                        onPositiveClick: () => {
                            saveModifiedCases(message)
                        },
                        onNegativeClick: () => {
                        }
                    })
                }
            }, {
                default: () => '保存用例'
            }),
            key: 'save-cases',
        },
        {
            label: () => h(NButton, {
                text: true,
                onClick: (e: MouseEvent) => {
                    // e.stopPropagation() // 阻止菜单关闭
                    execNowCase(message)
                }
            }, {
                default: () => '执行用例'
            }),
            key: 'exec-cases',
        },
        {
            label: () => h(NButton, {
                text: true,
                onClick: (e: MouseEvent) => {
                    // e.stopPropagation() // 阻止菜单关闭
                    JsonCaseService.IsRunningRobotTest().then(res => {
                        if (res) {
                            JsonCaseService.StopRobotTest().then(res => {
                                message.info("机器人停止成功")
                            }).catch(err => {
                                message.error(err)
                            })
                        } else {
                            message.error("用例未运行")
                        }
                    }).catch(err => {
                        message.error(err)
                    })
                }
            }, {
                default: () => '停止用例'
            }),
            key: 'stop-cases',
        },
        {
            label: () => h(NButton, {
                text: true,
                onClick: (e: MouseEvent) => {
                    // e.stopPropagation() // 阻止菜单关闭
                    showFuncCaseOptionModal.value = true
                }
            }, {
                default: () => '设置'
            }),
            key: 'option',
        },
        {
            label: '其他选项',
            key: 'other-option',
            children: [
                {
                    label: '饮品',
                    key: 'beverage',
                    children: [
                        {
                            label: '威士忌',
                            key: 'whisky'
                        }
                    ]
                },
                {
                    label: '食物',
                    key: 'food',
                    children: [
                        {
                            label: '三明治',
                            key: 'sandwich'
                        }
                    ]
                },
            ]
        }
    ]
}

export const activeKey = ref<string | null>(null)