/**
 * 飞书消息劫持状态管理模块
 *
 * 管理飞书消息劫持开关和全局通知状态。
 * 收到劫持消息时显示通知条，点击展开单条消息详情。
 */
import { ref } from "vue";
import { Events } from "@wailsio/runtime";
import { InterceptService } from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu";

// 被劫持的消息类型定义
export interface InterceptedMessage {
    id: string;
    robotGuid: string;
    msgType: string;
    content: string;
    timestamp: string;
}

// 响应式状态
export const interceptEnabled = ref(false);

// 通知条状态
export const showNotification = ref(false);
export const currentMessage = ref<InterceptedMessage | null>(null);
export const unreadCount = ref(0);

// 详情面板状态
export const showDetail = ref(false);

// 全局事件监听（模块加载时执行）
// Wails v3 Events.On 回调接收 WailsEvent 对象，含 .name 和 .data 属性
// 参考：@wailsio/runtime/dist/events.js 中 dispatchWailsEvent 实现
Events.On("feishu:intercepted", (ev: any) => {
    const data = ev?.data ?? ev;
    if (data) {
        const msg = Array.isArray(data) ? data[0] as InterceptedMessage : data as InterceptedMessage;
        if (msg) {
            console.log("[Intercept] 收到劫持消息:", msg.id);
            currentMessage.value = msg;
            unreadCount.value++;
            showNotification.value = true;
            showDetail.value = false;
        }
    }
});

/**
 * 切换劫持开关
 */
export const toggleIntercept = async (enabled: boolean) => {
    try {
        await InterceptService.SetEnabled(enabled);
        interceptEnabled.value = enabled;
    } catch (err) {
        console.error("切换劫持开关失败:", err);
        interceptEnabled.value = !enabled;
    }
};

/**
 * 检查劫持开关状态
 */
export const checkInterceptEnabled = async () => {
    try {
        const enabled = await InterceptService.IsEnabled();
        interceptEnabled.value = enabled;
    } catch (err) {
        console.error("检查劫持开关状态失败:", err);
    }
};

/**
 * 打开消息详情
 */
export const openDetail = () => {
    showDetail.value = true;
};

/**
 * 关闭通知条（同时关闭详情面板）
 */
export const closeNotification = () => {
    showNotification.value = false;
    showDetail.value = false;
    unreadCount.value = 0;
};
