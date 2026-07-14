import { computed } from 'vue'
import type { RecordEntryView } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

/** 方向映射 */
const DIR_MAP: Record<string, string> = {
  '→': 'C->S',
  '←': 'S->C',
}

/** 配对条目类型 */
export type PairedEntry = {
  /** 行唯一ID */
  id: number
  /** 消息基础名（如 "Hello"） */
  baseName: string
  /** pair=Req/Ack配对, single=Ntf/其他单行 */
  type: 'pair' | 'single'
  /** Req 部分（type='pair'时有值） */
  req: RecordEntryView | null
  /** Ack 部分（type='pair'时有值，可能为null表示等待中） */
  ack: RecordEntryView | null
  /** Ntf 部分（type='single'时有值） */
  ntf: RecordEntryView | null
  /** 显示用偏移时间 */
  offset_ms: number
  /** 显示用MsgID */
  msg_id: number
  /** 显示用消息名 */
  msg_name: string
  /** 显示用方向（C->S / S->C / C->S,S->C） */
  direction: string
  /** 显示用SeqID */
  seq_id: number
  /** 结果（成功 / 超时 / -） */
  result: string
}

/**
 * 构建配对条目列表
 *
 * 算法：按消息顺序处理
 * - Ntf 结尾的消息 → 单行展示
 * - Req 结尾的消息 → 创建配对行，Ack 留空等待
 * - Ack 结尾的消息 → 向前查找最近一个未匹配的同名 Req，填充 Ack
 * - 其他消息 → 单行展示
 * - 找不到对应 Req 的 Ack → 单行展示
 *
 * TODO: 扩展 result 逻辑，解析 proto payload 中的业务错误码：
 *   - 遍历 Ack payload 中的 result / error_code / ret 等字段
 *   - 将业务错误码（如 -1、10001）作为 result 显示
 *   - 0 / nil 表示成功，非零显示具体错误码
 *
 * @param messages 原始消息列表
 * @param recordedAt 录制起始时间（ISO 字符串），用于计算绝对时间戳
 */
export function buildPairedEntries(messages: RecordEntryView[], recordedAt?: string): PairedEntry[] {
  const pairs: PairedEntry[] = []
  let nextId = 0
  const baseTime = recordedAt ? new Date(recordedAt).getTime() : 0

  for (const msg of messages) {
    const name = msg.msg_name
    const absTime = baseTime ? new Date(baseTime + msg.offset_ms) : null

    if (name.endsWith('Ntf')) {
      // Ntf：单行展示，已成功推送到客户端，视为成功
      pairs.push({
        id: nextId++,
        baseName: name,
        type: 'single',
        req: null,
        ack: null,
        ntf: msg,
        offset_ms: msg.offset_ms,
        msg_id: msg.msg_id,
        msg_name: name,
        direction: DIR_MAP[msg.direction] || msg.direction,
        seq_id: msg.seq_id,
        result: '成功',
      })
    } else if (name.endsWith('Req')) {
      // Req：创建配对行，Ack 留空等待
      const baseName = name.slice(0, -3)
      pairs.push({
        id: nextId++,
        baseName,
        type: 'pair',
        req: msg,
        ack: null,
        ntf: null,
        offset_ms: msg.offset_ms,
        msg_id: msg.msg_id,
        msg_name: name,
        direction: 'C->S', // 只有 Req，无 Ack
        seq_id: msg.seq_id,
        result: '超时',
      })
    } else if (name.endsWith('Ack')) {
      // Ack：向前查找最近一个未匹配的 Req
      const baseName = name.slice(0, -3)
      let matched = false
      for (let i = pairs.length - 1; i >= 0; i--) {
        const p = pairs[i]
        if (p.type === 'pair' && p.baseName === baseName && p.ack === null) {
          p.ack = msg
          p.msg_name = `${p.req?.msg_name} | ${msg.msg_name}`
          p.direction = 'C->S,S->C'
          p.result = '成功'
          matched = true
          break
        }
      }
      // 找不到对应 Req，作为单行展示（S->C）
      if (!matched) {
        pairs.push({
          id: nextId++,
          baseName: name,
          type: 'single',
          req: null,
          ack: msg,
          ntf: null,
          offset_ms: msg.offset_ms,
          msg_id: msg.msg_id,
          msg_name: name,
          direction: DIR_MAP[msg.direction] || msg.direction,
          seq_id: msg.seq_id,
          result: '-',
        })
      }
    } else {
      // 其他消息，单行展示
      pairs.push({
        id: nextId++,
        baseName: name,
        type: 'single',
        req: null,
        ack: null,
        ntf: msg,
        offset_ms: msg.offset_ms,
        msg_id: msg.msg_id,
        msg_name: name,
        direction: DIR_MAP[msg.direction] || msg.direction,
        seq_id: msg.seq_id,
        result: '-',
      })
    }
  }

  return pairs
}

/**
 * 根据配对条目和点击位置获取原始消息索引
 * @param entry 配对的条目
 * @param clickSide 点击的是哪一侧：'req' | 'ack' | 'ntf'
 * @returns 原始消息索引，找不到返回 null
 */
export function getOriginalIndex(entry: PairedEntry, clickSide: 'req' | 'ack' | 'ntf'): number | null {
  if (clickSide === 'req' && entry.req) return entry.req.index
  if (clickSide === 'ack' && entry.ack) return entry.ack.index
  if (clickSide === 'ntf' && entry.ntf) return entry.ntf.index
  return null
}

/**
 * 判断点击位置（用于表格行点击）
 * 对于配对行，根据点击的列判断；对于单行，直接返回 ntf
 */
export function detectClickSide(
  entry: PairedEntry,
  clickColumn: 'req' | 'ack' | 'ntf' | null
): 'req' | 'ack' | 'ntf' {
  if (entry.type === 'single') return 'ntf'
  if (clickColumn === 'req' || clickColumn === 'ack') return clickColumn
  // 默认点击 Req 列
  return 'req'
}
