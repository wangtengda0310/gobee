# function-test/ - 战斗测试页面

路由: `/Test`

## 职责

测试用例的创建、编辑、管理和执行。支持演武初始化配置、武将/卡牌/技能选择、步骤动作编辑和机器人执行。

## 文件清单

### components/
| 文件 | 作用 |
|------|------|
| `steps-panel.vue` | 测试步骤编辑面板 |
| `init-yanwu-panel.vue` | 演武初始化配置（座位标题显示身份阵营色圆点，见「座位颜色体系」） |
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
| `use-asset-ai-desc.ts` | 调用后端 `JsonCaseService.GenerateAssetDesc` 生成/缓存断言描述（仅 `asset-card-header-extra.vue` 消费）。watch 仅在 asset 的 uuid/msgName 变化时刷新建议描述，**不 immediate 不 deep**：不 immediate 避免 N 个断言组件批量挂载时各发一次 RPC 卡顿，不 deep 避免「回调同步回写 attr → 跨组件 watch 重入」死循环。AI 建议按需生成（点「应用智能描述」或切断言类型） |

### config/
| 文件 | 作用 |
|------|------|
| `Card.ts` | 卡牌配置（模块加载时从后端获取） |
| `Skill.ts` | 技能配置 |
| `Identity.ts` | 身份配置 + 座位颜色映射（color 下拉框值 + 显示底色，详见「座位颜色体系」） |
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

## 座位颜色体系（座位卡片 / 动作卡片 / 执行日志 + config/Identity.ts）

座位区域涉及**两个独立颜色维度**，勿混淆：

1. **显示底色**（座位标题圆点）— 按「身份大类 IdentityClass」着色，4 色：
   - 主 `#FF4848` / 忠 `#FFCA59` / 反 `#5DD857` / 内 `#6DD1FF`
   - 权威来源：客户端 `D:/work/client/Master/Card/Assets/Scripts/HotUpdate/Logic/Util/ColorUtil.cs` 的 `GetColorByIdentity`
   - 实现：`Identity.ts` 的 `identityClassMap`（identity→大类，来自配表 `IdentityRulesTemplate.json`）+ `classColorHexMap` + `getIdentityColorHex(identity)`
   - 按座位号取色：`getSeatColorHex(customHeroes, seatNo)`（座位号 n → customHeroes[n-1].identity → 阵营色），用于「按座位号」场景
   - 应用（三处统一这套 4 色）：座位卡片标题圆点（`init-yanwu-panel.vue`，按 identity）、动作卡片标题圆点（`steps-panel.vue`，按 step.robotIdx 响应座位下拉）、执行日志 `ID[x]`（`robot-test-log.vue`，按 log.msg.ID）
2. **color 下拉框值**（`CustomHero.color`）— 身份阵营颜色编号（同身份内座位序号），**不是 RGB**：
   - 权威来源：游戏配表 `IdentityEncodeRule_身份编码规则表`（`config/excel/`），前端 `excelIdentityColorMap` 是其复刻
   - 计算公式（客户端 `UIUtils.Room.cs:9` `GetIdentityColor`）：主公系 `8*identity+count-8`、反贼系固定 `8*9-7=65`、黄巾 `8*identity-7`、内奸/刺客/伪帝 `8*identity+count-8`
   - ⚠️ 后端 `reverse_translate.go` 误用 `countryMap`（ECountry 国家表）翻译 color，已知 bug（见 `docs/TODO.md`）

> 关联：后端 `services.go` 的 `CustomHero.Color` 字段注释、`reverse_translate.go` color 解析处均标注了上述语义与 bug。

## E2E 测试

| 测试文件 | Page Object | 覆盖范围 |
|----------|-------------|----------|
| `e2e/function-test/function-test.spec.ts` | [`FunctionTestPage`](../../e2e/shared/pages/FunctionTestPage.ts) | Header（加载/保存/执行/停止用例、设置弹窗）、左侧树面板（搜索、过滤开关、展开/点击节点、右键菜单）、Tab 面板（用例配置 — 名称/描述/负责人/牌堆/武将增删改、用例步骤 — 动作/断言/增删改/拖拽/锚点、执行日志）、Footer（用例数/动作数统计）、布局验证 |
| `e2e/function-test/function-test-steps-layout.spec.ts` | [`FunctionTestPage`](../../e2e/shared/pages/FunctionTestPage.ts) | 用例步骤 Tab：动作/断言卡片标题行统一布局（拖动+智能描述+描述输入）、断言正文仅类型下拉、序号连续 |
| `e2e/function-test/function-test-tooltip-and-divider.spec.ts` | [`FunctionTestPage`](../../e2e/shared/pages/FunctionTestPage.ts) | 下拉分组分割线+染色（卡牌 牌堆灰/座位身份色）、应用智能描述按钮 tooltip、技能下拉选项 tooltip、资产断言「卡」下拉分组（hover 文字区触发） |
