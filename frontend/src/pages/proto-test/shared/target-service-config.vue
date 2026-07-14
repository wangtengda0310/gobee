<template>
  <div style="display: flex; gap: 8px; flex-wrap: wrap; align-items: center;">
    <span style="font-size: 13px; white-space: nowrap;">目标服务:</span>
    <n-input
      :value="serverAddr"
      placeholder="TCP地址 (如 10.254.114.204:18000)"
      style="flex: 2; min-width: 200px;"
      size="small"
      @update:value="handleServerAddrChange"
    />
    <n-input
      :value="httpAddr"
      placeholder="HTTP地址 (如 10.254.114.204:20144)"
      style="flex: 2; min-width: 200px;"
      size="small"
      @update:value="handleHttpAddrChange"
    />
    <n-input
      :value="openID"
      placeholder="登录账号 (如 test)"
      style="flex: 1; min-width: 120px;"
      size="small"
      @update:value="handleOpenIDChange"
    />
    <n-input-number
      :value="rangeStart"
      :min="1"
      placeholder="起始"
      style="width: 80px;"
      size="small"
      @update:value="handleRangeStartChange"
    />
    <n-input-number
      :value="rangeEnd"
      :min="1"
      placeholder="终止"
      style="width: 80px;"
      size="small"
      @update:value="handleRangeEndChange"
    />
    <span v-if="rangeEnd > rangeStart" style="font-size: 12px; color: var(--n-primary-color); white-space: nowrap;">
      共 {{ rangeEnd - rangeStart + 1 }} 个账号
    </span>
    <n-button size="small" data-testid="target-service-settings-btn" @click="showSettings = true">设置</n-button>

    <n-drawer v-model:show="showSettings" :width="360" placement="right">
      <n-drawer-content title="重放设置">
        <div style="display: flex; flex-direction: column; gap: 16px;">
          <!-- 监听端口配置 -->
          <div style="font-size: 14px; font-weight: bold; color: #fff;">监听端口配置</div>
          <div style="display: flex; align-items: center; gap: 8px;">
            <span style="font-size: 13px; white-space: nowrap; min-width: 100px;">TCP监听端口:</span>
            <n-input-number
              :value="tcpListenPort"
              :min="1"
              :max="65535"
              placeholder="TCP监听端口"
              style="flex: 1;"
              size="small"
              @update:value="handleTcpListenPortChange"
            />
          </div>
          <div style="display: flex; align-items: center; gap: 8px;">
            <span style="font-size: 13px; white-space: nowrap; min-width: 100px;">HTTP监听端口:</span>
            <n-input-number
              :value="httpListenPort"
              :min="1"
              :max="65535"
              placeholder="HTTP监听端口"
              style="flex: 1;"
              size="small"
              @update:value="handleHttpListenPortChange"
            />
          </div>
          <div style="display: flex; justify-content: flex-end;">
            <n-button
              size="small"
              type="primary"
              :loading="injecting"
              @click="handleInjectUnityServer"
            >
              注入 unity 服务器列表
            </n-button>
          </div>
          <div style="font-size: 12px; color: #999;">
            点击后会向策划配表的服务器配置表写入/更新 Id=999 的 rain-qa-func 条目，并执行客户端导出。
          </div>
          <div style="font-size: 12px; color: #999;">
            目标服务地址由页面顶部「目标服务」输入框设置，修改后会自动保存并重启监听。
          </div>
          <n-divider style="margin: 8px 0;" />

          <!-- 重放参数配置 -->
          <div style="font-size: 14px; font-weight: bold; color: #fff;">重放参数配置</div>
          <div style="display: flex; align-items: center; gap: 8px;">
            <span style="font-size: 13px; white-space: nowrap; min-width: 100px;">发送间隔(ms):</span>
            <n-input-number
              :value="sendIntervalMs"
              :min="0"
              placeholder="发送间隔"
              style="flex: 1;"
              size="small"
              @update:value="handleSendIntervalChange"
            />
          </div>
          <div style="display: flex; align-items: center; gap: 8px;">
            <span style="font-size: 13px; white-space: nowrap; min-width: 100px;">等待Ack(ms):</span>
            <n-input-number
              :value="ackWaitMs"
              :min="0"
              placeholder="等待Ack"
              style="flex: 1;"
              size="small"
              @update:value="handleAckWaitMsChange"
            />
          </div>
          <div style="display: flex; align-items: center; gap: 8px;">
            <span style="font-size: 13px; white-space: nowrap; min-width: 100px;">最大并发:</span>
            <n-input-number
              :value="maxConcurrency"
              :min="0"
              placeholder="0=不限"
              style="flex: 1;"
              size="small"
              @update:value="handleMaxConcurrencyChange"
            />
          </div>
        </div>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { useMessage, NInput, NInputNumber, NButton, NDrawer, NDrawerContent, NDivider } from 'naive-ui'
import { createWailsReplayControlService } from './replay-control.requirement'
import { createWailsProtoTestConfigService, createWailsRecordControlService } from './protocol-content.requirement'
import { createWailsServerConfigService, type ServerXlsxConfig } from './target-service-config.requirement'

const props = defineProps<{
  serverAddr: string
  httpAddr: string
  openID: string
  rangeStart: number
  rangeEnd: number
  tcpListenPort?: number
  httpListenPort?: number
}>()

const message = useMessage()

// 重放参数设置
const sendIntervalMs = ref(1000)
const ackWaitMs = ref(2000)
const maxConcurrency = ref(0)
const showSettings = ref(false)
const injecting = ref(false)

const emit = defineEmits<{
  'update:serverAddr': [value: string]
  'update:httpAddr': [value: string]
  'update:openID': [value: string]
  'update:rangeStart': [value: number]
  'update:rangeEnd': [value: number]
  'update:tcpListenPort': [value: number]
  'update:httpListenPort': [value: number]
}>()

function handleSendIntervalChange(value: number | null) {
  sendIntervalMs.value = value ?? 1000
  saveReplaySettings()
}

function handleAckWaitMsChange(value: number | null) {
  ackWaitMs.value = value ?? 2000
  saveReplaySettings()
}

function handleMaxConcurrencyChange(value: number | null) {
  maxConcurrency.value = value ?? 0
  saveReplaySettings()
}

async function saveReplaySettings() {
  try {
    const svc = createWailsReplayControlService()
    await svc.setReplaySettings(sendIntervalMs.value, ackWaitMs.value, maxConcurrency.value)
  } catch (e) {
    console.warn('保存重放设置失败:', e)
  }
}

async function handleServerAddrChange(value: string) {
  emit('update:serverAddr', value)
  await nextTick()
  saveProtoTestSettings()
}

async function handleHttpAddrChange(value: string) {
  emit('update:httpAddr', value)
  await nextTick()
  saveProtoTestSettings()
}

function handleOpenIDChange(value: string) {
  emit('update:openID', value)
}

function handleRangeStartChange(value: number | null) {
  emit('update:rangeStart', value ?? 1)
}

function handleRangeEndChange(value: number | null) {
  emit('update:rangeEnd', value ?? 1)
}

async function handleTcpListenPortChange(value: number | null) {
  const port = value ?? 18000
  emit('update:tcpListenPort', port)
  await nextTick()
  saveProtoTestSettings()
}

async function handleHttpListenPortChange(value: number | null) {
  const port = value ?? 20144
  emit('update:httpListenPort', port)
  await nextTick()
  saveProtoTestSettings()
}

async function handleInjectUnityServer() {
  if (injecting.value) return
  injecting.value = true
  try {
    const svc = createWailsServerConfigService()
    await svc.injectUnityServer({
      id: 999,
      server_name: 'rain-qa-func',
      is_save: 1,
      ip_port: '',
      ip_port_hero_point: 'http://localhost:30244',
      keep_alive: 1,
      server_zone_id: 42,
      excel_dir: '',
      http_listen_port: props.httpListenPort ?? 20144,
    } as ServerXlsxConfig)
    await svc.exportClientConfig()
    message.success('已注入 unity 服务器列表并触发客户端导出')
  } catch (e: any) {
    console.error('注入 unity 服务器列表失败:', e)
    message.error('注入失败: ' + (e.message || e))
  } finally {
    injecting.value = false
  }
}

async function saveProtoTestSettings() {
  try {
    const config = {
      tcp_listen_port: props.tcpListenPort ?? 18000,
      http_listen_port: props.httpListenPort ?? 20144,
      target_server_addr: props.serverAddr,
      target_http_addr: props.httpAddr,
    }
    const configSvc = createWailsProtoTestConfigService()
    await configSvc.saveConfig(config)

    // 配置变更后自动重启监听，使新端口/目标地址立即生效
    const recordSvc = createWailsRecordControlService()
    await recordSvc.stopListen()
    await recordSvc.startListen(config.target_server_addr, config.target_http_addr, config.tcp_listen_port, config.http_listen_port)
  } catch (e) {
    console.warn('保存 proto-test 监听配置或重启监听失败:', e)
  }
}

// 组件挂载时从后端加载当前设置
onMounted(async () => {
  try {
    const replaySvc = createWailsReplayControlService()
    const settings = await replaySvc.getReplaySettings()
    if (settings) {
      sendIntervalMs.value = settings.send_interval_ms ?? 1000
      ackWaitMs.value = settings.ack_wait_ms ?? 2000
      maxConcurrency.value = settings.max_concurrency ?? 0
    }
  } catch (e) {
    console.warn('加载重放设置失败:', e)
  }

  try {
    const protoTestSvc = createWailsProtoTestConfigService()
    const config = await protoTestSvc.getConfig()
    if (config) {
      emit('update:tcpListenPort', config.tcp_listen_port)
      emit('update:httpListenPort', config.http_listen_port)
      // 目标服务地址由顶部输入框作为唯一来源
      emit('update:serverAddr', config.target_server_addr)
      emit('update:httpAddr', config.target_http_addr)
    }
  } catch (e) {
    console.warn('加载 proto-test 监听配置失败:', e)
  }
})
</script>
