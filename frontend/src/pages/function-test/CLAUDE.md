# function-test/ - 战斗测试页面

路由: `/Test`

## 职责

测试用例的创建、编辑、管理和执行。支持演武初始化配置、武将/卡牌/技能选择、步骤动作编辑和机器人执行。

## 文件清单

### components/
| 文件 | 作用 |
|------|------|
| `steps-panel.vue` | 测试步骤编辑面板 |
| `init-yanwu-panel.vue` | 演武初始化配置 |
| `asset-card.vue` | 资产卡牌显示 |
| `footer-case-log-statistic.vue` | 底部用例日志统计 |
| `robot-test-log.vue` | 机器人测试日志 |
| `asset-sections/` | 11 个资产变更区段 Vue 组件 |
| `modals/` | 弹窗组件（添加用例、分类、重命名等） |

### composables/
| 文件 | 作用 |
|------|------|
| `Func.ts` | 核心业务逻辑：用例加载/保存/执行，协调所有 composable |
| `Menu.ts` | 菜单操作 |
| `Option.ts` | 选项配置（服务器地址、端口等） |
| `Modals.ts` | 弹窗状态管理 |
| `Tree.ts` | 用例树操作 |
| `AssetProtoOptions.ts` | 资产 proto 枚举选项 |
| `AssetMapTrans.ts` | 资产映射转换 |
| `HeroAndCardsAndSkillsSelect.ts` | 武将/卡牌/技能选择逻辑 |
| `StepActionsAndAssetsSelect.ts` | 步骤动作和资产选择 |
| `RobotTestLog.ts` | 机器人测试日志管理 |
| `use-case-data.ts` | 用例数据 composable |
| `use-asset-ai-desc.ts` | AI 描述生成 |

### config/
| 文件 | 作用 |
|------|------|
| `Card.ts` | 卡牌配置（模块加载时从后端获取） |
| `Skill.ts` | 技能配置 |
| `Identity.ts` | 身份配置（含颜色映射） |
| `ECountry.ts` | 国家/阵营配置 |
| `ErrorCode.ts` | 错误码配置 |
| 武将配置 | 共享 `@shared/config/hero.ts` |

## 关键数据流

1. `Func.ts` 是中枢，协调 `Tree`、`Option`、`Modals` 等 composable
2. `config/` 下文件在模块加载时通过 Wails bindings 从后端获取数据，导出响应式列表
3. 用例数据（JSON）通过 `JsonCaseService` binding 读写，格式定义见 `cases/fight_cases_schema.json`
4. 飞书通知配置从 `settings/composables/use-settings` 导入

## 开发注意事项

- `Func.ts` 体积较大，修改前需理解其与其他 composable 的调用关系
- `config/` 下的配置数据是模块级副作用（import 即触发后端请求）
- 新增资产类型需要同步更新 `AssetProtoOptions.ts` 和对应的 `asset-sections/` 组件
- 武将配置在 `@shared/config/hero.ts`，不在本目录内

## E2E 测试

| 测试文件 | Page Object | 覆盖范围 |
|----------|-------------|----------|
| `e2e/function-test/function-test.spec.ts` | [`FunctionTestPage`](../../e2e/shared/pages/FunctionTestPage.ts) | Header（加载/保存/执行/停止用例、设置弹窗）、左侧树面板（搜索、过滤开关、展开/点击节点、右键菜单）、Tab 面板（用例配置 — 名称/描述/负责人/牌堆/武将增删改、用例步骤 — 动作/断言/增删改/拖拽/锚点、执行日志）、Footer（用例数/动作数统计）、布局验证 |
