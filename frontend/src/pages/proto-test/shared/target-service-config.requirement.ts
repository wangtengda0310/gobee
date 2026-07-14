// target-service-config.requirement.ts
// Stream Proxy 目标服务配置组件的依赖接口定义

import { ServerConfigService } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/server-config'
import type { ServerXlsxConfig } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/server-config'

// ========== DTO ==========
export interface TargetServiceConfig {
  serverAddr: string
  httpAddr: string
  openID: string
  rangeStart: number
  rangeEnd: number
}

export type { ServerXlsxConfig }

// ========== Service 接口 ==========
export interface ServerConfigService {
  injectUnityServer(cfg: ServerXlsxConfig): Promise<void>
  exportClientConfig(excelDir?: string): Promise<void>
}

// ========== Wails 实现 ==========
export function createWailsServerConfigService(): ServerConfigService {
  return {
    async injectUnityServer(cfg: ServerXlsxConfig): Promise<void> {
      await ServerConfigService.InjectUnityServer(cfg)
    },
    async exportClientConfig(excelDir?: string): Promise<void> {
      await ServerConfigService.ExportClientConfig(excelDir || '')
    },
  }
}

// ========== Mock 实现 ==========
export function createMockServerConfigService(): ServerConfigService {
  return {
    async injectUnityServer(_cfg: ServerXlsxConfig): Promise<void> {},
    async exportClientConfig(_excelDir?: string): Promise<void> {},
  }
}

// ========== 组件接口 ==========
export interface TargetServiceConfigProps {
  serverAddr: string
  httpAddr: string
  openID: string
  rangeStart: number
  rangeEnd: number
}

export interface TargetServiceConfigEmits {
  'update:serverAddr': [value: string]
  'update:httpAddr': [value: string]
  'update:openID': [value: string]
  'update:rangeStart': [value: number]
  'update:rangeEnd': [value: number]
  'update:sendIntervalMs': [value: number]
  'update:ackWaitMs': [value: number]
}
