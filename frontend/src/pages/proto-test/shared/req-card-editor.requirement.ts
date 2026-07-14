// req-card-editor.requirement.ts — Req 卡片编辑器接口封装
// 职责：提供 Req Payload 的卡片式编辑功能（复用 RecordFileWriteService）

// 本组件复用 paired-payload-editor.requirement.ts 中的 RecordFileWriteService
// 无需新增接口，仅作为引用导入
export { createWailsRecordFileWriteService, createMockRecordFileWriteService } from './paired-payload-editor.requirement'
export type { RecordFileWriteService } from './paired-payload-editor.requirement'
