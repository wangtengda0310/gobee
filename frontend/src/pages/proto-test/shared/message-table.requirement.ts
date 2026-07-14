// message-table.requirement.ts — RecordFileService 只读接口封装
// 职责：加载录制文件数据（只读）

import { LoadRecordFile } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/recordfileservice'
import { RecordFileData } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

// ========== DTO ==========
export interface RecordFileDTO {
	version: number
	recorded_at: string
	server_addr: string
	message_count: number
	messages: RecordEntryViewDTO[]
}

export interface RecordEntryViewDTO {
	index: number
	offset_ms: number
	msg_id: number
	msg_name: string
	seq_id: number
	payload: Record<string, any>
	direction: string
}

// ========== Service 接口 ==========
export interface RecordFileReadService {
	/** 加载录制文件 */
	loadRecordFile(path: string): Promise<RecordFileData>
}

// ========== Wails 实现 ==========
export function createWailsRecordFileReadService(): RecordFileReadService {
	return {
		async loadRecordFile(path: string): Promise<RecordFileData> {
			const data = await LoadRecordFile(path)
			if (!data) throw new Error(`录制文件不存在: ${path}`)
			return data
		},
	}
}

// ========== Mock 实现 ==========
export function createMockRecordFileReadService(): RecordFileReadService {
	return {
		async loadRecordFile(path: string): Promise<RecordFileData> {
			// Mock 返回空数据结构
			return new RecordFileData({
				version: 1,
				recorded_at: new Date().toISOString(),
				server_addr: '',
				message_count: 0,
				messages: [],
			})
		},
	}
}
