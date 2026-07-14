<!--
    InterceptDetail - 单条飞书消息详情面板

    展示被劫持的单条飞书消息完整内容，支持复制。
-->
<script setup lang="ts">
import type { InterceptedMessage } from "../composables/use-intercept";

const props = defineProps<{
    message: InterceptedMessage;
}>();

const emit = defineEmits<{
    close: [];
}>();

// 格式化时间
const formatTime = (timestamp: string) => {
    if (!timestamp) return "";
    return new Date(timestamp).toLocaleString("zh-CN");
};

// 复制内容到剪贴板
const copyContent = async (content: string) => {
    try {
        await navigator.clipboard.writeText(content);
    } catch (err) {
        console.error("复制失败:", err);
    }
};
</script>

<template>
    <div class="intercept-detail">
        <!-- 头部 -->
        <div class="detail-header">
            <div class="detail-meta">
                <n-tag
                    :type="message.msgType === 'text' ? 'info' : 'success'"
                    size="small"
                    :bordered="false"
                >
                    {{ message.msgType === 'text' ? '文本消息' : '卡片消息' }}
                </n-tag>
                <span class="detail-guid">{{ message.robotGuid }}</span>
            </div>
            <div class="detail-actions">
                <span class="detail-time">{{ formatTime(message.timestamp) }}</span>
                <button class="detail-close" @click="emit('close')">✕</button>
            </div>
        </div>

        <!-- 内容 -->
        <div class="detail-body">
            <pre class="detail-content">{{ message.content }}</pre>
        </div>

        <!-- 底部操作 -->
        <div class="detail-footer">
            <n-button size="tiny" @click="copyContent(message.content)">
                复制内容
            </n-button>
            <n-button size="tiny" @click="emit('close')">
                关闭
            </n-button>
        </div>
    </div>
</template>

<style scoped>
.intercept-detail {
    background: #1e1e1e;
    border: 1px solid #333;
    border-radius: 8px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
    overflow: hidden;
}

.detail-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 14px;
    background: #2d2d2d;
    border-bottom: 1px solid #333;
}

.detail-meta {
    display: flex;
    align-items: center;
    gap: 8px;
}

.detail-guid {
    color: #858585;
    font-size: 11px;
    font-family: 'Cascadia Code', 'SF Mono', Monaco, monospace;
}

.detail-actions {
    display: flex;
    align-items: center;
    gap: 10px;
}

.detail-time {
    color: #858585;
    font-size: 11px;
    font-family: 'Cascadia Code', 'SF Mono', Monaco, monospace;
}

.detail-close {
    background: none;
    border: none;
    color: #858585;
    cursor: pointer;
    font-size: 14px;
    padding: 2px 4px;
    border-radius: 4px;
    transition: all 0.15s ease;
}

.detail-close:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
}

.detail-body {
    padding: 12px 14px;
    max-height: 50vh;
    overflow: auto;
}

.detail-content {
    white-space: pre-wrap;
    word-break: break-word;
    margin: 0;
    padding: 12px;
    font-size: 13px;
    line-height: 1.6;
    font-family: 'Cascadia Code', 'SF Mono', Monaco, 'Courier New', monospace;
    color: #d4d4d4;
    background-color: #2d2d2d;
    border: 1px solid #333;
    border-radius: 6px;
}

.detail-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 8px 14px;
    border-top: 1px solid #333;
}

/* 浅色主题适配 */
@media (prefers-color-scheme: light) {
    .intercept-detail {
        background: #fff;
        border-color: #e0e0e0;
        box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
    }

    .detail-header {
        background: #f5f5f7;
        border-bottom-color: #e0e0e0;
    }

    .detail-guid, .detail-time {
        color: #666;
    }

    .detail-close {
        color: #666;
    }

    .detail-close:hover {
        color: #333;
        background: rgba(0, 0, 0, 0.06);
    }

    .detail-content {
        color: #1a1a1a;
        background-color: #f5f5f7;
        border-color: #e0e0e0;
    }

    .detail-footer {
        border-top-color: #e0e0e0;
    }
}
</style>
