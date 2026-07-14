<!--
    开发路线图抽屉面板

    以抽屉方式展示功能路线图列表，支持筛选排序、投票、评论和提交新建议。
-->
<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NButton, NInput, NSelect, NTag, NSpin, NEmpty, NDrawer, NDrawerContent, useMessage } from 'naive-ui'
import type { RoadmapItem, FilterOptions, VoteType } from '../config/roadmap-types'
import { STATUS_CONFIG, PRIORITY_CONFIG } from '../config/roadmap-types'
import { roadmapService } from '../config/roadmap-api'
import RoadmapDetail from './roadmap/RoadmapDetail.vue'
import SubmitModal from './roadmap/SubmitModal.vue'

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
const message = useMessage()

// 数据状态
const items = ref<RoadmapItem[]>([])
const loading = ref(false)
const selectedItem = ref<RoadmapItem | null>(null)
const showDetail = ref(false)
const showSubmitModal = ref(false)

// 筛选状态
const filters = ref<FilterOptions>({
    status: 'all',
    sortBy: 'date',
    keyword: ''
})

// 筛选选项
const statusOptions = [
    { label: '全部状态', value: 'all' },
    { label: '规划中', value: 'planning' },
    { label: '开发中', value: 'in_progress' },
    { label: '已完成', value: 'completed' },
    { label: '已拒绝', value: 'rejected' }
]

const sortOptions = [
    { label: '按时间排序', value: 'date' },
    { label: '按投票排序', value: 'votes' },
    { label: '按优先级排序', value: 'priority' }
]

// 过滤和排序后的列表
const filteredItems = computed(() => {
    let result = [...items.value]

    if (filters.value.status !== 'all') {
        result = result.filter(item => item.status === filters.value.status)
    }

    if (filters.value.keyword.trim()) {
        const keyword = filters.value.keyword.toLowerCase()
        result = result.filter(item =>
            item.title.toLowerCase().includes(keyword) ||
            item.description.toLowerCase().includes(keyword)
        )
    }

    switch (filters.value.sortBy) {
        case 'votes':
            result.sort((a, b) => (b.votes.up - b.votes.down) - (a.votes.up - a.votes.down))
            break
        case 'priority':
            const priorityOrder = { high: 0, medium: 1, low: 2 }
            result.sort((a, b) => priorityOrder[a.priority] - priorityOrder[b.priority])
            break
        case 'date':
        default:
            result.sort((a, b) => b.created_at - a.created_at)
    }

    return result
})

// 加载数据
async function loadData() {
    loading.value = true
    try {
        items.value = await roadmapService.getItems()
    } catch (e) {
        message.error('加载数据失败')
        console.error(e)
    } finally {
        loading.value = false
    }
}

// 查看详情
function viewDetail(item: RoadmapItem) {
    selectedItem.value = item
    showDetail.value = true
}

// 关闭详情
function closeDetail() {
    showDetail.value = false
    selectedItem.value = null
}

// 投票
async function handleVote(item: RoadmapItem, vote: VoteType) {
    try {
        const updated = await roadmapService.vote(item.id, vote)
        if (updated) {
            const index = items.value.findIndex(i => i.id === item.id)
            if (index !== -1) {
                items.value[index] = updated
            }
            if (selectedItem.value?.id === item.id) {
                selectedItem.value = updated
            }
            message.success('投票成功')
        }
    } catch (e) {
        message.error('投票失败')
        console.error(e)
    }
}

// 提交新建议
async function handleSubmit(data: { title: string; description: string; priority: 'low' | 'medium' | 'high' }) {
    try {
        const newItem = await roadmapService.submitSuggestion(data.title, data.description, data.priority)
        if (newItem) {
            items.value.unshift(newItem)
            showSubmitModal.value = false
            message.success('提交成功')
        }
    } catch (e) {
        message.error('提交失败')
        console.error(e)
    }
}

// 添加评论后的回调 - 局部更新
async function onCommentAdded(itemId: string) {
    try {
        const updated = await roadmapService.getItem(itemId)
        if (updated) {
            const index = items.value.findIndex(i => i.id === itemId)
            if (index !== -1) {
                items.value[index] = updated
            }
            if (selectedItem.value?.id === itemId) {
                selectedItem.value = updated
            }
        }
    } catch (e) {
        console.error('更新评论数据失败:', e)
        loadData()
    }
}

// 格式化时间
function formatTime(timestamp: number): string {
    const now = Date.now()
    const diff = now - timestamp
    const days = Math.floor(diff / 86400000)

    if (days === 0) return '今天'
    if (days === 1) return '昨天'
    if (days < 7) return `${days}天前`
    if (days < 30) return `${Math.floor(days / 7)}周前`
    if (days < 365) return `${Math.floor(days / 30)}个月前`
    return `${Math.floor(days / 365)}年前`
}

// 渲染星星
function renderStars(count: number): string {
    return '★'.repeat(count) + '☆'.repeat(5 - count)
}

// 抽屉首次打开时加载数据
watch(() => props.visible, (show) => {
    if (show && items.value.length === 0) {
        loadData()
    }
})
</script>

<template>
    <n-drawer
        :show="visible"
        @update:show="(val: boolean) => emit('update:visible', val)"
        :width="700"
        placement="right"
    >
        <n-drawer-content>
            <template #header>
                <span>开发路线图</span>
            </template>

            <!-- 搜索栏 -->
            <div class="filter-bar">
                <div class="filter-left">
                    <n-select
                        v-model:value="filters.status"
                        :options="statusOptions"
                        style="width: 120px"
                        placeholder="状态筛选"
                    />
                    <n-select
                        v-model:value="filters.sortBy"
                        :options="sortOptions"
                        style="width: 130px"
                        placeholder="排序方式"
                    />
                </div>
                <div class="filter-right">
                    <n-input
                        v-model:value="filters.keyword"
                        placeholder="搜索功能..."
                        clearable
                        style="width: 200px"
                    />
                    <n-button type="primary" size="small" @click="showSubmitModal = true">
                        + 提交新建议
                    </n-button>
                </div>
            </div>

            <!-- 列表区域 -->
            <n-spin :show="loading">
                <div v-if="filteredItems.length === 0 && !loading" class="empty-state">
                    <n-empty description="暂无数据" />
                </div>
                <div v-else class="item-list">
                    <div
                        v-for="item in filteredItems"
                        :key="item.id"
                        class="roadmap-item"
                        @click="viewDetail(item)"
                    >
                        <div class="item-header">
                            <span class="item-emoji">🚀</span>
                            <span class="item-title">{{ item.title }}</span>
                            <n-tag :type="STATUS_CONFIG[item.status].color" size="small">
                                {{ STATUS_CONFIG[item.status].label }}
                            </n-tag>
                        </div>
                        <div class="item-meta">
                            <span class="priority">
                                优先级: <span :class="'priority-' + item.priority">{{ renderStars(PRIORITY_CONFIG[item.priority].stars) }}</span>
                            </span>
                            <span class="divider">|</span>
                            <span class="votes">
                                👍 {{ item.votes.up }} &nbsp; 👎 {{ item.votes.down }}
                            </span>
                            <span class="divider">|</span>
                            <span class="comments">💬 {{ item.comments.length }}</span>
                        </div>
                        <div class="item-desc">
                            {{ item.description.split('\n')[0].slice(0, 80) }}{{ item.description.length > 80 ? '...' : '' }}
                        </div>
                        <div class="item-footer">
                            <span class="latest-comment" v-if="item.comments.length > 0">
                                💬 最新评论: "{{ item.comments[item.comments.length - 1].content.slice(0, 30) }}..."
                            </span>
                            <span class="item-time">{{ formatTime(item.created_at) }}</span>
                        </div>
                        <div class="item-actions" @click.stop>
                            <n-button size="small" @click="handleVote(item, item.votes.user_vote === 'up' ? null : 'up')">
                                {{ item.votes.user_vote === 'up' ? '👍 已支持' : '👍 支持' }}
                            </n-button>
                            <n-button size="small" @click="handleVote(item, item.votes.user_vote === 'down' ? null : 'down')">
                                {{ item.votes.user_vote === 'down' ? '👎 已反对' : '👎 反对' }}
                            </n-button>
                        </div>
                    </div>
                </div>
            </n-spin>

            <!-- 详情弹窗 -->
            <RoadmapDetail
                v-if="showDetail && selectedItem"
                :item="selectedItem"
                @close="closeDetail"
                @vote="handleVote"
                @comment-added="onCommentAdded"
            />

            <!-- 提交弹窗 -->
            <SubmitModal
                v-if="showSubmitModal"
                @close="showSubmitModal = false"
                @submit="handleSubmit"
            />
        </n-drawer-content>
    </n-drawer>
</template>

<style scoped>
.filter-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
    padding: 8px 12px;
    background: #252525;
    border-radius: 8px;
    border: 1px solid #333;
}

.filter-left, .filter-right {
    display: flex;
    gap: 8px;
    align-items: center;
}

.empty-state {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 200px;
}

.item-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.roadmap-item {
    background: #2b2b2b;
    border-radius: 8px;
    padding: 14px;
    border: 1px solid #333;
    cursor: pointer;
    transition: all 0.2s ease;
}

.roadmap-item:hover {
    background: #323232;
    border-color: #444;
    transform: translateY(-1px);
}

.item-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
}

.item-emoji {
    font-size: 16px;
}

.item-title {
    flex: 1;
    font-size: 15px;
    font-weight: 500;
    color: #fff;
}

.item-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: #888;
    margin-bottom: 6px;
}

.divider {
    color: #444;
}

.priority-high {
    color: #f59e0b;
}

.priority-medium {
    color: #3b82f6;
}

.priority-low {
    color: #6b7280;
}

.item-desc {
    font-size: 13px;
    color: #aaa;
    line-height: 1.5;
    margin-bottom: 6px;
}

.item-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 11px;
    color: #666;
    padding-top: 6px;
    border-top: 1px solid #333;
}

.latest-comment {
    color: #888;
}

.item-actions {
    display: flex;
    gap: 8px;
    margin-top: 10px;
}
</style>
