// paired-payload-editor.requirement.ts — RecordFileService 写操作接口封装
// 职责：更新消息 Payload 并保存录制文件

import {
	UpdateMessagePayload,
	UpdateMessageDescript,
	UpdateMessageFieldValues,
	SaveRecordFile,
} from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/recordfileservice'
import { RecordFileData, FieldValues } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

// ========== DTO ==========
// 复用 models.ts 中的 RecordFileData，无需额外 DTO

// ========== Service 接口 ==========
export interface RecordFileWriteService {
	/** 更新指定 index 消息的 payload，返回更新后的完整数据 */
	updateMessagePayload(path: string, index: number, payload: Record<string, any>): Promise<RecordFileData>
	/** 更新指定 index 消息的描述，返回更新后的完整数据 */
	updateMessageDescript(path: string, index: number, descript: string): Promise<RecordFileData>
	/** 更新指定 index 消息的字段4态值（input_type/range/enum/combo/variable 配置），返回更新后的完整数据 */
	updateMessageFieldValues(path: string, index: number, fieldValues: Record<string, FieldValues>): Promise<RecordFileData>
	/** 保存录制文件 */
	saveRecordFile(path: string, data: RecordFileData): Promise<void>
}

// ========== Wails 实现 ==========
export function createWailsRecordFileWriteService(): RecordFileWriteService {
	return {
		async updateMessagePayload(path: string, index: number, payload: Record<string, any>): Promise<RecordFileData> {
			const data = await UpdateMessagePayload(path, index, payload)
			if (!data) throw new Error('更新 payload 失败')
			return data
		},
		async updateMessageDescript(path: string, index: number, descript: string): Promise<RecordFileData> {
			const data = await UpdateMessageDescript(path, index, descript)
			if (!data) throw new Error('更新描述失败')
			return data
		},
		async updateMessageFieldValues(path: string, index: number, fieldValues: Record<string, FieldValues>): Promise<RecordFileData> {
			const data = await UpdateMessageFieldValues(path, index, fieldValues)
			if (!data) throw new Error('更新字段配置失败')
			return data
		},
		async saveRecordFile(path: string, data: RecordFileData): Promise<void> {
			await SaveRecordFile(path, data)
		},
	}
}

// ========== Mock 实现 ==========
export function createMockRecordFileWriteService(): RecordFileWriteService {
	return {
		async updateMessagePayload(_path: string, _index: number, _payload: Record<string, any>): Promise<RecordFileData> {
			// Mock：返回空数据结构
			return new RecordFileData({
				version: 1,
				recorded_at: new Date().toISOString(),
				server_addr: '',
				message_count: 0,
				messages: [],
			})
		},
		async updateMessageDescript(_path: string, _index: number, _descript: string): Promise<RecordFileData> {
			return new RecordFileData({
				version: 1,
				recorded_at: new Date().toISOString(),
				server_addr: '',
				message_count: 0,
				messages: [],
			})
		},
		async updateMessageFieldValues(_path: string, _index: number, _fieldValues: Record<string, FieldValues>): Promise<RecordFileData> {
			return new RecordFileData({
				version: 1,
				recorded_at: new Date().toISOString(),
				server_addr: '',
				message_count: 0,
				messages: [],
			})
		},
		async saveRecordFile(path: string, data: RecordFileData): Promise<void> {
			// Mock：无操作
		},
	}
}
