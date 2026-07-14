<!--
    服务端日志抽屉面板

    以抽屉方式展示 Go 后端实时推送的服务端日志。
    支持日志级别过滤、关键词搜索、自动滚动、清除日志等功能。
    独立运行的日志面板组件。
-->
<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import { DeleteOutlined } from '@vicons/material'
import { useServerLogs, clearServerLogs } from '../composables/use-server-logs'

// Props
interface Props {
    visible?: boolean
}

const props = withDefaults(defineProps<Props>(), {
    visible: false
})

// Emits
interface Emits {
    (e: 'update:visible', value: boolean): void
}

const emit = defineEmits<Emits>()

// 使用服务端日志管理
const { filteredLogs, stats, levelFilters, searchQuery } = useServerLogs()

// 滚动条引用
const scrollbarRef = ref<InstanceType<typeof import('naive-ui').NScrollbar>>()
const contentRef = ref<HTMLElement>()

// 自动滚动
let autoScroll = true

// 监听日志变化自动滚动
watch(filteredLogs, () => {
    if (autoScroll) {
        nextTick(() => scrollToBottom())
    }
}, { deep: true })

// 滚动到底部
const scrollToBottom = () => {
    if (scrollbarRef.value && contentRef.value) {
        scrollbarRef.value.scrollTo({
            top: contentRef.value.scrollHeight,
            behavior: 'smooth'
        })
    }
}

// 滚动事件处理：用户上滚暂停自动滚动，到底部恢复
const handleScroll = (e: Event) => {
    const target = e.target as HTMLElement
    const isAtBottom = target.scrollHeight - target.scrollTop - target.clientHeight < 50
    autoScroll = isAtBottom
}

// 清除日志
const handleClearLogs = () => {
    clearServerLogs()
}
</script>

<template>
    <n-drawer
        :show="visible"
        @update:show="(val: boolean) => emit('update:visible', val)"
        :width="1000"
        placement="right"
    >
        <n-drawer-content>
            <template #header>
                <div style="display: flex; justify-content: space-between; align-items: center;">
                    <span>服务端日志</span>
                    <n-button size="small" @click="handleClearLogs" quaternary type="error">
                        <template #icon>
                            <n-icon><DeleteOutlined /></n-icon>
                        </template>
                        清除日志
                    </n-button>
                </div>
            </template>

            <!-- 工具栏：级别过滤 + 搜索 -->
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; gap: 12px;">
                <div style="display: flex; gap: 12px;">
                    <n-checkbox v-model:checked="levelFilters.DEBUG" size="small">
                        <span style="color: #888">DEBUG</span>
                    </n-checkbox>
                    <n-checkbox v-model:checked="levelFilters.INFO" size="small">
                        <span style="color: #66ff5c">INFO</span>
                    </n-checkbox>
                    <n-checkbox v-model:checked="levelFilters.WARN" size="small">
                        <span style="color: #ffc74d">WARN</span>
                    </n-checkbox>
                    <n-checkbox v-model:checked="levelFilters.ERROR" size="small">
                        <span style="color: #ff4141">ERROR</span>
                    </n-checkbox>
                </div>
                <n-input
                    v-model:value="searchQuery"
                    placeholder="搜索日志..."
                    clearable
                    size="small"
                    style="width: 200px;"
                />
            </div>

            <!-- 日志内容 -->
            <n-scrollbar
                ref="scrollbarRef"
                style="height: calc(100vh - 180px)"
                @scroll="handleScroll"
            >
                <div ref="contentRef" style="padding: 8px; font-family: monospace; font-size: 12px;">
                    <div
                        v-for="(log, index) in filteredLogs"
                        :key="index"
                        :style="{
                            color: log.isManual ? '#8f8' : (
                                log.level === 'ERROR' ? '#ff4141' :
                                log.level === 'WARN' ? '#ffc74d' :
                                log.level === 'INFO' ? '#66ff5c' : '#888'
                            ),
                            background: log.isManual ? 'rgba(30, 58, 30, 0.5)' : 'transparent',
                            padding: log.isManual ? '2px 4px' : '0',
                            borderRadius: log.isManual ? '2px' : '0',
                        }"
                        style="margin-bottom: 3px; line-height: 1.6; word-break: break-all;"
                    >
                        <span style="color: #555">[{{ log.timestamp }}]</span>
                        <span style="font-weight: bold;">[{{ log.level }}]</span>
                        {{ log.message }}
                    </div>
                    <div v-if="filteredLogs.length === 0" style="color: #888; text-align: center; margin-top: 40px;">
                        暂无日志
                    </div>
                </div>
            </n-scrollbar>

            <!-- 底部统计 -->
            <template #footer>
                <div style="display: flex; justify-content: space-between; align-items: center; font-size: 12px; color: #999;">
                    <span>
                        共 {{ filteredLogs.length }} 条日志
                        (DEBUG: {{ stats.DEBUG }} | INFO: {{ stats.INFO }} | WARN: {{ stats.WARN }} | ERROR: {{ stats.ERROR }})
                    </span>
                    <span v-if="!autoScroll" style="color: #ffc74d;">自动滚动已暂停</span>
                </div>
            </template>
        </n-drawer-content>
    </n-drawer>
</template>

<style scoped>
:deep(.n-drawer-content) {
    display: flex;
    flex-direction: column;
}
</style>
