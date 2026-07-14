# excel-test/ - 配表测试页面

路由: `/Excel`

## 职责

Excel 配表的检查规则管理、执行和结果展示。支持列规则和表级规则的配置，通过树形结构管理规则。

## 文件清单

### components/
| 文件 | 作用 |
|------|------|
| `excel-check-panel.vue` | 检查面板编排层（组合三个子组件） |
| `table-rule-card.vue` | 表级校验规则卡片（封装 TableRulePanel + 数据加载/同步） |
| `field-card-list.vue` | 字段展示区域（列级规则卡片列表 + 增删/执行检查） |
| `field-anchor-nav.vue` | 右侧锚点导航（纯展示，跳转到对应卡片） |
| `excel-check-log.vue` | 执行日志面板 |
| `excel-check-manager.vue` | 检查管理器 |
| `option-modal.vue` | 选项弹窗 |

### composables/
| 文件 | 作用 |
|------|------|
| `func.ts` | 核心检查逻辑（执行、加载规则） |
| `menu.ts` | 菜单操作 |
| `option.ts` | 选项配置 |
| `excel-rules-template.ts` | 规则类型到参数组件的映射表 |
| `use-excel-check-data.ts` | 检查数据管理 |
| `use-excel-check-log.ts` | 日志管理 |
| `use-tree.ts` | 树操作 |
| `use-tree-and-history.ts` | 树和历史记录 |

### composables/rules/ — 规则参数组件
| 目录/文件 | 作用 |
|-----------|------|
| `*-params.ts` | 每种规则类型的导出适配器（`wrapSFC` / `wrapSFCWithProps`） |
| `components/*.vue` | 对应的 SFC 参数编辑组件 |
| `base/StandardBaseRow.vue` | 标准基础三参数行（允许空值+允许注释+空N行截断） |
| `base/AllBaseRow.vue` | 全基础行（不能为空+允许注释+截断+唯一+中文+自增） |
| `base/init-defaults.ts` | 参数默认值初始化逻辑 |
| `base/types.ts` | `wrapSFC` / `wrapSFCWithProps` 适配器函数 |
| `chain-reference-params.ts` | 关系链共享常量和工具函数（regexOptions、类型定义等） |

## 关键数据流

1. `excel-check-panel.vue` 是编排层，组合 `TableRuleCard`（表级规则）+ `FieldCardList`（字段卡片）+ `FieldAnchorNav`（锚点导航）
2. `excel-rules-template.ts` 是规则类型注册中心：将 `EColRule` 枚举值映射到对应的参数 Vue 组件
3. 检查执行通过 Wails bindings 调用后端 `ExcelCheckService`
4. 树结构管理通过 `use-tree.ts` / `use-tree-and-history.ts` 实现撤销/重做

## 开发注意事项

- **新增规则类型**时，必须同步在 `excel-rules-template.ts` 中注册映射
- `rules/components/` 下每个 .vue 文件对应一种规则类型的参数编辑 UI
- 简单参数组件（只有标准三参数行）复用 `SimpleParams.vue`，通过 `wrapSFCWithProps` 传入 defaults
- 有额外参数的组件各自独立 SFC，包含 `StandardBaseRow` + 额外输入控件
- 规则类型枚举定义在后端 `rain-excel-checker/xlsx/json_rule` 中，前端通过 Wails bindings 获取

## 规则变更同步规范

### 参数组件描述同步

修改或新增校验规则时，**必须同步更新**对应参数 Vue 组件中的描述信息，确保前端展示与后端实际行为一致：

- 组件顶部 `<!-- -->` 注释需准确概括规则行为
- 组件内部"说明"文本（如 `CellTypeCheckParams.vue` 中的说明段落）必须随规则语义变化同步更新
- 若规则名称或行为发生变更，需检查该规则对应的所有参数组件和配置说明

示例：`CellTypeCheckParams.vue` 对应 `EColRule.STRING`，用于检测字符串列中是否存在数值格式单元格，其注释和说明文本需与后端 `StringCheckRule` 的实际检查逻辑保持一致。

### 规则枚举类型注释规范

在 `excel-rules-template.ts` 等规则注册/映射文件中，注释必须标注对应的后端 `EColRule` 枚举类型，便于快速定位规则定义：

```typescript
// EColRule.STRING — 检测字符串类型
{label: '检测字符串类型', value: EColRule.STRING},

// EColRule.STRING — 检测字符串类型
[EColRule.STRING, CellTypeCheckParams],
```

- `ruleOptions` 中的每个选项需用注释标注对应的 `EColRule` 枚举值
- `ruleComponents` 中的每个映射项需用注释标注对应的 `EColRule` 枚举值
- 禁止仅通过 `label` 文字推断规则类型，避免 label 调整后无法快速定位代码

## E2E 测试

| 测试文件 | Page Object | 覆盖范围 |
|----------|-------------|----------|
| `e2e/excel-test/excel-test.spec.ts` | [`ExcelTestPage`](../../e2e/shared/pages/ExcelTestPage.ts) | Header（加载/保存/执行/停止检查、设置弹窗、目录配置）、左侧树面板（搜索、展开/勾选节点、右键菜单）、Tab 面板（负责人管理、用例配置 — 规则增删改/开关/拖拽/锚点、执行日志 — 级别筛选/清空/滚动）、Footer（配表数/Sheet数/成功数/错误数/错误单元格统计）、集成测试（Tab 切换、页面布局） |

## 更多文档

- [配表测试页布局文档](../../../docs/layout/pages/excel-test/index.md) — ASCII 布局可视化、组件层次、数据流