<!--
    Settings 页面 - 全局配置管理

    提供飞书通知、MCP 服务、服务端日志和 IPFS 分布式存储的配置界面。
    飞书通知用于测试完成后发送消息到飞书群。
    MCP 服务用于外部 AI 工具调用。
    服务端日志自动捕获 Go 后端输出。
    IPFS 分布式存储提供文件上传和访问的 Demo 功能。
-->
<script setup lang="ts">
import { onMounted, ref } from "vue";
import {
    FeiShuNtf, FeiShuGuid, MCPEnabled, MCPPort, MCPHost,
    MCPRunning, MCPLoading, ExcelDir, loadAllSettings, saveFeishuConfig,
    saveMCPConfig, saveExcelConfig, toggleMCPEnabled
} from "./composables/use-settings";
import {
    interceptEnabled,
    toggleIntercept,
    checkInterceptEnabled,
} from "./composables/use-intercept";
import IpfsPanel from "./components/ipfs-panel.vue";
import WebTorrentPanel from "./components/webtorrent-panel.vue";
import RoadmapPanel from "./components/roadmap-panel.vue";
import { useServerLogs } from "./composables/use-server-logs";
import { useIpfs } from "./composables/use-ipfs";
import { useUpdate } from "./composables/use-update";
import ExampleDataCard from "./components/example-data-card.vue";

onMounted(() => {
    loadAllSettings();
    checkInterceptEnabled();
});

// IPFS 面板显示状态
const showIpfsPanel = ref(false);
// WebTorrent 面板显示状态
const showWebTorrentPanel = ref(false);
// 路线图面板显示状态
const showRoadmapPanel = ref(false);
// 获取服务端日志统计（用于卡片显示）
const { stats, logPanelVisible, statusBarLogEnabled, toggleStatusBarLog } = useServerLogs();
// 获取 IPFS 状态（用于卡片显示）
const { ipfsNodeRunning, ipfsConnectionCount, ipfsUploadHistory, MAX_CONNECTIONS } = useIpfs();
// Android APK 自动更新(仅 Android 端激活,见 composables/use-update.ts)
const {
    isAndroid, checking, downloading, progress, updateInfo,
    errorMsg: updateError, checkUpdate, downloadAndInstall
} = useUpdate();
</script>

<template>
    <div id="settings-page">
        <n-scrollbar style="height: 100%">
            <div class="settings-container">
                <!-- 飞书通知配置卡片 -->
                <n-card title="飞书通知配置" class="setting-card">
                    <div class="setting-row">
                        <span class="setting-label">飞书通知:</span>
                        <n-switch v-model:value="FeiShuNtf" :round="false" @update:value="saveFeishuConfig">
                            <template #checked>开启</template>
                            <template #unchecked>关闭</template>
                        </n-switch>
                    </div>
                    <div class="setting-row">
                        <span class="setting-label">机器人GUID:</span>
                        <n-input v-model:value="FeiShuGuid" placeholder="飞书机器人GUID" @blur="saveFeishuConfig" />
                    </div>
                    <n-divider style="margin: 10px 0" />
                    <div class="setting-row">
                        <span class="setting-label">消息劫持:</span>
                        <n-switch v-model:value="interceptEnabled" :round="false" @update:value="toggleIntercept">
                            <template #checked>开启</template>
                            <template #unchecked>关闭</template>
                        </n-switch>
                    </div>
                    <div class="setting-hint">开启后消息不会发送到飞书，改为弹窗显示（用于测试）</div>
                </n-card>

                <!-- MCP 服务配置卡片 -->
                <n-card title="MCP 服务配置" class="setting-card">
                    <div class="setting-row">
                        <span class="setting-label">启用 MCP 服务:</span>
                        <n-switch v-model:value="MCPEnabled" :round="false" :loading="MCPLoading" @update:value="toggleMCPEnabled">
                            <template #checked>开启</template>
                            <template #unchecked>关闭</template>
                        </n-switch>
                        <n-tag :type="MCPRunning ? 'success' : 'default'" size="small">
                            {{ MCPRunning ? '运行中' : '已停止' }}
                        </n-tag>
                    </div>
                    <div class="setting-row">
                        <span class="setting-label">绑定地址:</span>
                        <n-input v-model:value="MCPHost" placeholder="127.0.0.1" @blur="saveMCPConfig" />
                        <span class="setting-label" style="margin-left: 10px">端口:</span>
                        <n-input-number v-model:value="MCPPort" :min="1" :max="65535" @blur="saveMCPConfig" />
                    </div>
                    <div class="setting-hint">MCP 服务用于外部 AI 工具调用，修改后自动重启</div>
                </n-card>

                <!-- 策划配表目录配置卡片 -->
                <n-card title="策划配表目录" class="setting-card">
                    <div class="setting-row">
                        <span class="setting-label">Excel目录:</span>
                        <n-input v-model:value="ExcelDir" placeholder="如 ../../config" @blur="saveExcelConfig" style="flex: 1;" />
                    </div>
                    <div class="setting-hint">统一配置各模块读取的策划配表目录，proto-test 注入服务器列表等功能使用此路径</div>
                </n-card>

                <!-- 服务端日志卡片 -->
                <n-card title="服务端日志" class="setting-card">
                    <div class="setting-row">
                        <span class="setting-label">已捕获日志:</span>
                        <span style="color: #ccc;">
                            {{ stats.DEBUG + stats.INFO + stats.WARN + stats.ERROR }} 条
                        </span>
                    </div>
                    <div class="setting-row">
                        <span class="setting-label">状态栏实时日志:</span>
                        <n-switch
                            :value="statusBarLogEnabled"
                            :round="false"
                            @update:value="toggleStatusBarLog"
                        >
                            <template #checked>开启</template>
                            <template #unchecked>关闭</template>
                        </n-switch>
                    </div>
                    <n-button type="primary" @click="logPanelVisible = true">
                        查看服务端日志
                    </n-button>
                    <div class="setting-hint">自动捕获 Go 后端输出，支持手动标记重要事件</div>
                </n-card>

                <!-- 应用更新卡片(Android APK 自动更新,仅 Android 端可用) -->
                <n-card title="应用更新" class="setting-card">
                    <template v-if="!isAndroid">
                        <div class="setting-hint">自动更新仅在 Android 端可用(桌面用 wails3 自带 updater)</div>
                    </template>
                    <template v-else>
                        <div class="setting-row">
                            <n-button type="primary" :loading="checking" @click="checkUpdate">检查更新</n-button>
                        </div>
                        <div v-if="updateInfo" style="margin-top: 10px;">
                            <div style="color: #fff;">
                                发现新版本: {{ updateInfo.versionName }}(versionCode {{ updateInfo.versionCode }})
                            </div>
                            <div class="setting-hint">{{ updateInfo.releaseNotes }}</div>
                            <n-button type="success" :loading="downloading" :disabled="downloading" @click="downloadAndInstall" style="margin-top: 8px;">
                                {{ downloading ? ('下载中 ' + progress + '%') : '下载并安装' }}
                            </n-button>
                            <n-progress v-if="downloading" :percentage="progress" :height="6" :show-indicator="false" style="width: 100%; margin-top: 4px;" />
                        </div>
                        <div v-if="updateError" class="setting-hint" style="color: #ff6b6b; margin-top: 8px;">{{ updateError }}</div>
                    </template>
                    <div class="setting-hint" style="margin-top: 8px;">检查 itsnot.fun 服务端版本,下载 APK 后调系统安装器(需用户确认安装)</div>
                </n-card>

                <!-- 加载示例数据卡片(C 方案:内置示例,独立组件方便推翻删除) -->
                <ExampleDataCard />

                <!-- IPFS 分布式存储卡片 -->
                <n-card title="IPFS 分布式存储" class="setting-card">
                    <div class="setting-row">
                        <span class="setting-label">节点状态:</span>
                        <n-tag :type="ipfsNodeRunning ? 'success' : 'default'" size="small">
                            {{ ipfsNodeRunning ? '运行中' : '已停止' }}
                        </n-tag>
                        <span class="setting-label" style="margin-left: 10px">连接数:</span>
                        <span style="color: #ccc;">{{ ipfsConnectionCount }} / {{ MAX_CONNECTIONS }}</span>
                    </div>
                    <div class="setting-row">
                        <span class="setting-label">上传记录:</span>
                        <span style="color: #ccc;">{{ ipfsUploadHistory.length }} 条</span>
                    </div>
                    <n-button type="primary" @click="showIpfsPanel = true">
                        打开 IPFS 面板
                    </n-button>
                    <div class="setting-hint">分布式文件存储 Demo，上传使用 P2P 模式（连接上限 {{ MAX_CONNECTIONS }}），下载使用 HTTP 网关</div>
                </n-card>

                <!-- WebTorrent P2P 传输卡片 -->
                <n-card title="WebTorrent P2P 传输" class="setting-card">
                    <div class="setting-row">
                        <span class="setting-label">传输方式:</span>
                        <n-tag type="info" size="small">WebRTC</n-tag>
                    </div>
                    <n-button type="primary" @click="showWebTorrentPanel = true">
                        打开 P2P 传输面板
                    </n-button>
                    <div class="setting-hint">浏览器端点对点文件传输，支持做种（上传）和 Magnet 链接下载</div>
                </n-card>

                <!-- 开发路线图卡片 -->
                <n-card title="开发路线图" class="setting-card">
                    <n-button type="primary" @click="showRoadmapPanel = true">
                        查看开发路线图
                    </n-button>
                    <div class="setting-hint">功能开发计划与建议，支持投票和评论</div>
                </n-card>
            </div>
        </n-scrollbar>

        <!-- IPFS 分布式存储面板 -->
        <IpfsPanel v-model:visible="showIpfsPanel" />

        <!-- WebTorrent P2P 传输面板 -->
        <WebTorrentPanel v-model:visible="showWebTorrentPanel" />

        <!-- 开发路线图面板 -->
        <RoadmapPanel v-model:visible="showRoadmapPanel" />
    </div>
</template>

<style scoped>
#settings-page {
    height: 100%;
    width: 100%;
}

.settings-container {
    padding: 10px;
}

.setting-card {
    margin-bottom: 10px;
}

.setting-row {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 10px;
}

.setting-label {
    flex-shrink: 0;
    min-width: 100px;
}

.setting-hint {
    color: #999;
    font-size: 12px;
    margin-top: 10px;
}
</style>
