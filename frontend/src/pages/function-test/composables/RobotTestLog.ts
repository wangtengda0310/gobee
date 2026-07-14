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

/** 断言失败日志（与执行日志面板红色断言行一致） */
export const isAssertionFailureLog = (msg: LogMsg) =>
    msg.Level === LogLevel.ERROR && msg.Type === LogType.ASSET

/**
 * 可计入底部「断言错误数目」的失败。
 * 前序断言失败后，同一步骤其余断言常会级联产生「等待消息错误 / 未匹配到的Asset」，
 * 不应与真实判断失败重复累计。
 */
export const isCountableAssertionFailureLog = (msg: LogMsg) => {
    if (!isAssertionFailureLog(msg)) {
        return false
    }
    const text = msg.Msg ?? ''
    return !(
        text.includes('等待消息错误') ||
        text.includes('未匹配到的Asset') ||
        text.includes('等待消息超时')
    )
}

type LogEntry = {
    msg: LogMsg
    seq: number
}

/** 单用例执行期断言统计（底部计数 vs Tab 标红） */
export type CaseRunStats = {
    /** 真实断言判断失败次数（不含超时/未匹配级联） */
    primaryAssertionErrors: number
    /** 是否出现任意断言失败（含级联，供 Tab 标红） */
    hasAnyAssertionFailure: boolean
}

const emptyCaseRunStats = (): CaseRunStats => ({
    primaryAssertionErrors: 0,
    hasAnyAssertionFailure: false,
})

/** 单屏最多渲染的日志条数（更早的仍保留在 logCache 中） */
export const LOG_RENDER_LIMIT = 3000

// 修改为按 Case 分组的 Map 结构
export const logCache = reactive<{
    [key: string]: LogEntry[]
}>({})

/** 批量刷入 logCache 后递增，供日志面板监听（避免 deep watch） */
export const logCacheRevision = ref(0)

const pendingByCase: Record<string, LogEntry[]> = {}
let flushRafId: number | null = null

function mergeSorted(existing: LogEntry[], incoming: LogEntry[]): LogEntry[] {
    if (incoming.length === 0) return existing
    if (existing.length === 0) return incoming.slice()
    incoming.sort((a, b) => a.seq - b.seq)
    const merged: LogEntry[] = []
    let i = 0
    let j = 0
    while (i < existing.length && j < incoming.length) {
        if (existing[i].seq <= incoming[j].seq) {
            merged.push(existing[i++])
        } else {
            merged.push(incoming[j++])
        }
    }
    while (i < existing.length) merged.push(existing[i++])
    while (j < incoming.length) merged.push(incoming[j++])
    return merged
}

function enqueue(caseName: string, seqId: number, message: LogMsg) {
    if (!pendingByCase[caseName]) {
        pendingByCase[caseName] = []
    }
    pendingByCase[caseName].push({seq: seqId, msg: message})
}

function scheduleFlush() {
    if (flushRafId !== null) return
    flushRafId = requestAnimationFrame(() => {
        flushRafId = null
        flushPendingLogs()
    })
}

function flushPendingLogs() {
    let changed = false
    for (const [caseName, batch] of Object.entries(pendingByCase)) {
        if (batch.length === 0) continue
        if (!logCache[caseName]) {
            logCache[caseName] = []
        }
        logCache[caseName] = mergeSorted(logCache[caseName], batch)
        updateCaseStatisticsBatch(caseName, batch)
        batch.length = 0
        changed = true
    }
    if (changed) {
        logCacheRevision.value++
    }
}

/** 用例结束时刷尽待写入日志 */
export const flushLogCacheNow = () => {
    if (flushRafId !== null) {
        cancelAnimationFrame(flushRafId)
        flushRafId = null
    }
    flushPendingLogs()
}

export const insertLogCache = (seqId: number, message: LogMsg) => {
    const caseName = message.Case

    if (caseName === 'NO CASE, DEBUG') {
        const targets = Object.keys(logCache)
        if (targets.length === 0) return
        for (const name of targets) {
            enqueue(name, seqId, message)
        }
    } else {
        enqueue(caseName, seqId, message)
    }
    scheduleFlush()
}

const updateCaseStatisticsBatch = (caseName: string, entries: LogEntry[]) => {
    const caseIndex = nowRunningCase.value.findIndex(c => c.label === caseName)
    if (caseIndex === -1) return

    for (const entry of entries) {
        nowRunningCaseMaxStep.value[caseIndex] = Math.max(
            nowRunningCaseMaxStep.value[caseIndex] || 0,
            entry.msg.ID
        )
        const stats = nowRunningCaseStats.value[caseIndex]
        if (isAssertionFailureLog(entry.msg)) {
            stats.hasAnyAssertionFailure = true
        }
        if (isCountableAssertionFailureLog(entry.msg)) {
            stats.primaryAssertionErrors++
        }
    }
}

export const resetLogCache = () => {
    flushLogCacheNow()
    Object.keys(pendingByCase).forEach(key => {
        delete pendingByCase[key]
    })
    Object.keys(logCache).forEach(key => {
        logCache[key] = []
        delete logCache[key]
    })
    nowRunningCase.value = []
    nowRunningCaseMaxStep.value = []
    nowRunningCaseStats.value = []
    logCacheRevision.value++
}

// 获取指定 Case 的日志（按顺序）
export const getCaseLogs = (caseName: string) => {
    return logCache[caseName] || []
}

/** 供列表渲染：仅返回尾部窗口，避免大批量 DOM */
export const getCaseLogsForDisplay = (caseName: string) => {
    const all = logCache[caseName] || []
    if (all.length <= LOG_RENDER_LIMIT) return all
    return all.slice(-LOG_RENDER_LIMIT)
}

export const getCaseLogOmittedCount = (caseName: string) => {
    const total = (logCache[caseName] || []).length
    return Math.max(0, total - LOG_RENDER_LIMIT)
}

// 获取所有日志（按 Case 分组）
export const getAllLogs = () => {
    return logCache
}

// ----------------------------下面是根据logCache的统计部分----------------------------------

export const initLogStatistic = (cases: (TreeOption & ExtraCaseTreeOption)[]) => {
    nowRunningCase.value = cases
    nowRunningCaseMaxStep.value = Array(cases.length).fill(0)
    nowRunningCaseStats.value = Array.from({length: cases.length}, emptyCaseRunStats)
}

export const nowRunningCase = ref<(TreeOption & ExtraCaseTreeOption)[]>([])
export const nowRunningCaseMaxStep = ref<number[]>([])
export const nowRunningCaseStats = ref<CaseRunStats[]>([])

/** 底部状态栏：全部用例断言判断失败总数 */
export const nowRunningCaseAssertionErrorTotal = computed(() =>
    nowRunningCaseStats.value.reduce((sum, stats) => sum + stats.primaryAssertionErrors, 0)
)

// 计算每个 Case 的进度
export const nowRunningCaseProcess = computed(() => {
    return nowRunningCase.value.map((c: TreeOption & ExtraCaseTreeOption, i) => {
        if (!c.caseSteps) {
            return 0
        }
        return (nowRunningCaseMaxStep.value[i] || 0) / c.caseSteps!.length
    })
})

const isCaseExecutionFinished = (c: TreeOption & ExtraCaseTreeOption, index: number) => {
    const stepCount = c.caseSteps?.length ?? 0
    if (stepCount === 0) {
        return false
    }
    return (nowRunningCaseMaxStep.value[index] || 0) >= stepCount
}

/** 批量执行中已跑完（日志已覆盖全部步骤）的用例数 */
export const nowRunningCaseCompletedCount = computed(() =>
    nowRunningCase.value.reduce(
        (count, c, i) => count + (isCaseExecutionFinished(c, i) ? 1 : 0),
        0,
    ),
)

/** 批量执行计划运行的用例总数 */
export const nowRunningCaseTotalCount = computed(() => nowRunningCase.value.length)

/** 批量执行进度条百分比（按用例数，非步骤平均） */
export const nowRunningCaseBatchProgress = computed(() => {
    const total = nowRunningCaseTotalCount.value
    if (total === 0) {
        return 0
    }
    return (100 * nowRunningCaseCompletedCount.value) / total
})

// 获取指定 Case 的进度
export const getCaseProcess = (caseName: string) => {
    const caseIndex = nowRunningCase.value.findIndex(c => c.label === caseName)
    if (caseIndex === -1 || !nowRunningCase.value[caseIndex].caseSteps) {
        return 0
    }
    return (nowRunningCaseMaxStep.value[caseIndex] || 0) / nowRunningCase.value[caseIndex].caseSteps!.length
}
