// protocol-content.requirement.ts
// 录制控制服务接口定义

import {
  StartRecord,
  StopRecord,
  GetRecordStatus,
  ReleasePendingMessages,
  ReleaseAllPending,
  SetFilterMode,
  StartListen,
  StopListen,
} from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/recordcontrolservice'
import {
  GetConfig as GetProtoTestConfig,
  SaveConfig as SaveProtoTestConfig,
} from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/prototestconfigservice'
import type { ProtoTestConfig } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

// ========== DTO ==========
export interface RecordProgressDTO {
  status: string
  message_count: number
  error: string
}

export interface ProtoTestConfigDTO {
  tcp_listen_port: number
  http_listen_port: number
  target_server_addr: string
  target_http_addr: string
}

// ========== Service 接口 ==========
export interface RecordControlService {
  startListen(serverAddr: string, httpAddr: string, tcpListenPort: number, httpListenPort: number): Promise<void>
  startRecord(filterMode: boolean): Promise<void>
  stopRecord(): Promise<void>
  stopListen(): Promise<void>
  getRecordStatus(): Promise<RecordProgressDTO>
  releasePendingMessages(connID: number, editsJSON: string): Promise<void>
  releaseAllPending(editsJSON: string): Promise<void>
  setFilterMode(enabled: boolean): Promise<void>
}

export interface ProtoTestConfigService {
  getConfig(): Promise<ProtoTestConfigDTO>
  saveConfig(config: ProtoTestConfigDTO): Promise<void>
}

// ========== Wails 实现 ==========
export function createWailsRecordControlService(): RecordControlService {
  return {
    async startListen(serverAddr: string, httpAddr: string, tcpListenPort: number, httpListenPort: number): Promise<void> {
      await StartListen(serverAddr, httpAddr, tcpListenPort, httpListenPort)
    },
    async startRecord(filterMode: boolean): Promise<void> {
      await StartRecord(filterMode)
    },
    async stopRecord(): Promise<void> {
      await StopRecord()
    },
    async stopListen(): Promise<void> {
      await StopListen()
    },
    async getRecordStatus(): Promise<RecordProgressDTO> {
      const status = await GetRecordStatus()
      if (!status) {
        return {
          status: 'idle',
          message_count: 0,
          error: '',
        }
      }
      return {
        status: status.status,
        message_count: status.message_count,
        error: '',
      }
    },
    async releasePendingMessages(connID: number, editsJSON: string): Promise<void> {
      await ReleasePendingMessages(connID, editsJSON)
    },
    async releaseAllPending(editsJSON: string): Promise<void> {
      await ReleaseAllPending(editsJSON)
    },
    async setFilterMode(enabled: boolean): Promise<void> {
      await SetFilterMode(enabled)
    },
  }
}

export function createWailsProtoTestConfigService(): ProtoTestConfigService {
  return {
    async getConfig(): Promise<ProtoTestConfigDTO> {
      const config = await GetProtoTestConfig()
      if (!config) {
        return {
          tcp_listen_port: 18000,
          http_listen_port: 20144,
          target_server_addr: '10.254.114.204:18000',
          target_http_addr: '10.254.114.204:20144',
        }
      }
      return {
        tcp_listen_port: config.tcp_listen_port ?? 18000,
        http_listen_port: config.http_listen_port ?? 20144,
        target_server_addr: config.target_server_addr ?? '10.254.114.204:18000',
        target_http_addr: config.target_http_addr ?? '10.254.114.204:20144',
      }
    },
    async saveConfig(config: ProtoTestConfigDTO): Promise<void> {
      const payload: ProtoTestConfig = {
        tcp_listen_port: config.tcp_listen_port,
        http_listen_port: config.http_listen_port,
        target_server_addr: config.target_server_addr,
        target_http_addr: config.target_http_addr,
      }
      await SaveProtoTestConfig(payload)
    },
  }
}

// ========== Mock 实现 ==========
export function createMockRecordControlService(): RecordControlService {
  return {
    async startListen(_serverAddr: string, _httpAddr: string, _tcpListenPort: number, _httpListenPort: number): Promise<void> {},
    async startRecord(_filterMode: boolean): Promise<void> {},
    async stopRecord(): Promise<void> {},
    async stopListen(): Promise<void> {},
    async getRecordStatus(): Promise<RecordProgressDTO> {
      return {
        status: 'idle',
        message_count: 0,
        error: '',
      }
    },
    async releasePendingMessages(_connID: number, _editsJSON: string): Promise<void> {},
    async releaseAllPending(_editsJSON: string): Promise<void> {},
    async setFilterMode(_enabled: boolean): Promise<void> {},
  }
}

export function createMockProtoTestConfigService(): ProtoTestConfigService {
  return {
    async getConfig(): Promise<ProtoTestConfigDTO> {
      return {
        tcp_listen_port: 18000,
        http_listen_port: 20144,
        target_server_addr: '10.254.114.204:18000',
        target_http_addr: '10.254.114.204:20144',
      }
    },
    async saveConfig(_config: ProtoTestConfigDTO): Promise<void> {},
  }
}
