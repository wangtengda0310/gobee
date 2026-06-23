---
name: configure-hero-mengjiang
description: |
  配置名将杀游戏的"萌将道具"（武将 Q 版形象），根据武将名自动从 Hero.xlsx 获取拼音，
  在 Item.xlsx 的 #武将形象道具 区域新增萌将道具行。
  当用户说"配置XXX萌将"、"新增XXX萌将"时触发，其中XXX为武将名称（如"孙权"）。
  基础规则引用 common-rules.md。
---

# 配置萌将道具 (Configure Hero Mengjiang)

配置一个新武将的萌将道具（只涉及 Item.xlsx），引用 `skills/common-rules.md` 中的通用铁律和检查点。

## 前置条件

- 本地 Excel 目录: `/Users/zt-3803045/.openclaw/workspace/名将杀配置/`
- Python + openpyxl: `/Users/zt-3803045/.openclaw/skills/aicconfig/scripts/aicconfig/venv/bin/python`
- 通用规则: `/Users/zt-3803045/.openclaw/skills/common-rules.md`

## 工作流程

### Step 1 — 解析用户需求

从自然语言中提取武将名。例如:
- "配置孙权萌将" → 武将名: `孙权`
- "新增曹操萌将" → 武将名: `曹操`

### Step 2 — 读取 Hero.xlsx 获取武将拼音

打开 `Hero.xlsx` 的「武将表|Hero」sheet，搜索武将名，获取 **C 列的值（PascalCase 拼音）**。

```python
import openpyxl
wb = openpyxl.load_workbook('名将杀配置/Hero.xlsx')
ws = wb['武将表|Hero']
for row in range(5, ws.max_row + 1):
    if ws.cell(row=row, column=2).value and str(ws.cell(row=row, column=2).value).strip() == hero_name:
        hero_pinyin = ws.cell(row=row, column=3).value  # PascalCase
        break
```

### Step 3 — 执行配置脚本

```bash
/Users/zt-3803045/.openclaw/skills/aicconfig/scripts/aicconfig/venv/bin/python \
  /Users/zt-3803045/.openclaw/skills/configure-hero-mengjiang/scripts/configure_hero_mengjiang.py \
  "/Users/zt-3803045/.openclaw/workspace/名将杀配置/" \
  "<武将名>"
```

**脚本自动完成以下操作:**

#### 3.1 读取 Hero.xlsx → 获取 PascalCase 拼音
- 查找武将名，获取 C 列的值（PascalCase 拼音）
- 计算 `pinyin_lower`（全小写）用于 icon 路径

#### 3.2 定位插入位置
- 打开 `Item.xlsx` 的「道具表|Item」sheet
- 找到 `#武将形象道具` 区域（Row 252）
- 在该区域内搜索所有名称包含 "·萌将" 的行
- 找到**最后一个萌将道具**所在行作为插入参考行

#### 3.3 插入新行（引用 common-rules.md 规则 1）
- 在最后一个萌将道具行下方插入新行
- 从参考行**全量复制**值+样式（`copy_row_style`）
- 只覆盖以下差异字段:

| 列 | 字段 | 修改规则 |
|----|------|---------|
| A(1) | 道具id | 当前区域内最大萌将 ID + 1 |
| B(2) | 道具名称 | `{武将名}·萌将` |
| D(4) | 道具品质(Rarity) | 固定为 4 |
| N(14) | 分解成为的道具(ItemResolve) | `{1000002;1000}` |
| V(22) | 图标icon | 将模板行图标路径中的拼音部分，替换为 Step 2 中 Hero.xlsx 的 `pinyin_lower` |
| AB(28) | 道具描述 | `{武将名}·萌将形象，可以在形象系统中使用，将自己的形象设置为{武将名}·萌将形象` |

其他所有列（Type/IsCanBuy/IsUseable/IsGetHint/IsHide/HaveCheck 等）保持模板行值不变。

### Step 4 — 验证结果（引用 common-rules.md 检查点）

配置完成后按 `common-rules.md` 的 11 项检查点逐项确认：
- ID 唯一性、连续性
- 样式继承
- 拼音一致性（icon 路径中 pinyin 与 Hero.xlsx C 列一致）
- 武将名替换完整性（道具名称和描述中）

## 数据规律总结

| 字段 | 值/格式 |
|------|---------|
| 道具类型(Type) | `Image` |
| 道具品质(Rarity) | 始终为 `4` |
| 分解道具(ItemResolve) | `{1000002;1000}` |
| 图标格式 | `UI/Images/Item/ui_s1_daoju_pifuxingxiang_{pinyin_lower}_Qban.png` |
| 描述格式 | `{武将名}·萌将形象，可以在形象系统中使用，将自己的形象设置为{武将名}·萌将形象` |
| ID 编号 | 萌将 ID 递增，无固定前缀模式 |

## 涉及文件

```
名将杀配置/
├── Hero.xlsx    ← 武将表 (Sheet: 武将表|Hero, Col C = PascalCase 拼音)
└── Item.xlsx    ← 道具表 (Sheet: 道具表|Item, #武将形象道具 区域)
```

## 注意事项

1. 本 skill 只涉及 Item.xlsx 单表操作，是所有 configure-* skill 中最简单的
2. **通用规则引用** — 插入行、样式复制、ID 排序等遵循 `common-rules.md` 铁律 1-4
3. **拼音来源** — icon 路径使用 Hero.xlsx C 列的 `pinyin_lower`（全小写），而非 PascalCase
4. **萌将识别** — 通过名称中包含 "·萌将" 来识别萌将道具行
5. **ID 递增** — 新 ID = 当前最大萌将 ID + 1，不是整个区域的最大 ID
