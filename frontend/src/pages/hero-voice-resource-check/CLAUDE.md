# hero-voice-resource-check -- 武将语音资源检查页面

路由 `/HeroRes`。单文件页面，检查武将语音资源完整性。

## 文件清单

| 文件 | 职责 |
|------|------|
| `index.vue` | 页面入口，包含全部 UI 和逻辑 |

## 关键依赖

- `@bindings/` -- Wails 生成的 Go 后端 bindings
- `@shared/config/hero.ts` -- 武将配置（ID、名称映射）

## 开发注意

- 单文件页面，无需拆分组件
- 如需扩展武将配置字段，修改 `@shared/config/hero.ts`（跨页面共享）

## E2E 测试

| 测试文件 | Page Object | 覆盖范围 |
|----------|-------------|----------|
| `e2e/hero-voice-resource-check/hero-voice-resource-check.spec.ts` | [`HeroVoiceResourceCheckPage`](../../e2e/shared/pages/HeroVoiceResourceCheckPage.ts) | 路径配置（配表路径、Card 文件夹）、执行检索（开始检索/加载状态）、错误列表（错误显示/武将分组/错误详情/音频 ID/重复使用次数）、页面布局、集成测试（完整流程） |

