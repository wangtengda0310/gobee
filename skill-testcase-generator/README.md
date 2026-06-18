# 名将杀：技能测试用例生成器

输入一个技能名称，自动从飞书源表收集背景知识、拆解技能定义、生成结构化功能测试用例，并将 Excel + Markdown 双份产物上传到飞书云空间。**仅适用于名将杀项目。**

这是一个 Claude Code 自定义 Skill：完整流程定义在 [`SKILL.md`](SKILL.md)，由 Claude 在对话中识别触发条件后自动执行。

## 触发方式

消息中**同时满足两点**即启动全流程：

1. 包含技能名称（如：武圣义绝、天不负我、二虎竞食 等）
2. 包含意图关键词：`测试` / `用例` / `测` / `生成`

口语化表达也会触发，例如：「生成武圣义绝的技能测试用例」「测一下天不负我」「天不负我技能测试」「给 XXX 出一份测例」「XXX 用例」。

## 数据来源

| 项目 | 配置 |
|---|---|
| 源文档 | 飞书在线表格 [`所有技能效果`](https://ztgame.feishu.cn/sheets/Z9kFs9JWdhqxQ5tt0I9csmytnVg)（token `Z9kFs9JWdhqxQ5tt0I9csmytnVg`） |
| 关键 sheet | 技能（`iwM7X5`）/ 牌（`a109ea`）/ 基本术语（`cFEl74`）/ 所有机制（`n2eNub`）/ 测试用例生成点模块 |
| 产出统计 | 飞书 Bitable [测试用例生成数据统计](https://ztgame.feishu.cn/base/AyOlbNrnzavuCzscvYPchXwtnUU)（表 `tblnlTnPgX12x9gP`） |

## 工作流概览

```
阶段零  初始化任务专属子目录（<技能名>_<时间戳>，隔离并行任务）
  │
阶段一  数据采集
  ├─ 1.0 快照校验/刷新（前置硬门控：比对 revision，过期则重建本地快照）
  ├─ 1.1 定位技能，取全部条目（合并单元格解析武将/势力归属）
  ├─ 1.2 反向查表解析文案引用（牌 / 术语 / 机制 / 关联技能 / 授予技能）
  ├─ 1.3 拆解技能定义 → 原子项 → 归入 5 个二级模块
  └─ 1.4 关联技能 + 组合交互分析（去重成集合，作完整性校验基准）
  │
阶段二  落盘 overview.json（条目 + 各 section）
  │
阶段三  生成 testcases.json（8 列用例）
  ├─ 3.1 技能定义 5 子模块（列结构相对源表反排）
  ├─ 3.2 其余一级模块（封禁/被重置/被删除/被转移/刻写复制/多个在场/重新初始化/断线重连）
  ├─ 3.3 关联技能交互（辅助模块）
  ├─ 3.4 机制测试点（逐条预设生成，追加在表末）
  └─ 3.5 完整性校验（唯一闸门，差集非空则按名称补齐后重校验）
  │
阶段四  输出交付
  ├─ 4.1 生成 Excel + Markdown（脚本统一格式）
  ├─ 4.2 上传飞书 Drive
  ├─ 4.3–4.5 授权（固定人员+发起人「可管理」/ 群聊「可读」/ 组织内链接分享）
  ├─ 4.6 返回文件夹链接
  ├─ 4.7 写入 Bitable 产出统计（非关键收尾，失败不中断；捕获 RECORD_ID）
  └─ 4.8 触发反馈收集（交给 user-feedback skill 发反馈卡片、回写记录）
```

## 输出产物

每次运行在任务子目录下生成：

- `overview.json` — 阶段一采集结果（技能拆解映射、概念枚举、核心术语、机制测试点、关联技能、背景知识）
- `testcases.json` — 8 列测试用例（编号 / 一级模块 / 二级模块 / 拆解项 / 标题 / 前置条件 / 步骤 / 预期结果）
- `<技能名>_测试用例.xlsx` — Sheet 1 测试概览 + Sheet 2 测试用例（自动筛选、冻结首行）
- `<技能名>_测试用例.md` — 同内容 Markdown

最终通过飞书云空间文件夹链接交付，xlsx 与 md 均开启组织内链接分享。

## 目录结构

```
skill-testcase-generator/
├── SKILL.md                      # 技能定义（Claude Code 加载的主文件）
├── README.md                     # 本文件
├── scripts/
│   ├── generate_outputs.py       # overview/testcases JSON → Excel + Markdown
│   └── append_stats.py           # 产出统计写入飞书 Bitable
└── references/
    ├── snapshot-validation.md         # 1.0 快照校验链
    ├── sheet-structure.md             # 源表结构、合并单元格解析、牌类型层级
    ├── skill-decomposition.md         # 技能拆解、5 模块归类、发动阶段 16 节点
    ├── associated-skill-interaction.md# 关联技能交互类型与组合分析
    ├── overview-format.md             # overview.json 字段与 section 规范
    ├── testcase-format.md             # 用例 8 列结构与预期结果撰写规范
    ├── test-point-modules.md          # 测试用例生成点一级模块清单（快照）
    └── test-point-knowledge.md        # 测试点背景知识（快照）
```

> `references/` 下的 `sheet-structure` / `test-point-modules` / `test-point-knowledge` 是源表快照，由阶段 1.0 的 revision 门控按需刷新——**改 revision 号前必须从 live 数据全量重建内容**。

## 运行环境

- **Python 环境**: `/usr/bin/python3`（已含 openpyxl）
- **飞书 CLI**：`lark-cli`（失败后降级到 `feishu` 插件，再失败提示 OAuth 授权）
- **依赖 skill**：`user-feedback`（阶段 4.8 调用，负责发反馈卡片并回写 Bitable 记录）
- **本地输出根目录**：`/Users/zt-3803045/.openclaw/workspace/名将杀技能测试用例/`

## 关键约定

1. **合并单元格先解析** — 技能归属与条目范围全靠合并范围判定，不靠 null 推断。
2. **照搬原文** — 技能文案、术语定义、牌效果一律用源文档原文，禁止简化；推断结论标「待验证」。
3. **JSON 中文引号用「」** — 写入 JSON 的中文文本内引号必须用「」，否则解析失败。
4. **快照刷新不是改号** — revision 变化后四份快照须从 live 数据全量重建，再改号。
5. **统计写表非关键** — 4.7 失败只告警、不中断主交付。
