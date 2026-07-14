<script setup lang="ts">
/**
 * 路线图详情弹窗组件
 *
 * 功能：
 * - 显示功能完整描述
 * - 显示投票详情
 * - 评论列表和添加评论
 */
import { ref, computed } from 'vue'
import { NModal, NButton, NTag, NInput, NScrollbar, useMessage } from 'naive-ui'
import type { RoadmapItem, VoteType } from '../../config/roadmap-types'
import { STATUS_CONFIG, PRIORITY_CONFIG } from '../../config/roadmap-types'
import { roadmapService } from '../../config/roadmap-api'

const props = defineProps<{
  item: RoadmapItem
}>()

const emit = defineEmits<{
  close: []
  vote: [item: RoadmapItem, vote: VoteType]
  commentAdded: [itemId: string]
}>()

const message = useMessage()
const newComment = ref('')
const submitting = ref(false)

// 渲染星星
function renderStars(count: number): string {
  return '★'.repeat(count) + '☆'.repeat(5 - count)
}

// 格式化时间
function formatTime(timestamp: number): string {
  const date = new Date(timestamp)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 投票
function handleVote(vote: VoteType) {
  emit('vote', props.item, vote)
}

// 提交评论
async function submitComment() {
  if (!newComment.value.trim()) {
    message.warning('请输入评论内容')
    return
  }

  submitting.value = true
  try {
    const updated = await roadmapService.addComment(props.item.id, newComment.value.trim())
    if (updated) {
      newComment.value = ''
      message.success('评论成功')
      emit('commentAdded', props.item.id)
    }
  } catch (e) {
    message.error('评论失败')
    console.error(e)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <n-modal
    :show="true"
    @update:show="$emit('close')"
    preset="card"
    style="width: 700px; max-width: 90vw; max-height: 85vh"
    :bordered="false"
    :mask-closable="true"
  >
    <template #header>
      <div class="detail-header">
        <span class="detail-emoji">🚀</span>
        <span class="detail-title">{{ item.title }}</span>
      </div>
    </template>

    <template #header-extra>
      <n-button size="small" quaternary @click="$emit('close')">关闭</n-button>
    </template>

    <n-scrollbar style="max-height: 60vh">
      <div class="detail-content">
        <!-- 元信息 -->
        <div class="detail-meta">
          <span>状态: <n-tag :type="STATUS_CONFIG[item.status].color" size="small">{{ STATUS_CONFIG[item.status].label }}</n-tag></span>
          <span class="meta-divider">|</span>
          <span>优先级: <span :class="'priority priority-' + item.priority">{{ PRIORITY_CONFIG[item.priority].label }} {{ renderStars(PRIORITY_CONFIG[item.priority].stars) }}</span></span>
          <span class="meta-divider">|</span>
          <span>创建者: {{ item.author }}</span>
          <span class="meta-divider">|</span>
          <span>创建时间: {{ formatTime(item.created_at) }}</span>
        </div>

        <!-- 描述 -->
        <div class="detail-section">
          <h4>功能描述</h4>
          <div class="description-box">
            {{ item.description }}
          </div>
        </div>

        <!-- 投票 -->
        <div class="detail-section">
          <h4>投票</h4>
          <div class="vote-section">
            <span class="vote-count">👍 {{ item.votes.up }} 人支持</span>
            <span class="vote-count">👎 {{ item.votes.down }} 人反对</span>
            <div class="vote-buttons">
              <n-button
                :type="item.votes.user_vote === 'up' ? 'success' : 'default'"
                size="small"
                @click="handleVote(item.votes.user_vote === 'up' ? null : 'up')"
              >
                {{ item.votes.user_vote === 'up' ? '已支持' : '支持' }}
              </n-button>
              <n-button
                :type="item.votes.user_vote === 'down' ? 'error' : 'default'"
                size="small"
                @click="handleVote(item.votes.user_vote === 'down' ? null : 'down')"
              >
                {{ item.votes.user_vote === 'down' ? '已反对' : '反对' }}
              </n-button>
            </div>
          </div>
        </div>

        <!-- 评论 -->
        <div class="detail-section">
          <h4>💬 评论区 ({{ item.comments.length }}条)</h4>
          <div class="comment-scroll">
            <div class="comment-list">
              <div v-if="item.comments.length === 0" class="no-comments">
                暂无评论，快来抢沙发吧！
              </div>
              <div v-for="comment in item.comments" :key="comment.id" class="comment-item">
                <div class="comment-header">
                  <span class="comment-author">👤 {{ comment.author }}</span>
                  <span class="comment-time">{{ formatTime(comment.created_at) }}</span>
                </div>
                <div class="comment-content">{{ comment.content }}</div>
              </div>
            </div>
          </div>

          <!-- 添加评论 - 固定在底部 -->
          <div class="add-comment">
            <n-input
              v-model:value="newComment"
              type="textarea"
              placeholder="输入评论..."
              :autosize="{ minRows: 2, maxRows: 4 }"
              @keydown.enter.ctrl="submitComment"
            />
            <n-button
              type="primary"
              size="small"
              :loading="submitting"
              :disabled="!newComment.trim()"
              @click="submitComment"
            >
              发送
            </n-button>
          </div>
        </div>
      </div>
    </n-scrollbar>
  </n-modal>
</template>

<style scoped>
.detail-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.detail-emoji {
  font-size: 20px;
}

.detail-title {
  font-size: 18px;
  font-weight: 500;
}

.detail-content {
  padding: 0 8px;
}

.detail-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 12px 16px;
  background: #252525;
  border-radius: 6px;
  font-size: 13px;
  color: #888;
  margin-bottom: 16px;
}

.meta-divider {
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

.detail-section {
  margin-bottom: 16px;
}

.detail-section h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: #aaa;
  font-weight: 500;
}

.description-box {
  padding: 12px 16px;
  background: #252525;
  border-radius: 6px;
  font-size: 14px;
  line-height: 1.6;
  color: #ccc;
  white-space: pre-wrap;
}

.vote-section {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 16px;
  background: #252525;
  border-radius: 6px;
}

.vote-count {
  font-size: 14px;
  color: #aaa;
}

.vote-buttons {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

.comment-scroll {
  max-height: 180px;
  overflow-y: auto;
  background: #1e1e1e;
  border-radius: 6px;
  margin-bottom: 12px;
}

.comment-list {
  padding: 8px;
}

.no-comments {
  text-align: center;
  color: #666;
  padding: 20px;
}

.comment-item {
  padding: 12px;
  background: #252525;
  border-radius: 6px;
  margin-bottom: 8px;
}

.comment-item:last-child {
  margin-bottom: 0;
}

.comment-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
}

.comment-author {
  font-size: 13px;
  color: #3b82f6;
}

.comment-time {
  font-size: 12px;
  color: #666;
}

.comment-content {
  font-size: 14px;
  color: #ccc;
  line-height: 1.5;
}

.add-comment {
  display: flex;
  gap: 8px;
  padding: 12px;
  background: #252525;
  border-radius: 6px;
  align-items: flex-end;
}

.add-comment :deep(.n-input) {
  flex: 1;
}
</style>
