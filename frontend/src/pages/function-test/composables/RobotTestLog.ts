import {computed, reactive, ref} from 'vue'
import {ExtraCaseTreeOption} from "./use-case-data";
import {TreeOption} from "naive-ui";

export enum LogType {
    None = "None",
    ASSET = "ASSET",
    CORE = "CORE",
    MAIN = "MAIN",
    SYNC = "SYNC",
    STEP = "STEP",
}

export enum LogLevel {
    None = 0,
    DEBUG = 1,
    INFO = 2,
    WARN = 3,
    ERROR = 4,
    PANIC = 5,
}

export type LogMsg = {
    JobTime: string
    JobCase: string
    Description: string
    Case: string
    ID: number
    Level: LogLevel
    Type: LogType
    RobotName: string
    CodeLocation: string
    Time: string
    Msg: string
}

// 修改为按 Case 分组的 Map 结构
export const logCache = reactive<{
    [key: string]: {
        msg: LogMsg
        seq: number
    }[]
}>({})

export const insertLogCache = (seqId: number, message: LogMsg) => {
    const caseName = message.Case;

    if (caseName != 'NO CASE, DEBUG' && !logCache[caseName]) {
        // 获取或创建该 Case 的日志数组
        logCache[caseName] = [];
    }

    if (caseName != 'NO CASE, DEBUG') {
        // 二分查找插入位置
        let left = 0
        let right = logCache[caseName].length
        while (left < right) {
            const mid = Math.floor((left + right) / 2)
            if (logCache[caseName][mid].seq < seqId) {
                left = mid + 1
            } else {
                right = mid
            }
        }

        // 使用 splice 触发响应式更新
        logCache[caseName].splice(left, 0, {seq: seqId, msg: message})

        // 更新对应 Case 的统计信息
        updateCaseStatistics(caseName, message);
    } else {
        // 全部更新
        Object.entries(logCache).forEach(([k, v]) => {
            // 二分查找插入位置
            let left = 0
            let right = v.length
            while (left < right) {
                const mid = Math.floor((left + right) / 2)
                if (v[mid].seq < seqId) {
                    left = mid + 1
                } else {
                    right = mid
                }
            }

            // 使用 splice 触发响应式更新
            v.splice(left, 0, {seq: seqId, msg: message})

            // 更新对应 Case 的统计信息
            updateCaseStatistics(k, message);
        })
    }
}

// 更新单个 Case 的统计信息
const updateCaseStatistics = (caseName: string, message: LogMsg) => {
    const caseIndex = nowRunningCase.value.findIndex(c => c.label === caseName);
    if (caseIndex === -1) return;

    // 更新最大步骤
    nowRunningCaseMaxStep.value[caseIndex] = Math.max(
        nowRunningCaseMaxStep.value[caseIndex] || 0,
        message.ID
    );

    // 更新错误计数
    if (message.Level >= LogLevel.ERROR) {
        nowRunningCaseError.value[caseIndex] = (nowRunningCaseError.value[caseIndex] || 0) + 1;
    }
}

export const resetLogCache = () => {
    Object.keys(logCache).forEach(key => {
        logCache[key] = []
        delete logCache[key]
    })
    nowRunningCase.value = [];
    nowRunningCaseMaxStep.value = [];
    nowRunningCaseError.value = [];
}

// 获取指定 Case 的日志（按顺序）
export const getCaseLogs = (caseName: string) => {
    return logCache[caseName] || [];
}

// 获取所有日志（按 Case 分组）
export const getAllLogs = () => {
    return logCache.value;
}

// ----------------------------下面是根据logCache的统计部分----------------------------------

export const initLogStatistic = (cases: (TreeOption & ExtraCaseTreeOption)[]) => {
    nowRunningCase.value = cases;
    nowRunningCaseMaxStep.value = Array(cases.length).fill(0);
    nowRunningCaseError.value = Array(cases.length).fill(0);
}

export const nowRunningCase = ref<(TreeOption & ExtraCaseTreeOption)[]>([]);
export const nowRunningCaseMaxStep = ref<number[]>([]);
export const nowRunningCaseError = ref<number[]>([]);

// 计算每个 Case 的进度
export const nowRunningCaseProcess = computed(() => {
    return nowRunningCase.value.map((c: TreeOption & ExtraCaseTreeOption, i) => {
        if (!c.caseSteps) {
            return 0;
        }
        return (nowRunningCaseMaxStep.value[i] || 0) / c.caseSteps!.length;
    });
});

// 获取指定 Case 的进度
export const getCaseProcess = (caseName: string) => {
    const caseIndex = nowRunningCase.value.findIndex(c => c.label === caseName);
    if (caseIndex === -1 || !nowRunningCase.value[caseIndex].caseSteps) {
        return 0;
    }
    return (nowRunningCaseMaxStep.value[caseIndex] || 0) / nowRunningCase.value[caseIndex].caseSteps!.length;
}

// 获取指定 Case 的错误数量
export const getCaseErrorCount = (caseName: string) => {
    const caseIndex = nowRunningCase.value.findIndex(c => c.label === caseName);
    return caseIndex === -1 ? 0 : (nowRunningCaseError.value[caseIndex] || 0);
}