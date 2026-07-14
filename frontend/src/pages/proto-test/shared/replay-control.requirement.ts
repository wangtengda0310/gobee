// replay-control.requirement.ts — ReplayControlService 接口封装
// 职责：发送消息重放、停止重放、迭代发送

import {
	SendMessages,
	SendMessagesWithFieldValues,
	SendIterativeMessages,
	StopReplay,
	GetReplaySettings,
	SetReplaySettings,
} from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/replaycontrolservice'
import type { RecordEntryView } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

// ========== DTO ==========
export interface SendMessageItemDTO {
	msg_id: number
	msg_name: string
	payload: Record<string, any>
	offset_ms: number
	direction: string
}

// ========== Service 接口 ==========
export interface ReplayControlService {
	/** 发送消息列表到服务器（替代 StartReplay）。rangeStart/rangeEnd: 账号序号范围，默认1,1=单账号 */
	sendMessages(
		serverAddr: string,
		httpAddr: string,
		openID: string,
		messages: SendMessageItemDTO[],
		repeatCount: number,
		rangeStart?: number,
		rangeEnd?: number
	): Promise<void>
	/** 统一发送消息（后端内部处理迭代/变量展开），接受完整 RecordEntryView 含 field_values */
	sendMessagesWithFieldValues(
		serverAddr: string,
		httpAddr: string,
		openID: string,
		entries: any[],
		repeatCount: number,
		rangeStart: number,
		rangeEnd: number
	): Promise<void>
	/** 迭代发送：根据字段4态配置展开为多条消息后一次性发送 */
	sendIterativeMessages(
		serverAddr: string,
		httpAddr: string,
		openID: string,
		msgId: number,
		msgName: string,
		payloadJson: string,
		direction: string,
		fieldValues: Record<string, any>
	): Promise<void>
	/** 停止重放 */
	stopReplay(): Promise<void>
	/** 获取重放参数设置（SendIntervalMs / AckWaitMs） */
	getReplaySettings(): Promise<{ [key: string]: number | undefined } | null>
	/** 设置重放参数 */
	setReplaySettings(sendIntervalMs: number, ackWaitMs: number, maxConcurrency?: number): Promise<void>
}

// ========== Wails 实现 ==========
export function createWailsReplayControlService(): ReplayControlService {
	return {
		async sendMessages(
			serverAddr: string,
			httpAddr: string,
			openID: string,
			messages: SendMessageItemDTO[],
			repeatCount: number,
			rangeStart?: number,
			rangeEnd?: number
		): Promise<void> {
			const msgs = messages.map((m) => ({
				msg_id: m.msg_id,
				msg_name: m.msg_name,
				payload_json: JSON.stringify(m.payload),
				offset_ms: m.offset_ms,
				direction: m.direction,
			}))
			await SendMessages(serverAddr, httpAddr, openID, JSON.stringify(msgs), repeatCount, rangeStart ?? 1, rangeEnd ?? 1)
		},
		async sendIterativeMessages(
			serverAddr: string,
			httpAddr: string,
			openID: string,
			msgId: number,
			msgName: string,
			payloadJson: string,
			direction: string,
			fieldValues: Record<string, any>
		): Promise<void> {
			// field-item.getFourState() 已返回 snake_case 键名，与 Go FieldValues JSON tag 一致，无需转换
			await SendIterativeMessages(
				serverAddr, httpAddr, openID,
				msgId, msgName, payloadJson, direction,
				fieldValues
			)
		},
		async sendMessagesWithFieldValues(
			serverAddr: string,
			httpAddr: string,
			openID: string,
			entries: any[],
			repeatCount: number,
			rangeStart: number,
			rangeEnd: number
		): Promise<void> {
			await SendMessagesWithFieldValues(serverAddr, httpAddr, openID, entries, repeatCount, rangeStart, rangeEnd)
		},
		async stopReplay(): Promise<void> {
			await StopReplay()
		},
		async getReplaySettings(): Promise<Record<string, number> | null> {
			const result = await GetReplaySettings()
			return result as Record<string, number> | null
		},
		async setReplaySettings(sendIntervalMs: number, ackWaitMs: number, maxConcurrency?: number): Promise<void> {
			await SetReplaySettings(sendIntervalMs, ackWaitMs, maxConcurrency ?? 0)
		},
	}
}

// ========== Mock 实现 ==========
export function createMockReplayControlService(): ReplayControlService {
	return {
		async sendMessages(
			_serverAddr: string,
			_httpAddr: string,
			_openID: string,
			_messages: SendMessageItemDTO[],
			_repeatCount: number,
			_rangeStart?: number,
			_rangeEnd?: number
		): Promise<void> {
			await new Promise((resolve) => setTimeout(resolve, 500))
		},
		async sendIterativeMessages(
			_serverAddr: string,
			_httpAddr: string,
			_openID: string,
			_msgId: number,
			_msgName: string,
			_payloadJson: string,
			_direction: string,
			_fieldValues: Record<string, any>
		): Promise<void> {
			await new Promise((resolve) => setTimeout(resolve, 800))
		},
		async sendMessagesWithFieldValues(
			_serverAddr: string,
			_httpAddr: string,
			_openID: string,
			_entries: any[],
			_repeatCount: number,
			_rangeStart: number,
			_rangeEnd: number
		): Promise<void> {
			await new Promise((resolve) => setTimeout(resolve, 800))
		},
		async stopReplay(): Promise<void> {
			// Mock: no-op
		},
		async getReplaySettings(): Promise<Record<string, number> | null> {
			return { send_interval_ms: 1000, ack_wait_ms: 2000, max_concurrency: 0 }
		},
		async setReplaySettings(_sendIntervalMs: number, _ackWaitMs: number, _maxConcurrency?: number): Promise<void> {
			// Mock: no-op
		},
	}
}
