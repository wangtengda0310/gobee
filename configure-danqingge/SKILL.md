---
name: configure-danqingge
description: |
  配置名将杀游戏的"丹青阁"活动，完整的端到端配置流程：从飞书 Wiki 读取丹青阁规划 →
  武将皮肤/边框/手牌皮肤名称查询 → Item.xlsx 道具 ID/图标读取 → Draw.xlsx 皮肤抽奖配置 →
  ShopGoods_商品表.xlsx 商品上架 → Drop.xlsx 掉落配置。
  当用户说"配置丹青阁"、"配置第N期丹青阁"、"丹青阁上新"时触发。
  适用于名将杀项目中新一期丹青阁活动的全链路配置。
---

# 配置丹青阁 (Configure Danqing Pavilion)

引用 `skills/common-rules.md` 中的通用铁律、检查点和使用记录规则。

从飞书 Wiki 读取丹青阁规划信息，自动完成 Draw、ShopGoods、Drop 三个 Excel 表的全链路配置。

## 前置条件

- 本地 Excel 目录: `/Users/zt-3803045/.openclaw/workspace/名将杀配置/`
- 飞书 Wiki (丹青阁规划): https://ztgame.feishu.cn/wiki/E4q1wOl58i4DmQkgpMvcre9hnee
- 飞书 Wiki (武将皮肤/边框/手牌皮肤): https://ztgame.feishu.cn/wiki/MHZLwRAunieeUfkjm2ecNqrCnfe
- Python + openpyxl: `/Users/zt-3803045/.openclaw/skills/aicconfig/scripts/aicconfig/venv/bin/python`
- 需要 `lark-cli` (bot 身份) 读取 Wiki

## 列定位规则（铁律）

⚠️ **所有 Excel 列的读写必须按表头名称动态定位，严禁硬编码列号（如 `col=23`、`col=34`）。**

每个 Excel 操作脚本的第一步，就是读表头建立 `{字段名 → 列号}` 映射：

```python
def build_header_map(ws, header_row=3):
    """读取表头行，返回 {字段名: 1-based列号} 的映射"""
    hdr = {}
    for cell in ws[header_row]:
        if cell.value is not None:
            hdr[str(cell.value).strip()] = cell.column
    return hdr

# 使用示例 — 默认读 Row 3（英文字段名），比中文表头更稳定
hdr = build_header_map(ws, header_row=3)

# 后续所有列操作都用 hdr['FieldName']，不要写死数字：
ws.cell(row=r, column=hdr['Icon']).value = icon_value        # ✅ 正确
ws.cell(row=r, column=23).value = icon_value                 # ❌ 禁止
```

为什么不可以用 `col=23`：Excel 一旦在左侧插入新列，所有列号全部偏移，配置写到错误列里。
用 `hdr['Icon']` 读取表头自动定位，无论列怎么移动，永远对准正确的列。

**默认读 Row 3（英文字段名）**，因为英文字段名比中文表头更稳定。如果 Row 3 为空，退回 Row 1（中文表头）。

下面每个 Step 中标注的字段名（如 `Icon`、`OnShelfTime`）就是 `hdr` 的 key。

## 工作流程

### Step 1 — 解析用户需求

从用户自然语言中提取期数。例如:
- "配置第21期丹青阁" → 期数 `21`
- "配置丹青阁-21期" → 期数 `21`
- "丹青阁上新" → 提示用户指定期数

### Step 2 — 读取丹青阁规划 Wiki

```bash
# 获取 wiki obj_token
lark-cli api GET "/open-apis/wiki/v2/spaces/get_node" --params '{"token":"E4q1wOl58i4DmQkgpMvcre9hnee"}'

# 读取 sheet 信息 (obj_token 从返回结果获取)
lark-cli sheets +info --spreadsheet-token "PdtjsYqxnh7iybtISNecV2dZneb"

# 读取"丹青阁规划（新）" sheet
lark-cli sheets +read --spreadsheet-token "PdtjsYqxnh7iybtISNecV2dZneb" --sheet-id "DNIiqw"
```

在「丹青阁规划（新）」sheet 中查找期数为 `丹青阁-N期` 的行，提取:
- **武将名** (B列, 主产-武将皮肤) — 如 "如姬"
- **萌将名** (C列, 副产-萌将形象) — 如 "【萌将】如姬"
- **副产物名称** (D列, 副产-边框/手牌皮肤) — 如 "如姬边框"
- **开始时间** (F列, Excel数字如 46156)
- **活动天数** (G列)
- **资产名称** (E列, 可选)

同时读取 `N-1期` 的行信息（用于 ShopGoods 配置上一期道具）:
- 上一期武将名、萌将名、副产物名称

**时间转换**: Excel 整数日期 (如 46156) → 实际日期:
```python
from datetime import datetime, timedelta
base = datetime(1899, 12, 30)
actual_date = base + timedelta(days=int(excel_int))
```

### Step 3 — 查询皮肤名称和拼音

打开 https://ztgame.feishu.cn/wiki/MHZLwRAunieeUfkjm2ecNqrCnfe 的「武将皮肤」sheet。

**注意: 同一个武将可能有多套皮肤，必须筛选丹青阁投放的那一套。**

判断规则：丹青阁皮肤一般是该武将的第2套皮肤（第1套是首充/通行证等），且品质为"至臻"。在「武将皮肤」sheet 中查找该武将的所有皮肤行，选择最新且非首充/非战令的那一套。

```bash
# 获取 wiki obj_token
lark-cli api GET "/open-apis/wiki/v2/spaces/get_node" --params '{"token":"MHZLwRAunieeUfkjm2ecNqrCnfe"}'

# 读取表格
lark-cli sheets +info --spreadsheet-token "<obj_token>"
lark-cli sheets +read --spreadsheet-token "<obj_token>" --sheet-id "<武将皮肤_sheet_id>"
```

从「武将皮肤」sheet 中提取:
- **皮肤名称** — 如 "愿栖棠枝"
- **命名拼音** — 全小写如 "yuanqitangzhi"
- **武将名** — 确认匹配
- **等级** — 确认是"至臻"

### Step 4 — 查询副产物信息

根据副产物名称判断类型:
- 名称包含 "边框" → 查找「边框套组」sheet
- 名称包含 "手牌皮肤" → 查找「手牌皮肤」sheet

同样在 https://ztgame.feishu.cn/wiki/MHZLwRAunieeUfkjm2ecNqrCnfe 中定位对应 sheet。

如果名称如 "如姬边框"，则在「边框套组」sheet 中以武将名搜索，提取:
- **边框名称** — 如 "愿栖棠枝" (和皮肤同名)
- **命名拼音**
- **道具 ID** (如果已经有)

对于手牌皮肤，同理在「手牌皮肤」sheet 中搜索。

**注意**: 边框/手牌皮肤通常和武将皮肤同名，命名拼音也相同。

⚠️ **如果找不到匹配条目**：不要自行推断名称或拼音。立即暂停配置，向用户汇报"在 [表名] 中未找到 [名称]，请确认该 [边框/手牌皮肤] 的准确名称是什么"，等待用户回复后再继续。

### Step 5 — 提取道具 ID 和图标 (Item.xlsx)

使用 Python openpyxl 读取 `Item.xlsx`。

⚠️ **第一步：建立列映射**（不要假设列号）:

```python
import openpyxl
wb = openpyxl.load_workbook('名将杀配置/Item.xlsx', data_only=True)
ws = wb['道具表|Item']

# 建立表头映射（build_header_map 定义见上方「列定位规则」章节）
hdr = build_header_map(ws)
# 现在通过 hdr['Id'] / hdr['Icon'] / hdr['Name'] / hdr['Type'] 等定位列
```

根据皮肤名称匹配道具行，获取:
- **道具id** — 列: `hdr['Id']`
- **图标 Icon** — 列: `hdr['Icon']`
- **道具品质/稀有度** — 列: `hdr['Rarity']` 或 `hdr['Quality']`
- **道具类型** — 列: `hdr['Type']`

对于萌将道具（MengJiang 类型）和边框/手牌皮肤道具，同样按名称匹配后通过 `hdr` 取值。

### Step 6 — 配置 Draw.xlsx (皮肤抽奖|DrawSkin)

打开 `Draw.xlsx` 的「皮肤抽奖|DrawSkin」sheet。

⚠️ **第一步：建表头映射**:

```python
wb = openpyxl.load_workbook('名将杀配置/Draw.xlsx', data_only=True)
ws = wb['皮肤抽奖|DrawSkin']
hdr = build_header_map(ws)  # 读取 Row 3 英文字段名
```

**插入位置:** 在 N-1 期（上一期）那行的**紧下方**插入一个新行。
- 例如：21 期在 Row 25，则 `ws.insert_rows(26, 1)` 在 Row 26 插入空白行
- ⚠️ 不要修改 N-1 期之前就存在的预设占位行（策划可能已预留了 Row 3X 的框架），插入后不要改动

**如果表中已存在第 N 期的预设占位行（如 Row 31 已有 Id=2022 的框架）:**
1. 先在 N-1 期下方插入新行并填入完整数据
2. 然后**删除**那个预设占位行
3. 严禁直接在预设占位行上修改——预设行位置不对（离 N-1 期隔了好几行），且字段可能不完整

**新行数据:** 复制 N-1 行的样式（font/fill/border/alignment），然后填入以下数据。所有未列出的列直接从 N-1 行复制。

| 字段名 (hdr key) | 说明 | 值 |
|------|------|------|
| `Id` | 活动ID | N-1行Id + 1 |
| `Name` | 活动名称 | 第N期皮肤名称 |
| `BigAwardItemId` | 大奖道具ID | 第N期皮肤道具id |
| `StartTime` | 开始时间 | `YYYY-MM-DD 05:00:00` |
| `EndTime` | 结束时间 | `YYYY-MM-DD 23:59:59` |
| `BgSpinePath` | 背景动画路径 | 复制 N-1 行，仅修改 `hero_{武将拼音}_{皮肤拼音}` 部分 |
| `SkinNameSpriteName` | 皮肤名称 | 第N期皮肤名称 |
| `byproduct` | 副产物 | `{皮肤道具id},{萌将道具id},{副产物道具id}` |
| `TitleIconName` | 标题图标名 | 复制 N-1 行，仅修改拼音部分为第N期皮肤拼音 |

**写入示例**（用 hdr 定位所有必覆盖字段）:
```python
# 1. 从 N-1 行全量复制值+样式到新行（openpyxl 逐列 copy）
# 2. 只覆盖以下差异字段：
ws.cell(row=new_row, column=hdr['Id']).value = prev_id + 1
ws.cell(row=new_row, column=hdr['Name']).value = skin_name
ws.cell(row=new_row, column=hdr['BigAwardItemId']).value = skin_item_id
ws.cell(row=new_row, column=hdr['StartTime']).value = f"{start_date} 05:00:00"
ws.cell(row=new_row, column=hdr['EndTime']).value = f"{end_date} 23:59:59"
ws.cell(row=new_row, column=hdr['BgSpinePath']).value = spine_path  # 修改拼音部分
ws.cell(row=new_row, column=hdr['SkinNameSpriteName']).value = skin_name
ws.cell(row=new_row, column=hdr['byproduct']).value = f"{skin_item_id},{meng_item_id},{byprod_item_id}"
ws.cell(row=new_row, column=hdr['TitleIconName']).value = title_icon  # 修改拼音部分
```

**BgSpinePath 修改示例:**
```
N-1行: hero_zhugeliang_kongchengheli_limit_skin_bg.prefab
新行:   hero_ruji_yuanqitangzhi_limit_skin_bg.prefab
```

**TitleIconName 修改示例:**
```
N-1行: ui_s1_danqingge_text_kongchengheli
新行:   ui_s1_danqingge_text_yuanqitangzhi
```

### Step 7 — 读取上一期丹青阁信息

从「丹青阁规划（新）」中读取 N-1 期 (如 20期) 的信息:
- 上一期武将名 (如 "诸葛亮")
- 上一期萌将名 (如 "【萌将】诸葛亮")
- 上一期副产物名称 (如 "诸葛亮手牌皮肤")

### Step 8 — 配置 ShopGoods_商品表.xlsx

打开 `ShopGoods_商品表.xlsx` 的「商品表|ShopGood」sheet。

⚠️ **第一步：建表头映射**:

```python
wb = openpyxl.load_workbook('名将杀配置/ShopGoods_商品表.xlsx', data_only=True)
ws = wb['商品表|ShopGood']
hdr = build_header_map(ws)  # 读取 Row 3 英文字段名
```

**插入位置**: `#丹青` (Row 204) 和 `#公会` (Row 330) 之间。在 `#公会` 的前一行下方依次插入。

**插入方式**: 从 `#丹青` 下方第一行数据（样式参考行，如 Row 205）**全量复制**值+样式到新行，然后只覆盖差异字段。这样 `CostId`、`ShopType` 等公共字段自动继承，无需显式赋值。

按照 **皮肤 → 萌将 → 副产物** 的顺序依次插入 3 行（逆序插入以保持最终顺序）。

**所有行的下架时间固定**: `2060-10-31 23:59:59`

**ShopGoods 字段映射（按 hdr 动态定位）:**

⚠️ **Item 字段格式铁律**: 必须是 **`{道具id;1}`** 字面量字符串，`{}` 花括号是字符串的一部分，不能省略！
Python f-string 正确写法: `f"{{{道具id变量};1}}"` （三重花括号：前两个转义为字面量 `{`，中间是变量，最后两个转义为字面量 `}`）
❌ 错误: `f"{item_id};1"` → 输出 `12345;1` (缺花括号)
✅ 正确: `f"{{{item_id};1}}"` → 输出 `{12345;1}`

| 字段名 (hdr key) | 说明 |
|------|------|
| `Id` | 商品id |
| `Name` | 商品名称 |
| `Desc` | 商品说明 |
| `Item` | 获得道具 (字面量: `{id};1`，花括号不可省略) |
| `CostId` | 购买消耗货币 |
| `ShopType` | 商店类型 |
| `Icon` | 商品图标 |
| `OnShelfTime` | 上架时间 |
| `OffShelfTime` | 下架时间 (固定 2060-10-31) |
| `IconInBuyWindow` | 商品在购买界面图标 |
| `RewardID` | 奖励预览窗口 (道具id) |
| `ItemDisplayTap` | 商品显示页签 |

**Tab 值:** 皮肤=0, 萌将=3, 边框=1, 手牌皮肤=2

#### 8a — 皮肤商品配置 (Tab=0)

| 字段 (hdr key) | 值 |
|----|-----|
| `Id` | 上一行Id + 1 |
| `Name` | 上一期皮肤名称 |
| `Desc` | `购买后获得{上一期皮肤所属武将名}的武将皮肤，获得后可以在游戏中改变武将{上一期皮肤所属武将名}的卡牌形象` |
| `Item` | `{上一期皮肤道具id};1` |
| `Icon` | 上一期皮肤图标 |
| `OnShelfTime` | 第N期开始时间 |
| `OffShelfTime` | `2060-10-31 23:59:59` |
| `IconInBuyWindow` | 上一期皮肤图标 |
| `RewardID` | 上一期皮肤道具id |
| `ItemDisplayTap` | 0 |

#### 8b — 萌将商品配置 (Tab=3)

| 字段 (hdr key) | 值 |
|----|-----|
| `Id` | 上一行Id + 1 |
| `Name` | `{上一期武将名}·萌将` |
| `Desc` | `购买后获得{上一期武将名}·萌将形象，可以在形象系统中使用，将自己的形象设置为{上一期武将名}·萌将形象` |
| `Item` | `{上一期萌将道具id};1` |
| `Icon` | 上一期萌将图标 |
| `OnShelfTime` | 第N期开始时间 |
| `OffShelfTime` | `2060-10-31 23:59:59` |
| `IconInBuyWindow` | 上一期萌将图标 |
| `RewardID` | 上一期萌将道具id |
| `ItemDisplayTap` | 3 |

#### 8c — 副产物商品配置

**Tab 值判断**: 边框=1, 手牌皮肤=2

**边框:**
- Desc: `购买后获得{名称}形象边框，个人形象边框道具，可以在形象系统中使用，将自己的形象边框设置为此边框形象`

**手牌皮肤:**
- Desc: `购买后获得{名称}手牌皮肤，可以在收藏-手牌中使用，改变游戏中的手牌样式`

其余字段配置同 8a/8b，`Item`、`Icon`、`IconInBuyWindow`、`RewardID` 均为副产物道具的值。

### Step 9 — 配置 Drop.xlsx (掉落道具表|DropItem)

打开 `Drop.xlsx` 的「掉落道具表|DropItem」sheet。

⚠️ **第一步：建表头映射**:

```python
wb = openpyxl.load_workbook('名将杀配置/Drop.xlsx', data_only=True)
ws = wb['掉落道具表|DropItem']
hdr = build_header_map(ws)  # 读取 Row 3 英文字段名
```

在以下 4 个掉落组中各新增一行。

---

#### ⚠️ 核心原则：先复制 N-1 行全部数据，再覆盖差异字段

**严禁**为 Deduplication、CheckExist、ExcludeExist、MustHave 等字段硬编码 0 或 None。每个新行必须：
1. 找到该掉落组内 N-1 期的行（插入位置的上方一行）
2. 用 `openpyxl` 的 `copy()` 复制该行所有单元格的值和样式到新行
3. 然后**只覆盖**以下需要变更的字段：Id、Name、Item、ValidDate、ExpireDate
4. 其余字段（Weight、WeightInc、Deduplication、CheckExist、ExcludeExist、MustHave、ReplaceGroup 等）**全部保持与 N-1 行一致**

**DropItem 字段映射（按 hdr 动态定位）:**

⚠️ **Item 字段格式铁律**（同上）: 必须是 **`{道具id;1}`** 字面量字符串。
Python f-string 正确写法: `f"{{{item_id};1}}"`
❌ 错误: `f"{item_id};1"` → `12345;1` | ✅ 正确: `f"{{{item_id};1}}"` → `{12345;1}`

| 字段名 (hdr key) | 说明 |
|------|------|
| `Id` | 掉落ID |
| `Name` | 掉落名称 |
| `DropGroup` | 掉落组ID |
| `Item` | 掉落道具 (字面量: `{id};1`，花括号不可省略) |
| `Weight` | 掉落权重 |
| `WeightInc` | 权重递增 |
| `Deduplication` | 是否去重 |
| `CheckExist` | 检查已有 |
| `ExcludeExist` | 排除已有 |
| `MustHave` | 必须拥有才会加入 |
| `ReplaceGroup` | 替代掉落组ID |
| `ValidDate` | 加入掉落时间 |
| `ExpireDate` | 移出掉落时间 |

---

#### 9a — 掉落组 90001 (当期皮肤)

**插入位置:** 组内 90901（固定奖励行）**上方**。实际就是 N-1 期皮肤行的下方，90901 的上方。

**覆盖字段:**
| 字段 (hdr key) | 值 |
|----|-----|
| `Id` | 上一行 Id + 1 |
| `Name` | 第N期皮肤名称 |
| `Item` | `{第N期皮肤道具id};1` |
| `ValidDate` | 第N期开始时间 |
| `ExpireDate` | 第N期结束时间 |

**其余所有字段**（`Weight`、`WeightInc`、`Deduplication`、`CheckExist`、`ExcludeExist`、`MustHave`、`ReplaceGroup` 等）**保持与 N-1 行完全一致**。

**写入示例**（先复制整行再覆盖差异字段）:
```python
# 1. 从 N-1 行全量复制值+样式到新行
# 2. 只覆盖差异字段：
ws.cell(row=new_row, column=hdr['Id']).value = prev_id + 1
ws.cell(row=new_row, column=hdr['Name']).value = skin_name
ws.cell(row=new_row, column=hdr['Item']).value = f"{skin_item_id};1"
ws.cell(row=new_row, column=hdr['ValidDate']).value = start_time_str
ws.cell(row=new_row, column=hdr['ExpireDate']).value = end_time_str
# DropGroup 等其余全部字段已从 N-1 行继承，无需再设
```

#### 9b — 掉落组 90002 (保底大奖)

**插入位置:** 90002 组内 N-1 期皮肤下方（组末尾）。

覆盖字段同 9a（皮肤名称、道具id、时间），`DropGroup` 为 90002。其余字段全部从 N-1 行继承。

#### 9c — 掉落组 90003 (萌将道具)

**插入位置:** 90003 组内 N-1 期萌将（如"如姬·萌将"）下方。

**覆盖字段:**
| 字段 (hdr key) | 值 |
|----|-----|
| `Id` | 上一行 Id + 1 |
| `Name` | 第N期萌将名（格式: `{武将名}·萌将`，如"曹仁·萌将"） |
| `Item` | `{第N期萌将道具id};1` |
| `ValidDate` | 第N期开始时间 |
| `ExpireDate` | 第N期结束时间 |

其余字段全部从 N-1 行继承。

#### 9d — 掉落组 90005 (副产物)

**插入位置:** 90005 组内 N-1 期副产物（如"棠花沁梦"）下方。

**覆盖字段:**
| 字段 (hdr key) | 值 |
|----|-----|
| `Id` | 上一行 Id + 1 |
| `Name` | 第N期副产物道具名称（如"炽虎衔炎"） |
| `Item` | `{第N期副产物道具id};1` |
| `ValidDate` | 第N期开始时间 |
| `ExpireDate` | 第N期结束时间 |

其余字段全部从 N-1 行继承。

---

## 执行流程总结

```
用户输入 "配置第N期丹青阁"
  │
  ├─ Step 1: 解析期数 N
  ├─ Step 2: 读取丹青阁规划 → 武将/萌将/副产物/时间
  ├─ Step 3: 查询武将皮肤表 → 皮肤名称/拼音
  ├─ Step 4: 查询边框/手牌皮肤表 → 副产物名称/拼音
  ├─ Step 5: 查询 Item.xlsx → 道具id/图标 (用 hdr 动态定位)
  ├─ Step 6: 配置 Draw.xlsx → DrawSkin 新行 (用 hdr 动态定位)
  ├─ Step 7: 读取 N-1 期信息
  ├─ Step 8: 配置 ShopGoods_商品表.xlsx → 3行 (用 hdr 动态定位)
  └─ Step 9: 配置 Drop.xlsx → 4行 (用 hdr 动态定位)
```

## 涉及文件

```
名将杀配置/
├── Item.xlsx                    ← 道具表 (读取道具id/图标)
├── Draw.xlsx                    ← 抽卡表 (Sheet: 皮肤抽奖|DrawSkin)
├── ShopGoods_商品表.xlsx         ← 商品表 (Sheet: 商品表|ShopGood)
└── Drop.xlsx                    ← 掉落表 (Sheet: 掉落道具表|DropItem)
```

## Wiki 数据源

- 丹青阁＆战令排期: `https://ztgame.feishu.cn/wiki/E4q1wOl58i4DmQkgpMvcre9hnee`
  - spreadsheet_token: `PdtjsYqxnh7iybtISNecV2dZneb`
  - 「丹青阁规划（新）」sheet_id: `DNIiqw`
- 武将皮肤/边框/手牌皮肤: `https://ztgame.feishu.cn/wiki/MHZLwRAunieeUfkjm2ecNqrCnfe`

## 时间计算

Excel 整数日期转实际日期:
```python
from datetime import datetime, timedelta
base = datetime(1899, 12, 30)
start_date = base + timedelta(days=int(excel_int))
end_date = start_date + timedelta(days=days - 1)
```

格式: `YYYY-MM-DD 05:00:00` (开始) / `YYYY-MM-DD 23:59:59` (结束)

## 注意事项

1. **备份优先** — 修改 Excel 前建议 git commit 或手动备份
2. **样式保留** — 所有 Excel 操作通过 openpyxl + `copy(src)` 保留单元格全部样式
3. **飞书操作降级链** — `lark-cli` 失败后才用 `feishu` 插件，feishu 插件失败则告知用户需要 OAuth 授权
4. **ID 连续性** — DrawSkin Id 在上一行基础上 +1，ShopGoods Id 同理
5. **时间一致性** — 所有表的开始/结束时间必须与丹青阁规划一致
6. **同名皮肤** — 边框/手牌皮肤通常与武将皮肤同名
7. **不要覆盖** — 插入新行而非覆盖已有数据
8. **上一期商品** — ShopGoods 配置的是 N-1 期（上一期）道具，但上架时间是第 N 期时间
9. **掉落组 90001/90002** — 配置的是第 N 期皮肤（当期），不是上一期
10. **掉落组 90003/90005** — 配置的是第 N 期萌将和副产物（当期）
11. ⚠️ **信息缺失必须暂停** — 在任意 Step 中发现关键信息缺失（如武器名、皮肤名、边框名、手牌皮肤名等无法在对应 Wiki/Excel 中找到匹配条目），**必须立即暂停配置，向用户询问**。严禁自行猜测或根据命名惯例推断。例如：丹青阁规划写的是"曹仁边框"，但边框表中找不到"曹仁边框"条目时，应停下来问用户该边框的准确名称是什么，而不是假定它和皮肤同名。
12. ⚠️ **Item.xlsx 新增道具后必须校验** — 通过脚本或手动在 Item.xlsx 创建新道具（如萌将/边框/手牌皮肤）后，必须校验所有文本字段中的武将名是否与道具所属武将一致。从其他行复制模板时，逐字段确认替换完成，严禁遗漏。
13. ⚠️ **Draw.xlsx 插入规则** — 必须在 N-1 期的紧下方插入新行。
    - 如果表中存在**2 行上一期数据**（实际数据行 + 预设占位行），**必须取第 1 行（顶部的实际数据行）作为插入参考**，忽略下方的预设占位行。
    - 如果表中存在第 N 期的预设占位行（策划提前预留的框架），插入新行后必须删除该占位行。
    - 严禁直接在预设占位行上修改——预设行的位置和字段都不正确。
14. ⚠️ **Drop.xlsx 核心原则：「复制 N-1 行全部字段，只覆盖差异」** — 严禁为 Deduplication、CheckExist、ExcludeExist、MustHave、ReplaceGroup 等字段硬编码 0 或 None。必须从 N-1 行（同组上一行）复制这些值。历史教训：硬编码 0 导致"多赔了很多 0"且替代掉落组 ID 丢失。
15. ⚠️ **列定位铁律：按表头名称动态定位，严禁硬编码列号** — 所有 Excel 读写操作必须先用 `build_header_map(ws)` 建立 `{字段名 → 列号}` 映射，然后通过 `hdr['FieldName']` 获取列号。严禁写 `col=23`、`col=34` 这类硬编码——Excel 左侧插入新列后所有数值都会偏移，导致配置写到错误的列。历史教训：ShopGoods 表 6 个字段的列号全部偏移，Icon 写到 IconInBuyWindow 的位置。
