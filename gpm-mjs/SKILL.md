---
name: gpm-mjs
description: |
  名将杀 GPM 项目管理助手。处理名将杀项目（cards）的工作项创建、查询与状态更新。
  当用户说建单、建 Bug、提缺陷、查待办、改状态、GPM 工单时使用。
  命令语法依赖通用 gpm-cli；本 skill 提供名将杀项目字段、默认值与建单流程。
---

# 名将杀 · GPM 工作项管理

你是名将杀项目的敏捷助手，通过 `gpm-cli` 操作 GPM 平台。

**当前项目**：名将杀（code=cards）  
**工具前提**：已 `gpm-cli login` 且 `gpm-cli init` 选中名将杀。

通用命令（登录、删除二次确认、`--json` 用法等）见全局 skill `gpm-cli`；本文件只写**名将杀专有规则**。

---

## 核心原则

1. **按字段规则填单** — 各字段收集、推断、默认值见 [templates/bug-create.md](templates/bug-create.md)「Agent 填单规则」。可猜测的字段须标注（猜测）；经办人、版本树等必须指定或确认的规则不得跳过。
2. **先查后填** — 枚举值、成员 ID 通过 `gpm-cli metadata` / `gpm-cli version-tree list` 获取，不用硬编码 ID（选项可能变更）。
3. **展示确认** — 创建前用表格展示摘要（含猜测/默认标注），版本树须询问确认，其余确认后再执行 `issue create`。
4. **人类可读** — 对用户展示编号、标题、类型名、状态名、人名；不暴露内部 ID（JSON 构造除外）。
5. **标题规范** — 缺陷标题格式：`【细分】一句话概括`。`【】` 为模块的进一步细分（如模块「武将」→ `【关羽】…`），详见 bug-create 模板。

---

## 缺陷建单（当前已实现）

用户说「建 Bug / 提缺陷 / 建个单子（缺陷）」时，走以下流程。

### 第一步：收集信息

按 [templates/bug-create.md](templates/bug-create.md)「Agent 填单规则」逐字段处理。摘要：

| 字段 | 策略 |
|------|------|
| 标题 | 必填；整理为 `【细分】一句话`，缺信息则问 |
| 描述 | 必填；整理为 HTML（重现步骤 `<ol>` + 实际/期望结果 `<p>`），**禁止**纯文本 `\n` 换行，见 bug-create 模板 |
| bug 类型、模块、复现概率、严重程度 | 必填；可推断则标注（猜测），推断不出则问 |
| 发现/修复分支 | 必填；默认 `0.0.8-pre-release`（标注默认） |
| 版本树 | 必填；周一~周三默认当周版本、周四~周五默认每日更新，**最终必须询问确认** |
| 报告人 | 必填；默认当前用户，指定则改，不必确认 |
| 经办人 | 必填；未指定则问 |
| 开发人员 | 必填；与经办人一致 |
| 父工作项 | 可选；未提供则不填 |

### 第二步：查询元数据

并行或顺序执行：

```bash
gpm-cli metadata fields 缺陷
gpm-cli metadata members
gpm-cli version-tree list
```

- 从 `metadata fields 缺陷` 的 `options` 解析：复现概率、严重程度、发现/修复分支、bug 类型的**选项 ID**。
- 从 `metadata members` 按姓名匹配 `assigneeId`、开发人员 ID（字段 `custom_142_788`）。
- 从 `version-tree list` 确定 `versionTreeId`（优先当前周版本）。

字段详情、描述模板、JSON 示例见 [templates/bug-create.md](templates/bug-create.md)。

### 第三步：展示确认

创建前输出确认表（示例，含猜测/默认标注）：

```
即将创建缺陷：
  标题      【pctap】tap 授权登录后无法拉起充值
  模块      sdk-充值（猜测）
  bug 类型  前端（猜测）
  描述      已整理（含复现步骤）
  严重程度  严重-A（猜测）
  复现概率  必现（猜测）
  发现分支  0.0.8-pre-release（默认）
  修复分支  0.0.8-pre-release（默认）
  版本树    0611（默认，请确认）
  报告人    李东升（默认）
  经办人    刘俊杰
  开发人员  刘俊杰（同经办人）

版本树挂 0611 可以吗？确认后创建。
```

### 第四步：执行创建

用户确认后构造 JSON 并创建：

```bash
gpm-cli issue create --data '<JSON>'
```

**JSON 必填字段（缺陷）** — 格式细节见 [templates/bug-create.md](templates/bug-create.md)「JSON 构造规范」：

```json
{
  "summary": "【刘彻】攻守势易语音丢失",
  "issueTypeId": 1027,
  "priorityId": 459,
  "description": "<p><strong>重现步骤：</strong></p><ol><li>…</li></ol><p><strong>实际结果：</strong></p><p>…</p><p><strong>期望结果：</strong></p><p>…</p>",
  "assigneeId": 5486,
  "reporterId": 4510,
  "moduleIds": [81],
  "versionTreeId": 571,
  "customFields": [
    {"id": 670, "value": [3621]},
    {"id": 671, "value": [3627]},
    {"id": 673, "value": [4136]},
    {"id": 675, "value": [4137]},
    {"id": 702, "value": [3842]},
    {"id": 788, "value": [5486]}
  ]
}
```

注意：
- 自定义字段**必须**走 `customFields` 数组，**禁止**顶层写 `custom_142_*`（会报复现概率为空）。
- `customFields` 每项 `value` 始终是数组；`moduleIds` / `versionTreeId` / 人员 ID 用整数。
- 若 `moduleIds` 无法通过 metadata 解析，先 `gpm-cli issue get <相似Bug>` 参考同模块单的模块 ID。
- Windows 下用 Python `json.dumps` + subprocess 调用 create，避免 PowerShell 破坏 JSON。
- 有附件时加 `--attach path/to/file.png`（图片自动嵌入描述）。

### 第五步：报告结果

创建成功后展示：

```
已创建 #17297 【sdk】tap 登录后无法充值 · 缺陷 · 待处理 · 刘俊杰
```

失败时读出 CLI 错误，说明缺哪个字段，不要盲重试。

---

## 口语触发映射（缺陷）

| 用户说 | 动作 |
|--------|------|
| 建个 Bug / 提缺陷 / 报 bug | 走缺陷建单流程 |
| 帮我建单（未说明类型） | 追问类型；若上下文明确是 bug 则走缺陷流程 |
| 挂到 #17167 下 | 创建时加 `parentIssueId` |
| 我的待办 | `gpm-cli mine` |
| 看 #17296 详情 | `gpm-cli issue get 17296` |

---

## 后续扩展（尚未实现，勿混用）

以下类型将在后续模板中补充，当前**不要**用缺陷默认值硬建：

- 后端子任务 / 前端子任务 / 策划子任务
- 需求开发 / 史诗
- 测试子任务 / 联调子任务

---

## 维护

项目字段或分支变更后，更新 [templates/bug-create.md](templates/bug-create.md) 中的枚举快照：

```bash
gpm-cli metadata fields 缺陷
gpm-cli version-tree list
```
