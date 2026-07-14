/**
 * 路线图数据类型定义
 *
 * 前端 UI 使用的类型定义，与 Wails bindings 类型保持兼容
 */

/** 功能状态 */
export type RoadmapStatus = 'planning' | 'in_progress' | 'completed' | 'rejected'

/** 优先级 */
export type Priority = 'low' | 'medium' | 'high'

/** 用户投票类型 */
export type VoteType = 'up' | 'down' | null

/** 投票信息 */
export interface Votes {
  up: number
  down: number
  user_vote?: string | null
}

/** 评论 */
export interface Comment {
  id: string
  author: string
  content: string
  created_at: number
}

/** 路线图项目 */
export interface RoadmapItem {
  id: string
  title: string
  description: string
  status: string
  priority: string
  author: string
  created_at: number
  updated_at: number
  votes: Votes
  comments: Comment[]
}

/** 提交新建议请求 */
export interface SubmitSuggestionRequest {
  title: string
  description: string
  priority: string
}

/** 投票请求 */
export interface VoteRequest {
  item_id: string
  vote: string | null
}

/** 评论请求 */
export interface CommentRequest {
  item_id: string
  content: string
}

/** 筛选选项 */
export interface FilterOptions {
  status: string | 'all'
  sortBy: 'votes' | 'date' | 'priority'
  keyword: string
}

/** 状态显示配置 */
export const STATUS_CONFIG: Record<string, { label: string; color: 'info' | 'warning' | 'success' | 'default' }> = {
  planning: { label: '规划中', color: 'info' },
  in_progress: { label: '开发中', color: 'warning' },
  completed: { label: '已完成', color: 'success' },
  rejected: { label: '已拒绝', color: 'default' }
}

/** 优先级显示配置 */
export const PRIORITY_CONFIG: Record<string, { label: string; stars: number }> = {
  low: { label: '低', stars: 1 },
  medium: { label: '中', stars: 3 },
  high: { label: '高', stars: 5 }
}
