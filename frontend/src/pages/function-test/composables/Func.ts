import {dataRef, ExtraCaseTreeOption} from "./use-case-data";
import {TreeOption} from "naive-ui";
import {FuncCaseConfigService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";
import {
    Concurrency, DebugLevel, DebugLog,
    Desc,
    ExcelResourcesDir,
    JsonsDir,
    LoginTime,
    RobotPrefix,
    RoomOpTime,
    ServerAddr,
    ServerPort
} from "./Option";
// 飞书配置从全局设置导入
import { FeiShuGuid, FeiShuNtf } from "../../settings/composables/use-settings";
import {JsonCaseService, QAFuncCase} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";
import {GameExcelService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game";
import {refreshHeroConfig, excelHeroList} from "@shared/config/hero";
import {refreshCardConfig, excelCardsList} from "../config/Card";
import {refreshSkillConfig, excelSkillList} from "../config/Skill";
import {initLogStatistic, resetLogCache} from "./RobotTestLog";
import {MessageApiInjection} from "naive-ui/es/message/src/MessageProvider";
import {nowCaseData} from "./use-case-data";
import {ref} from "vue";

/** 当前激活的页签名称，执行用例时自动切换到"执行日志" */
export const activeTab = ref('option')

// 把Case从内存序列化的初步结构体处理方法
const preSerializeJsonCaseSolve = (c: TreeOption & ExtraCaseTreeOption) => {
    const clone = JSON.parse(JSON.stringify(c));

    return new QAFuncCase({
        case: clone.label,
        desc: clone.caseDesc,
        manager: clone.caseManager,
        initYanWu: {
            presentId: clone.initYanWu!.presentId,
            cardPile: clone.initYanWu!.cardPile,
            cards: clone.initYanWu!.cards,
            operateTimeMs: clone.initYanWu!.operateTimeMs,
            customHeroes: clone.initYanWu!.customHeroes.map(h => {
                const m = {...h}
                delete m['uuid']
                return m
            }),
        },
        steps: clone.caseSteps?.map((s, i) => {
            s.id = i + 1

            if (s.assets) s.assets = s.assets.map(a => {
                const m = {...a}
                delete m['uuid']
                return m
            })

            const m = {...s}
            delete m['uuid']
            return m
        }),
    })
}

// 用于将内存中的case作为go识别的[]byte传给后端
const cases2GoByteUint8 = (stringifyCase: string) => {
    return [Array.from(new TextEncoder().encode(stringifyCase))] as any
}

const getWailsPort = () => {
    // 获取当前应用的完整 URL
    const currentUrl = window.location.href
    console.log('Current URL:', currentUrl)

    // 提取端口
    const getCurrentPort = (): string => {
        const url = new URL(window.location.href)
        return url.port || (url.protocol === 'https:' ? '443' : '80')
    }

    const port = getCurrentPort()
    console.log('Current port:', port)
    return port
}

// 最后封装case到API
const execCaseApi = (message: MessageApiInjection, cases: (TreeOption & ExtraCaseTreeOption)[], dir: string) => {
    const temp_file = Object.fromEntries(cases2GoByteUint8(JSON.stringify(cases.map((c: TreeOption & ExtraCaseTreeOption) => {
        return preSerializeJsonCaseSolve(c)
    }))).map((value, index) => [`key${index}`, value]))

    if (cases.length > 0)
        JsonCaseService.RunRobotTest(ServerAddr.value, ServerPort.value.toString(), RobotPrefix.value, Desc.value, FeiShuGuid.value, LoginTime.value, [], [],
            null, temp_file, RoomOpTime.value, FeiShuNtf.value, DebugLevel.value, DebugLog.value, Concurrency.value
        ).then(res => {
            message.info("用例运行结束")
        })
    else
        JsonCaseService.RunRobotTest(ServerAddr.value, ServerPort.value.toString(), RobotPrefix.value, Desc.value, FeiShuGuid.value, LoginTime.value, [], [],
            dir, {}, RoomOpTime.value, FeiShuNtf.value, DebugLevel.value, DebugLog.value, Concurrency.value
        ).then(res => {
            message.info("用例运行结束")
        })
}

// 保存case
export const saveModifiedCases = (message?: MessageApiInjection) => {
    dataRef.filter((cate: TreeOption & ExtraCaseTreeOption) => {
        return cate.modified || (cate.children && cate.children.find(c => c.modified))
    }).forEach((cate: TreeOption & ExtraCaseTreeOption) => {
        // 保存这个分类, case顺序是tree中的顺序
        const path = cate.fullPath?.split('\\')
        let oldFileName = path?.pop(); // 'xx.json'
        let newFileName = (oldFileName && /.*?.json/.test(oldFileName) && oldFileName != cate.label + '.json') ? cate.label + '.json' : null
        let saveFilePath = JsonsDir.value + '/' + cate.label + '.json'
        if (newFileName && path) {
            path.push(newFileName)
            cate.fullPath = path.join("\\")
            saveFilePath = JsonsDir.value + '/' + oldFileName
        }
        if (cate.children) JsonCaseService.SaveJSONFile(saveFilePath, newFileName, cate.children.filter(c => c.initYanWu)
            .map((c: TreeOption & ExtraCaseTreeOption) => {
                return preSerializeJsonCaseSolve(c)
            })
        ).then(res => {
            dataRef.forEach((c: TreeOption & ExtraCaseTreeOption) => {
                c.modified = false
                if (c.children)
                    c.children.forEach(c => c.modified = false)
            })
            message?.info("保存成功")
        }).catch(err => {
            message?.error(err)
        })
    })
}

// 执行case
export const execNowCase = (message: MessageApiInjection) => {
    // e.stopPropagation() // 阻止菜单关闭
    const nowCaseName = nowCaseData.value ? nowCaseData.value.label : null
    if (nowCaseName) {
        JsonCaseService.IsRunningRobotTest().then(res => {
            if (!res) {
                resetLogCache()
                let findCate = dataRef.find((c: TreeOption & ExtraCaseTreeOption) => c.children?.find(c => c.label === nowCaseName));
                let findCases = findCate ? findCate.children!.filter(_case => _case.label == nowCaseName) : undefined
                if (findCases) {
                    execCaseApi(message, findCases, "")
                    initLogStatistic(findCases)
                    activeTab.value = 'report'
                    message.info("用例开始运行")
                } else {
                    message.error("没找到用例")
                }
            } else {
                message.error("用例正在运行")
            }
        }).catch(err => {
            console.log(err)
            message.error("未知错误")
        })
    }
}

// 执行cases
export const execCateCases = (message: MessageApiInjection, option: TreeOption & ExtraCaseTreeOption, fromMem: boolean) => {
    JsonCaseService.IsRunningRobotTest().then(res => {
        if (!res && option.label) {
            resetLogCache()
            let findCate = dataRef.find((c: TreeOption & ExtraCaseTreeOption) => c.label == option.label);
            if (findCate && findCate.children) {
                if (fromMem)
                    execCaseApi(message, findCate.children, "")
                else if (option.fullPath)
                    execCaseApi(message, [], option.fullPath)
                else console.error("用例路径未找到")
                initLogStatistic(findCate.children)
                activeTab.value = 'report'
                message.info("用例开始运行")
            } else {
                message.error("没找到用例")
            }
        } else {
            message.warning("用例正在运行，请等待或先停止当前用例")
        }
    }).catch(err => {
        console.log(err)
        console.log("未知错误")
    })
}

// 加载case
export const loadCases = async () => {
    // 重新读配置(修竞态:Option.ts 的 GetConfig 是模块级 async,loadCases mount 时可能未 resolve,
    // 用默认值 ../rain-robot 或 "json" → InitExcel 失败 + GetCategories("json") 空 → 用例树默认"测试类别")
    try {
        const cfg = await FuncCaseConfigService.GetConfig()
        if (cfg?.jsons_dir) JsonsDir.value = cfg.jsons_dir
        if (cfg?.excel_resources_dir) ExcelResourcesDir.value = cfg.excel_resources_dir
    } catch (e) {
        console.warn('[Func.ts] loadCases 刷新配置失败', e)
    }
    console.log('[Func.ts] loadCases 开始, ExcelResourcesDir=', ExcelResourcesDir.value)
    // 加载用例前先初始化 Excel 数据（如果尚未初始化）
    if (ExcelResourcesDir.value) {
        console.log('[Func.ts] 调用 InitExcel, path=', ExcelResourcesDir.value)
        try {
            await GameExcelService.InitExcel(ExcelResourcesDir.value)
            console.log('[Func.ts] InitExcel 调用完成')
        } catch (err) {
            console.error('[Func.ts] InitExcel 调用失败, err=', err)
        }

        // 检查 Excel 初始化状态
        try {
            const isInit = await GameExcelService.IsInitialized()
            console.log('[Func.ts] IsInitialized()=', isInit)
            if (!isInit) {
                console.error('[Func.ts] Excel 初始化失败: IsInitialized() 返回 false')
            }
        } catch (err) {
            console.error('[Func.ts] 调用 IsInitialized() 失败, err=', err)
        }
    } else {
        console.warn('[Func.ts] ExcelResourcesDir 为空, 跳过 InitExcel')
    }
    // 刷新前端配置数据
    console.log('[Func.ts] 开始刷新配置数据')
    await Promise.all([refreshHeroConfig(), refreshCardConfig(), refreshSkillConfig()])
    console.log('[Func.ts] 配置数据刷新完成, heroList长度=', excelHeroList?.value?.length, 'cardList长度=', excelCardsList?.value?.length, 'skillList长度=', excelSkillList?.value?.length)

    // 检查刷新后的数据是否为空
    if (!excelHeroList?.value || excelHeroList.value.length === 0) {
        console.warn('[Func.ts] excelHeroList 为空, 可能未正确加载英雄配置')
    }
    if (!excelCardsList?.value || excelCardsList.value.length === 0) {
        console.warn('[Func.ts] excelCardsList 为空, 可能未正确加载卡牌配置')
    }
    if (!excelSkillList?.value || excelSkillList.value.length === 0) {
        console.warn('[Func.ts] excelSkillList 为空, 可能未正确加载技能配置')
    }

    if (JsonsDir.value) JsonCaseService.GetCategories(JsonsDir.value).then(res => {
        nowCaseData.value = {}
        Object.assign(dataRef, res.map((cate, index) => {
            return {
                label: cate.name,
                key: crypto.randomUUID(),
                isLeaf: cate.cases.length == 0,
                // 额外标注
                fullPath: cate.path,
                levelType: 'Categories',
                children: cate.cases.map(c => {
                    return {
                        label: c.case,
                        key: crypto.randomUUID(),
                        isLeaf: true,
                        // 额外标注
                        fullPath: c.fullPath,
                        levelType: 'Cases',
                        category: c.category,
                        // 直接加载case信息
                        caseDesc: c.qaFuncCase.desc,
                        caseManager: c.qaFuncCase.manager,
                        initYanWu: {
                            presentId: c.qaFuncCase.initYanWu.presentId,
                            cardPile: c.qaFuncCase.initYanWu.cardPile,
                            cards: c.qaFuncCase.initYanWu.cards,
                            operateTimeMs: c.qaFuncCase.initYanWu.operateTimeMs,
                            customHeroes: c.qaFuncCase.initYanWu.customHeroes.map(h => {
                                h['uuid'] = crypto.randomUUID()
                                return h
                            }),
                        },
                        caseSteps: c.qaFuncCase.steps.map(s => {
                            if (s.assets)
                                s.assets = s.assets.map(a => {
                                    a['uuid'] = crypto.randomUUID()
                                    return a
                                })
                            s['uuid'] = crypto.randomUUID()
                            return s
                        }),
                    }
                })
            }
        }))
    })
}

// 赋值用例, 同时生成新的UUID
export const copyCase = (option: TreeOption & ExtraCaseTreeOption) => {
    for (const dataElement of dataRef) {
        const children = dataElement.children as TreeOption[]
        const index = children.findIndex(child => child.label === option.label)

        if (index !== -1) {
            // 使用深拷贝函数
            const newVal =  JSON.parse(JSON.stringify(option))

            // 生成新的 UUID
            newVal.key = crypto.randomUUID()

            // 更新自定义英雄的 UUID
            newVal.initYanWu?.customHeroes?.forEach(hero => {
                hero['uuid'] = crypto.randomUUID()
            })

            // 更新步骤和断言的 UUID
            newVal.caseSteps?.forEach(step => {
                step.uuid = crypto.randomUUID()
                step.assets?.forEach(asset => {
                    asset['uuid'] = crypto.randomUUID()
                })
            })

            newVal.label += '_copy'
            newVal.modified = true

            // 插入到找到的元素后面
            children.splice(index + 1, 0, newVal)
            dataElement.modified = true

            break // 找到后就退出
        }
    }
}