---
name: configure-hero-companion
description: |
  配置名将杀伙伴（灵宠）道具。从飞书 Wiki 读取伙伴信息（品质/技能/属性/时间） →
  Item.xlsx 道具表新增 Partner 类型道具 → Item.xlsx 形象表新增形象 →
  Pet_灵宠表.xlsx 新增灵宠条目。
  当用户说"配置伙伴道具"、"配置XX伙伴"、"新增伙伴XX"、"配置灵宠"时触发。
  适用于名将杀项目中新伙伴/灵宠的配置。
---

# 配置伙伴道具 (Configure Companion Item)

引用 `skills/common-rules.md` 中的通用铁律、检查点和使用记录规则。

从飞书 Wiki 读取伙伴信息，自动完成 Item 道具表、形象表和 Pet 灵宠表的配置。

## 前置条件

- 本地 Excel 目录: `/Users/zt-3803045/.openclaw/workspace/名将杀配置/`
- 飞书 Wiki (伙伴养成和技能数值): https://ztgame.feishu.cn/wiki/FiciwL6KAi5sLfkUMVIcLu07nd7
  - spreadsheet_token: `YjDestOAchLLFSt80kicLcnynmb`
  - 「每个伙伴维度」sheet_id: `q6Dunq`
- 飞书 Wiki (伙伴规划): https://ztgame.feishu.cn/wiki/E4q1wOl58i4DmQkgpMvcre9hnee?sheet=4pfFqd
  - spreadsheet_token: `PdtjsYqxnh7iybtISNecV2dZneb` (同丹青阁&战令排期)
  - 「伙伴规划」sheet_id: `4pfFqd`
- Python + openpyxl: `/Users/zt-3803045/.openclaw/skills/aicconfig/scripts/aicconfig/venv/bin/python3`
- 脚本: `/Users/zt-3803045/.openclaw/skills/configure-hero-companion/scripts/configure_hero_companion.py`

## 工作流程

### Step 1 — 解析用户需求

从自然语言中提取伙伴名称。例如:
- "配置麒麟伙伴道具" → 伙伴名 `麒麟`
- "新增伙伴白泽" → 伙伴名 `白泽`
- "配置灵宠九尾" → 伙伴名 `九尾`

### Step 2a — 读取伙伴规划 Wiki（品质/拼音/投放时间）

从「伙伴规划」sheet (sheet_id: `4pfFqd`) 读取伙伴排期数据。

```bash
lark-cli sheets +find --spreadsheet-token "PdtjsYqxnh7iybtISNecV2dZneb" \
  --sheet-id "4pfFqd" --find "<伙伴名>"
lark-cli sheets +read --spreadsheet-token "PdtjsYqxnh7iybtISNecV2dZneb" \
  --sheet-id "4pfFqd" --range "4pfFqd!A<row>:F<row>"
```

**列映射:**
| Col | 字段 | 说明 | 用途 |
|-----|------|------|------|
| B(2) | 伙伴名称 | 如 "麒麟" | Name |
| C(3) | 拼音 | 全小写，如 "qilin" | 路径/图标名 |
| D(4) | 品质 | 传说/史诗/稀有/普通 | Rarity |
| E(5) | 投放时间 | Excel 序列日期 (整数) | StartTime |
| F(6) | 投放方式 | 如 "结缘亭" | — |

**时间转换:** Excel 整数日期 → 实际日期:
```python
from datetime import datetime, timedelta
base = datetime(1899, 12, 30)
actual_date = base + timedelta(days=int(excel_int))
# 46188 → 2026-06-15 (麒麟)
```

### Step 2b — 读取伙伴养成 Wiki（属性/技能名称）

```bash
# 搜索伙伴所在行
lark-cli sheets +find --spreadsheet-token "YjDestOAchLLFSt80kicLcnynmb" \
  --sheet-id "q6Dunq" --find "<伙伴名>"

# 读取该行完整数据
lark-cli sheets +read --spreadsheet-token "YjDestOAchLLFSt80kicLcnynmb" \
  --sheet-id "q6Dunq" --range "q6Dunq!A<row>:N<row>"
```

**列映射 (1-based):**
| Col | 字段 | 说明 | 用途 |
|-----|------|------|------|
| A(1) | 伙伴名 | 如 "麒麟" | Name |
| B(2) | 品质 | 传说/史诗/稀有/普通 | Rarity |
| C(3) | 洗练消耗 | 数值 | — |
| D(4) | 速度 | 属性值 | BattleAttrWeight |
| E(5) | 初始手牌 | 属性值 | BattleAttrWeight |
| F(6) | 摸牌 | 属性值 | BattleAttrWeight |
| G(7) | 手牌上限 | 属性值 | BattleAttrWeight |
| H(8) | 体力上限 | 属性值 | BattleAttrWeight |
| I(9) | 杀的伤害 | 属性值 | BattleAttrWeight |
| J(10) | 战法伤害 | 属性值 | BattleAttrWeight |
| K(11) | 出杀次数 | 属性值 | BattleAttrWeight |
| L(12) | 技能1名称 | 如 "天上麒麟" | Skill 名称 |
| M(13) | 技能1描述 | 完整技能描述 | — |
| N(14) | 技能2名称 | 如有第二个技能 | Skill 名称 |

### Step 2c — 查找技能ID (Skill.xlsx)

从 Wiki 获取技能名称后，在本地 `Skill.xlsx` 的「技能表|Skill」sheet 中查找 `#伙伴技能` 区域，通过技能名称匹配获取技能 ID。

```python
import openpyxl
wb = openpyxl.load_workbook('Skill.xlsx', data_only=True)
ws = wb['技能表|Skill']
# #伙伴技能 区域在 Row 751 开始（Row 750 是标记行）
# Col 1(A) = 技能ID, Col 2(B) = 技能名称
# 通过技能名称匹配获取对应的技能ID
```

**技能ID范围:** 9501-9527+
**区域位置:** Row 751 ~ Row 777（当前），新增技能累加在末尾

**示例**: 麒麟 技能1「天上麒麟」→ 9521, 技能2「麒麟吐书」→ 9525

⚠️ **如果技能名称未找到** — 说明该伙伴技能在 Skill.xlsx 中还未创建，需先配置技能。暂停并告知用户。

**伙伴规划 Wiki (TODO):**
| 字段 | 说明 | 用途 |
|------|------|------|
| 上线时间 | 格式 YYYY-MM-DD | StartTime |
| 拼音 | 全小写 | 路径名 |
| Prefab路径 | 预制体路径 | Pet 表 |
| 头像Icon路径 | UI路径 | 图标配置 |

如果用户未提供伙伴规划 Wiki，拼音和路径可参考已有伙伴的命名规则推断，但上线时间和确切路径应让用户确认。

**品质映射:**
| Wiki 品质 | 道具品质(Rarity) |
|-----------|-----------------|
| 传说     | 4               |
| 史诗     | 3               |
| 稀有     | 2               |
| 普通     | 1               |

### Step 3 — 配置 Item.xlsx 道具表|Item

在 `道具表|Item` sheet 中 Parther 类型道具区域新增一行。

**插入位置:** Partner 类型道具区域最后一行下方。

**插入方式:**
1. `ws.insert_rows(insert_row)` 插入空白行
2. 复制上一行所有样式（font/fill/border/alignment/number_format）
3. 覆盖差异列

**覆盖字段:**
| Col | 字段 | 值 | 来源 |
|-----|------|-----|------|
| A(1) | Id | 当前最大 Partner ID + 1 (范围 50000xx) | — |
| B(2) | Name | 伙伴名称 | Step 2 |
| C(3) | Type | `Partner` | 固定 |
| D(4) | Rarity | 品质对应数值 | Step 2 品质 |
| V(22) | Icon | 图标路径 | Step 2 |
| X(24) | IsGetHint | 1 | 固定 |
| Y(25) | IsHide | 1 | 固定 |
| AB(28) | Des | 道具描述 | Step 2 |
| AJ(36) | DisplayIcon | 展览图标 | Step 2 |
| AK(37) | HaveCheck | 1 | 固定 |

其余所有列保持与上一行一致（复制）。

### Step 4 — 配置 Item.xlsx 形象表|ImageItem

在 `形象表|ImageItem` sheet 中新增伙伴形象行。

**插入位置:** 形象表最后一行下方。

**覆盖字段:**
| Col | 字段 | 值 | 来源 |
|-----|------|-----|------|
| A(1) | Id | 与道具 ID 相同 | Step 3 |
| B(2) | Name | 伙伴名称 | Step 2 |
| E(5) | 小头像 | 小头像 UI 路径 | Step 2 |

### Step 5 — 配置 Pet_灵宠表.xlsx

在 `灵宠|Pet` sheet 中新增灵宠条目。

**插入位置:** Pet 表最后一行下方。

**覆盖字段:**
| Col | 字段 | 值 | 来源 |
|-----|------|-----|------|
| A(1) | Id | 与道具 ID 相同 | Step 3 |
| B(2) | Name | 伙伴名称 | Step 2 |
| D(4) | PrefabPath | Prefab 路径 | Step 2 |
| E(5) | StartTime | 上线时间 | Step 2 |
| F(6) | Skill | `{槽位;技能ID}` 格式，逗号分隔 | Step 2 |
| G(7) | BattleAttrWeight | KVPair[] 格式 | Step 2 |
| K(11) | SquareHeadIcon | 方型头像路径 | Step 2 |
| L(12) | HeadIcon | 头像路径 | Step 2 |
| M(13) | Silhouette | 剪影路径 | Step 2 |
| V(22) | InfoTextID | `待配置` | 固定 |

## 执行命令

```bash
/Users/zt-3803045/.openclaw/skills/aicconfig/scripts/aicconfig/venv/bin/python3 \
  /Users/zt-3803045/.openclaw/skills/configure-hero-companion/scripts/configure_hero_companion.py <伙伴名>
```

## 涉及文件

```
名将杀配置/
├── Skill.xlsx        ← 技能表 (Sheet: 技能表|Skill, #伙伴技能 区域查技能ID)
├── Item.xlsx         ← 道具表 (Sheet: 道具表|Item + 形象表|ImageItem)
└── Pet_灵宠表.xlsx   ← 灵宠表 (Sheet: 灵宠|Pet)
```

## Wiki 数据源

- 伙伴规划: https://ztgame.feishu.cn/wiki/E4q1wOl58i4DmQkgpMvcre9hnee?sheet=4pfFqd
  - spreadsheet_token: `PdtjsYqxnh7iybtISNecV2dZneb` (同丹青阁&战令排期)
  - 「伙伴规划」sheet_id: `4pfFqd`
- 伙伴养成和技能数值: https://ztgame.feishu.cn/wiki/FiciwL6KAi5sLfkUMVIcLu07nd7
  - spreadsheet_token: `YjDestOAchLLFSt80kicLcnynmb`
  - 「每个伙伴维度」sheet_id: `q6Dunq`

## 注意事项

1. **样式保留** — 所有 Excel 操作通过 openpyxl + `copy(src)` 保留单元格全部样式
2. **先复制再覆盖** — 插入新行后必须从上一行全量复制值+样式，再只覆盖差异字段
3. **飞书操作降级链** — 优先 `lark-cli` (bot 身份)，失败后告知用户需要授权
4. **ID 连续性** — Partner 道具 ID 在 50000xx 范围内递增
5. **三表联动** — Item (道具表+形象表) 和 Pet 表必须同步配置，ID 保持一致
6. **技能格式** — Pet.Skill 格式为 `{槽位;技能ID}`，多个技能用逗号分隔
7. **路径检查** — 图标/Prefab 路径需与美术确认实际资源路径
8. **备份提醒** — 配置前备份 Item.xlsx 和 Pet_灵宠表.xlsx
