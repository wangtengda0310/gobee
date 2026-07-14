<!--
    InterceptNotification - 飞书消息劫持全局通知条

    在应用右下角显示被劫持消息的通知，点击展开详情面板。
    悬浮于所有页面之上，不自动消失。
-->
<script setup lang="ts">
import {
    showNotification,
    currentMessage,
    unreadCount,
    showDetail,
    openDetail,
    closeNotification,
    interceptEnabled,
} from "../composables/use-intercept";
import InterceptDetail from "./intercept-detail.vue";

// 消息摘要：取前 80 个字符，超出显示省略号
const messageSummary = (content: string) => {
    if (!content) return "";
    const plain = content.replace(/[\n\r]/g, " ").trim();
    return plain.length > 80 ? plain.slice(0, 80) + "..." : plain;
};
</script>

<template>
    <!-- 仅在劫持开关开启且有通知时显示 -->
    <Teleport to="body">
        <Transition name="notification-slide">
            <div
                v-if="showNotification && interceptEnabled"
                class="intercept-notification"
            >
                <!-- 通知条主体 -->
                <div v-if="!showDetail" class="notification-bar" @click="openDetail">
                    <div class="notification-content">
                        <n-tag
                            :type="currentMessage?.msgType === 'text' ? 'info' : 'success'"
                            size="small"
                            :bordered="false"
                        >
                            {{ currentMessage?.msgType === 'text' ? '文本' : '卡片' }}
                        </n-tag>
                        <span class="notification-text">
                            {{ currentMessage ? messageSummary(currentMessage.content) : '' }}
                        </span>
                        <n-tag v-if="unreadCount > 1" size="small" round type="warning">
                            {{ unreadCount }}
                        </n-tag>
                    </div>
                    <button class="notification-close" @click.stop="closeNotification">✕</button>
                </div>

                <!-- 详情面板 -->
                <InterceptDetail
                    v-if="showDetail && currentMessage"
                    :message="currentMessage"
                    @close="closeNotification"
                />
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped>
.intercept-notification {
    position: fixed;
    bottom: 40px;
    right: 16px;
    z-index: 9999;
    max-width: 420px;
}

.notification-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding: 10px 14px;
    background: #1e1e1e;
    border: 1px solid #333;
    border-radius: 8px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.5);
    cursor: pointer;
    transition: all 0.2s ease;
}

.notification-bar:hover {
    border-color: #4ec9b0;
    box-shadow: 0 4px 16px rgba(78, 201, 176, 0.3);
}

.notification-content {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    flex: 1;
}

.notification-text {
    color: #d4d4d4;
    font-size: 13px;
    font-family: 'Cascadia Code', 'SF Mono', Monaco, monospace;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.notification-close {
    flex-shrink: 0;
    background: none;
    border: none;
    color: #858585;
    cursor: pointer;
    font-size: 14px;
    padding: 2px 4px;
    border-radius: 4px;
    transition: all 0.15s ease;
}

.notification-close:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
}

/* 进入/离开动画 */
.notification-slide-enter-active {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.notification-slide-leave-active {
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.notification-slide-enter-from {
    opacity: 0;
    transform: translateY(20px) translateX(20px);
}

.notification-slide-leave-to {
    opacity: 0;
    transform: translateY(10px) translateX(10px);
}

/* 浅色主题适配 */
@media (prefers-color-scheme: light) {
    .notification-bar {
        background: #fff;
        border-color: #e0e0e0;
        box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
    }

    .notification-bar:hover {
        border-color: #18a058;
        box-shadow: 0 4px 16px rgba(24, 160, 88, 0.2);
    }

    .notification-text {
        color: #1a1a1a;
    }

    .notification-close {
        color: #666;
    }

    .notification-close:hover {
        color: #333;
        background: rgba(0, 0, 0, 0.06);
    }
}
</style>
