/**
 * 路线图 API 服务
 *
 * 与后端 Go 服务通信的接口层
 * 使用 Wails 生成的 bindings 类型
 */

// @ts-ignore - Wails 生成的 binding
import * as RoadmapService from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings/roadmap/roadmapservice.js'
// @ts-ignore
import {
	VoteRequest,
	CommentRequest,
	SubmitSuggestionRequest,
	RoadmapItem,
	RoadmapStatus
} from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings/roadmap/models.js'

export type { RoadmapItem, RoadmapStatus, VoteRequest, CommentRequest, SubmitSuggestionRequest }

/**
 * 路线图 API 服务类
 */
class RoadmapServiceClass {
	/**
	 * 获取所有路线图项目
	 */
	async getItems(): Promise<RoadmapItem[]> {
		try {
			const result = await RoadmapService.GetItems()
			return result || []
		} catch (e) {
			console.error('获取路线图项目失败:', e)
			return []
		}
	}

	/**
	 * 获取单个路线图项目
	 */
	async getItem(id: string): Promise<RoadmapItem | null> {
		try {
			return await RoadmapService.GetItem(id)
		} catch (e) {
			console.error('获取项目失败:', e)
			return null
		}
	}

	/**
	 * 投票
	 */
	async vote(itemId: string, vote: string | null): Promise<RoadmapItem | null> {
		try {
			const req = new VoteRequest({ item_id: itemId, vote })
			return await RoadmapService.Vote(req)
		} catch (e) {
			console.error('投票失败:', e)
			return null
		}
	}

	/**
	 * 添加评论
	 */
	async addComment(itemId: string, content: string): Promise<RoadmapItem | null> {
		try {
			const req = new CommentRequest({ item_id: itemId, content })
			return await RoadmapService.AddComment(req)
		} catch (e) {
			console.error('添加评论失败:', e)
			return null
		}
	}

	/**
	 * 提交新建议
	 */
	async submitSuggestion(title: string, description: string, priority: string): Promise<RoadmapItem | null> {
		try {
			// @ts-ignore - Priority enum 兼容
			const req = new SubmitSuggestionRequest({ title, description, priority })
			return await RoadmapService.SubmitSuggestion(req)
		} catch (e) {
			console.error('提交建议失败:', e)
			return null
		}
	}

	/**
	 * 更新项目状态
	 */
	async updateStatus(id: string, status: RoadmapStatus): Promise<RoadmapItem | null> {
		try {
			return await RoadmapService.UpdateStatus(id, status)
		} catch (e) {
			console.error('更新状态失败:', e)
			return null
		}
	}

	/**
	 * 重新加载配置文件
	 */
	async reloadConfig(): Promise<void> {
		try {
			await RoadmapService.ReloadConfig()
		} catch (e) {
			console.error('重新加载配置失败:', e)
		}
	}

	/**
	 * 获取配置文件路径
	 */
	async getConfigFilePath(): Promise<string> {
		try {
			return await RoadmapService.GetConfigFilePath()
		} catch (e) {
			console.error('获取配置路径失败:', e)
			return ''
		}
	}

	/**
	 * 导出为 JSON 格式
	 */
	async exportToJSON(): Promise<string> {
		try {
			return await RoadmapService.ExportToJSON()
		} catch (e) {
			console.error('导出JSON失败:', e)
			return ''
		}
	}
}

// 导出单例
export const roadmapService = new RoadmapServiceClass()
