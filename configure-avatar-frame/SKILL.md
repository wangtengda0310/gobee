---
name: configure-avatar-frame
description: |
  配置名将杀游戏的"形象边框"（Avatar Frame）道具，将飞书 Wiki 中"边框套组"表的新形象边框信息自动写入 Item.xlsx 和 ItemFrame_边框道具表.xlsx。
  当用户说"配置XXX的形象边框"、"新增形象边框"、"加一个边框"、或者提到角色名+形象边框时触发。
  适用于需要从飞书知识库的"边框套组"页签中提取边框名称、拼音、等级，并同步到本地 Excel 配置表的场景。
---

# 配置形象边框 (Configure Avatar Frame)

引用 `skills/common-rules.md` 中的通用铁律、检查点和使用记录规则。

从飞书 Wiki 的「边框套组」表中提取信息，自动写入 `Item.xlsx` 和 `ItemFrame_边框道具表.xlsx`。

## 前置条件

- 本地 Excel 目录: `/Users/zt-3803045/.openclaw/workspace/名将杀配置/`
- 飞书 Wiki: https://ztgame.feishu.cn/wiki/MHZLwRAunieeUfkjm2ecNqrCnfe?sheet=wQs8oN
- 需要 `lark-cli` (bot 身份) 读取 Wiki 表格
- 需要 Python + `openpyxl` 操作 Excel（使用 aicconfig venv: `/Users/zt-3803045/.openclaw/skills/aicconfig/scripts/aicconfig/venv/bin/python`）
- openpyxl 保留单元格全部样式（字体/边框/填充/对齐/数字格式）

## 工作流程

### Step 1 — 解析用户需求

从用户自然语言中提取角色/主题名称。例如:

- "配置如姬的形象边框" → 角色名: `如姬`
- "新增嬴政的形象边框" → 角色名: `嬴政`
- "加一个刘邦的边框" → 角色名: `刘邦`

### Step 2 — 读取飞书 Wiki 边框套组表

使用 `lark-cli` 读取 Wiki 中的电子表格，定位到「边框套组」工作表。

```bash
# 获取表格信息
lark-cli sheets +info --url "https://ztgame.feishu.cn/wiki/MHZLwRAunieeUfkjm2ecNqrCnfe?sheet=wQs8oN"

# 读取边框套组 sheet 的数据（先用 info 获取 sheet_id，再用 read）
lark-cli sheets +read --url "https://ztgame.feishu.cn/wiki/MHZLwRAunieeUfkjm2ecNqrCnfe?sheet=wQs8oN" --sheet-id "<边框套组sheet_id>"
```

如果 lark-cli 无法直接读取 Wiki 内嵌的 sheet（可能返回 "Unsupported document type 'sheet'"），改用飞书插件（需要用户 OAuth 授权）:

```
feishu_sheet action=info url="https://ztgame.feishu.cn/wiki/MHZLwRAunieeUfkjm2ecNqrCnfe?sheet=wQs8oN"
feishu_sheet action=read url="..." range="<边框套组sheet_id>"
```

### Step 3 — 查找匹配行

在「边框套组」表中，找到 **「主题/玩法」** 列包含 Step 1 角色名的行。

从该行提取以下字段:
- **形象边框名称** — 如 "千嶂万象"
- **命名拼音** — 如 "qianzhangwanxiang"（全小写无分隔）
- **等级** — "精良" / "卓越" / "至臻"

### Step 4 — 执行配置

使用 `scripts/configure_frame.py` 脚本（openpyxl 保留样式）完成 Excel 操作:

```bash
/Users/zt-3803045/.openclaw/skills/aicconfig/scripts/aicconfig/venv/bin/python \
  /Users/zt-3803045/.openclaw/skills/configure-avatar-frame/scripts/configure_frame.py \
  "/Users/zt-3803045/.openclaw/workspace/名将杀配置/" \
  "<形象边框名称>" \
  "<命名拼音>" \
  "<等级>"
```

**脚本自动完成:**

1. 在 `Item.xlsx` 的「道具表|Item」中:
   - 找到道具类型为 `Frame` 且 ID 最大的行
   - 复制该行，在其下方插入新行
   - 修改: 道具 ID +1、道具名称 = 边框名称、道具品质 = 等级数值 (精良=2/卓越=3/至臻=4)
   - 修改 icon 路径中的拼音部分 (col 21)

2. 在 `ItemFrame_边框道具表.xlsx` 的「边框表|FrameItem」中:
   - 找到 ID 最大的行
   - 复制该行，在其下方插入新行
   - 修改: 道具 ID = Item 表新 ID、边框名 = 边框名称
   - 替换 C~G 列中所有旧拼音为新拼音

### Step 5 — 确认结果

向用户汇报配置结果:
- 新道具 ID
- 边框名称
- 等级
- 各个文件更新的列内容

## 等级映射

| 飞书等级 | Item.xlsx 品质值 |
|----------|-----------------|
| 精良     | 2               |
| 卓越     | 3               |
| 至臻     | 4               |

## 文件结构

```
名将杀配置/
├── Item.xlsx                    ← 道具表 (Sheet: 道具表|Item)
├── ItemFrame_边框道具表.xlsx    ← 边框表 (Sheet: 边框表|FrameItem)
└── ...
```

### Item.xlsx 关键列

| 列 | 字段 | 说明 |
|----|------|------|
| A (0) | 道具id | 新 ID = 最大 Frame ID + 1 |
| B (1) | 道具名称 | 形象边框名称 |
| C (2) | 道具类型 | Frame |
| D (3) | 道具品质 | 2/3/4 |
| V (21) | Icon 路径 | `UI/Images/Item/ui_s1_daoju_frame_character_{拼音}.png` |

### ItemFrame_边框道具表.xlsx 关键列

| 列 | 字段 | 说明 |
|----|------|------|
| A (0) | 道具id | 引用 Item 表新 ID |
| B (1) | 边框名 | 形象边框名称 |
| C (2) | 标识 | `ui_s1_frame_character_{拼音}_01` |
| D (3) | Image 路径 | `UI/Images/Global/ui_s1_frame_character_{拼音}_01.png` |
| E (4) | 头像边框 | `ui_s1_frame_head_{拼音}_01` |
| F (5) | Prefab 路径 | 含拼音和 PascalCase 目录名 |
| G (6) | Head Prefab 路径 | 含拼音和 PascalCase 目录名 |

## 注意事项

- 脚本执行前会自动备份逻辑（通过 deep copy），不会删除原有数据
- 拼音替换是基于上一行旧拼音的全局字符串替换，确保 C~G 列全部更新
- 如果 Wiki 中的角色名在边框套组中找不到，应向用户确认
- 数据行从第 5 行开始（前 4 行为表头: 类型提示 / 字段名 / client-server 标记）
