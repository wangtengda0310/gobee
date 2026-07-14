// case-selector.requirement.ts — TestCaseService 接口封装
// 职责：测试用例列表的加载、保存、读取、删除

import {
	LoadTestCaseList,
	SaveTestCase,
	AppendTestCase,
	LoadTestCase,
	DeleteTestCase,
} from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/testcaseservice'
import { RecordFileData, ProtoCaseMeta } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

// ========== DTO ==========
export interface CaseMetaDTO {
	name: string
	message_count: number
	server_addr: string
	created_at: string
}

// ========== Service 接口 ==========
export interface TestCaseService {
	/** 加载所有测试用例列表 */
	loadCaseList(): Promise<CaseMetaDTO[]>
	/** 保存测试用例（覆盖写入，用于 testcase-tab 的增删改排序） */
	saveTestCase(name: string, data: RecordFileData): Promise<void>
	/** 向已存在用例追加 Req 消息；用例不存在时创建新用例（用于 packet-tab 的"保存到用例"） */
	appendTestCase(name: string, data: RecordFileData): Promise<void>
	/** 加载指定测试用例 */
	loadTestCase(name: string): Promise<RecordFileData>
	/** 删除指定测试用例 */
	deleteTestCase(name: string): Promise<void>
}

// ========== Wails 实现 ==========
export function createWailsTestCaseService(): TestCaseService {
	return {
		async loadCaseList(): Promise<CaseMetaDTO[]> {
			const list = await LoadTestCaseList()
			return (list ?? []).filter((item): item is ProtoCaseMeta => item !== null).map((item) => ({
				name: item.name,
				message_count: item.message_count,
				server_addr: item.server_addr,
				created_at: item.created_at,
			}))
		},
		async saveTestCase(name: string, data: RecordFileData): Promise<void> {
			await SaveTestCase(name, data)
		},
		async appendTestCase(name: string, data: RecordFileData): Promise<void> {
			await AppendTestCase(name, data)
		},
		async loadTestCase(name: string): Promise<RecordFileData> {
			const data = await LoadTestCase(name)
			if (!data) throw new Error(`用例不存在: ${name}`)
			return data
		},
		async deleteTestCase(name: string): Promise<void> {
			await DeleteTestCase(name)
		},
	}
}

// ========== Mock 实现 ==========
export function createMockTestCaseService(): TestCaseService {
	const mockCases: Map<string, RecordFileData> = new Map()
	return {
		async loadCaseList(): Promise<CaseMetaDTO[]> {
			return Array.from(mockCases.entries()).map(([name, data]) => ({
				name,
				message_count: data.message_count,
				server_addr: data.server_addr,
				created_at: data.recorded_at,
			}))
		},
		async saveTestCase(name: string, data: RecordFileData): Promise<void> {
			mockCases.set(name, data)
		},
		async appendTestCase(name: string, data: RecordFileData): Promise<void> {
			const existing = mockCases.get(name)
			if (existing) {
				mockCases.set(name, {
					...existing,
					message_count: existing.message_count + data.message_count,
					messages: [...existing.messages, ...data.messages],
				})
			} else {
				mockCases.set(name, data)
			}
		},
		async loadTestCase(name: string): Promise<RecordFileData> {
			const data = mockCases.get(name)
			if (!data) throw new Error(`Mock 用例不存在: ${name}`)
			return data
		},
		async deleteTestCase(name: string): Promise<void> {
			mockCases.delete(name)
		},
	}
}
