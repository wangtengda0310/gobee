---
name: configure-hero-drop
description: |
  配置名将杀武将掉落时间。从飞书 Wiki "武将排班表" 读取武将信息 → Item.xlsx 查找道具 ID →
  Drop.xlsx 掉落道具表插入新掉落配置行。
  当用户说"配置武将掉落时间"、"配置XX的掉落时间"、"武将掉落"时触发。
  适用于名将杀项目中新增武将的掉落时间配置。
---

# 配置武将掉落时间

引用 `skills/common-rules.md` 中的通用铁律、检查点和使用记录规则。

从飞书 Wiki «武将排班表» 读取武将信息，自动完成 Drop.xlsx 掉落配置。

## 前置条件

- 本地 Excel 目录: `/Users/zt-3803045/.openclaw/workspace/名将杀配置/`
- 飞书 Wiki (武将排班表): https://ztgame.feishu.cn/wiki/XwMDwfskviuqL1kCtDQc7hEEnxe
- Python + openpyxl: `/Users/zt-3803045/.openclaw/skills/aicconfig/scripts/aicconfig/venv/bin/python3`
- 脚本: `/Users/zt-3803045/.openclaw/skills/configure-hero-drop/scripts/configure_hero_drop.py`

## 工作流程

### Step 1 — 解析用户需求

从用户自然语言中提取武将名。例如:
- "配置郑国的掉落时间" → 武将名 `郑国`
- "配置曹操掉落" → 武将名 `曹操`
- 未指定武将名 → 提示用户指定

### Step 2 — 读取飞书 Wiki «武将排班表»

```bash
# Wiki 解析（已完成）
# wiki_token: XwMDwfskviuqL1kCtDQc7hEEnxe
# spreadsheet_token: HODRsCYqVhuLGttKm61csaDAngf
# sheet_id: d50c43 (武将发布排班表)
```

在「武将发布排班表」sheet 中搜索武将名，获取该行以下信息:

| 列 | 字段 | 说明 |
|----|------|------|
| B(2) | 武将 | 武将名称 |
| E(5) | 发布时间 | Excel 序列日期 (整数) |
| M(13) | 品质 | 传说/史诗/稀有/普通 |
| N(14) | 投放方式 | 大将军/战令/空 |

**时间转换**: Excel 整数日期 (如 46170) → 实际日期:
```python
from datetime import datetime, timedelta
base = datetime(1899, 12, 30)
actual_date = base + timedelta(days=int(excel_int))
```

### Step 3 — 查找武将道具 ID (Item.xlsx)

用 Python openpyxl 读取 `Item.xlsx` 的「道具表|Item」sheet。

在 `#武将道具` 区域中搜索:
- 道具名称 (B列) = 武将名
- 道具类型 (C列) = `Hero`

获取: **道具id** (A列)

```python
import openpyxl
wb = openpyxl.load_workbook('Item.xlsx', data_only=True)
ws = wb['道具表|Item']
# 从 Row 5 开始遍历，匹配 Name=武将名 且 Type=Hero 的行
```

### Step 4 — 计算掉落时间

| 投放方式 | 掉落时间 | 格式 |
|----------|---------|------|
| 大将军 | 发布时间 + 1 个自然月 | `YYYY-MM-DD 00:00:00` |
| 战令 | 发布时间 + 4 个自然月 | `YYYY-MM-DD 00:00:00` |
| 空 (未投放) | 发布时间 | `YYYY-MM-DD 05:00:00` |

### Step 5 — 配置 Drop.xlsx (掉落道具表|DropItem)

打开 `Drop.xlsx` 的「掉落道具表|DropItem」sheet。

#### 品质 → 掉落组ID 映射

| 品质 | 掉落组ID |
|------|---------|
| 传说 | 10004 |
| 史诗 | 10003 |
| 稀有 | 10002 |
| 普通 | 10001 |

#### 插入流程

1. 定位 `#武将掉落` 标记行
2. 在该区域中找到当前最大掉落ID
3. 在最后一行数据下方插入新行
4. **复制上一行全部样式** (font/fill/border/alignment/number_format)
5. 覆盖差异字段

#### DropItem 列映射

| Col | 字段 | 新值 | 来源 |
|-----|------|------|------|
| A(1) | Id | 当前最大掉落ID + 1 | — |
| B(2) | Name | 武将名 | Step 2 |
| C(3) | DropGroup | 品质对应的掉落组ID | Step 2 品质 |
| D(4) | Item | `{道具id;1}` | Step 3 |
| E(5) | Weight | 复制上一行 | — |
| F(6) | WeightInc | 复制上一行 | — |
| G(7) | Deduplication | 复制上一行 | — |
| H(8) | CheckExist | 复制上一行 | — |
| I(9) | ExcludeExist | 复制上一行 | — |
| J(10) | MustHave | 复制上一行 | — |
| K(11) | ReplaceGroup | 复制上一行 | — |
| L(12) | ValidDate | 计算的掉落时间 | Step 4 |
| M(13) | ExpireDate | `2054-12-31 00:00:00` | 固定值 |

#### ⚠️ 核心原则：复制上一行全部，只覆盖差异

**严禁**为 G/H/I/J/K 等列硬编码值。必须：
1. 复制上一行所有单元格的值和样式
2. 只覆盖 Col A/B/C/D/L 五个差异列
3. 其余列（E/F/G/H/I/J/K 等）全部保持与上一行一致

### Step 6 — 验证结果

配置完成后，读取新插入的行验证:
- Id 正确递增
- Name/DropGroup/Item 正确
- ValidDate 格式和时间正确
- 样式与上一行一致

## 执行命令

```bash
/Users/zt-3803045/.openclaw/skills/aicconfig/scripts/aicconfig/venv/bin/python3 \
  /Users/zt-3803045/.openclaw/skills/configure-hero-drop/scripts/configure_hero_drop.py <武将名>
```

## 涉及文件

```
名将杀配置/
├── Item.xlsx    ← 道具表 (Sheet: 道具表|Item, 查询武将道具ID)
└── Drop.xlsx    ← 掉落表 (Sheet: 掉落道具表|DropItem, #武将掉落区域)
```

## Wiki 数据源

- 武将排班表: `https://ztgame.feishu.cn/wiki/XwMDwfskviuqL1kCtDQc7hEEnxe`
  - spreadsheet_token: `HODRsCYqVhuLGttKm61csaDAngf`
  - 「武将发布排班表」sheet_id: `d50c43`

## 注意事项

1. **样式保留** — 所有 Excel 操作通过 openpyxl + `copy(src)` 保留单元格全部样式
2. **飞书操作降级链** — 优先 `lark-cli` (bot 身份)，失败后告知用户需要授权
3. **ID 连续性** — 掉落ID 在 `#武将掉落` 区域内递增
4. **插入位置** — 必须在 `#武将掉落` 区域最后一行下方插入，不能覆盖标记行
5. **自然月计算** — 注意月份溢出处理（如 1月31日 + 1个月 → 2月28日）
6. **日期间基准** — Excel 日期基准为 1899-12-30 (1900 日期系统)
7. **复制先行** — 插入新行后先全量复制上一行值+样式，再只覆盖差异字段
8. **备份提醒** — 配置前备份 Drop.xlsx（脚本不自动备份）
