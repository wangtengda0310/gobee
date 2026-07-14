/**
 * 服务端日志管理 Composable
 *
 * 监听 serverLog 事件，管理服务端日志的接收、过滤和搜索。
 * 独立运行的日志管理模块。
 */
import { reactive, ref, computed } from 'vue'
import { Events } from '@wailsio/runtime'

// 服务端日志条目类型（对应 Go 后端 serverlog.LogEntry）
export interface ServerLogEntry {
    level: string      // DEBUG | INFO | WARN | ERROR
    message: string
    timestamp: string
    isManual: boolean  // 手动标记的日志用 ▸ 前缀
}

// 级别过滤状态（模块级单例，跨组件共享）
const levelFilters = reactive({
    DEBUG: true,
    INFO: true,
    WARN: true,
    ERROR: true,
})

// 搜索关键词（模块级单例）
const searchQuery = ref('')

// 日志缓存（模块级单例，跨组件共享）
const serverLogs = reactive<ServerLogEntry[]>([])
const MAX_LOGS = 2000

// 状态栏日志开关（持久化到 localStorage，默认关闭）
const statusBarLogEnabled = ref(
    localStorage.getItem('statusBarLogEnabled') === 'true'
)

// 日志面板可见状态（模块级单例，跨组件共享）
const logPanelVisible = ref(false)

// 切换状态栏日志开关
const toggleStatusBarLog = (enabled: boolean) => {
    statusBarLogEnabled.value = enabled
    try {
        localStorage.setItem('statusBarLogEnabled', String(enabled))
    } catch {
        // localStorage 不可用时静默失败
    }
}

// 获取日志级别颜色
export const getServerLogColor = (level: string, isManual: boolean): string => {
    if (isManual) return '#8f8'
    switch (level) {
        case 'ERROR': return '#ff4141'
        case 'WARN': return '#ffc74d'
        case 'INFO': return '#66ff5c'
        case 'DEBUG': return '#888888'
        default: return '#ffffff'
    }
}

// 插入日志（超过上限自动丢弃最旧的）
const insertLog = (entry: ServerLogEntry) => {
    serverLogs.push(entry)
    if (serverLogs.length > MAX_LOGS) {
        serverLogs.shift()
    }
}

// 清空日志
export const clearServerLogs = () => {
    serverLogs.splice(0, serverLogs.length)
}

// 最新一条日志（供状态栏使用，避免监听整个数组）
const latestLog = computed<ServerLogEntry | null>(() =>
    serverLogs.length > 0 ? serverLogs[serverLogs.length - 1] : null
)

// 全局事件监听（模块加载时执行，与 use-intercept.ts 模式一致）
// Wails v3 Events.On 回调接收 WailsEvent 对象，含 .data 属性
// 参考：@wailsio/runtime/dist/events.js 中 dispatchWailsEvent 实现
Events.On('serverLog', (ev: any) => {
    // 兼容多种数据格式：ev.data 可能是数组、对象或直接就是 LogEntry
    const payload = ev?.data ?? ev
    if (!payload) return

    let entry: ServerLogEntry
    if (Array.isArray(payload)) {
        // Wails v3 多参数 Emit 格式: { data: [arg1, arg2, ...] }
        entry = payload[0] as ServerLogEntry
    } else if (payload && typeof payload === 'object' && payload.level) {
        // 直接是 LogEntry 对象
        entry = payload as ServerLogEntry
    } else {
        console.warn('[use-server-logs] 无法解析 serverLog 事件:', ev)
        return
    }

    insertLog(entry)
})

// 日志管理 Hook
export const useServerLogs = () => {

    // 过滤后的日志（级别过滤 + 关键词搜索）
    const filteredLogs = computed(() => {
        let result = serverLogs

        // 级别过滤
        if (!levelFilters.DEBUG || !levelFilters.INFO || !levelFilters.WARN || !levelFilters.ERROR) {
            result = result.filter(log => levelFilters[log.level as keyof typeof levelFilters])
        }

        // 关键词搜索
        if (searchQuery.value.trim()) {
            const query = searchQuery.value.toLowerCase()
            result = result.filter(log =>
                log.message.toLowerCase().includes(query) ||
                log.timestamp.toLowerCase().includes(query)
            )
        }

        return result
    })

    // 统计信息
    const stats = computed(() => ({
        DEBUG: serverLogs.filter(log => log.level === 'DEBUG').length,
        INFO: serverLogs.filter(log => log.level === 'INFO').length,
        WARN: serverLogs.filter(log => log.level === 'WARN').length,
        ERROR: serverLogs.filter(log => log.level === 'ERROR').length,
    }))

    return {
        serverLogs,
        filteredLogs,
        stats,
        levelFilters,
        searchQuery,
        clearServerLogs,
        getServerLogColor,
        // 新增
        latestLog,
        logPanelVisible,
        statusBarLogEnabled,
        toggleStatusBarLog,
    }
}
