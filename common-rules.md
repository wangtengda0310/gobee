# 名将杀配置 Skill 通用规则

本文档是所有 `configure-*` skill 共享的铁律和检查点。每个 skill 的 SKILL.md 和 Python 脚本都必须遵守。

---

## 🔴 铁律（不可违反）

### 规则 1：插入行 = 全量复制 + 覆盖差异

**插入新行必须：先用 `copy_row_style` 全量复制上一行（值+样式），再只覆盖差异字段。**

```python
# 三步走
ws.insert_rows(target_row)                           # 1. 插入行
copy_row_style(ws, src_row, target_row, max_col)     # 2. 全量复制值+样式
ws.cell(row=target_row, column=X).value = new_val    # 3. 只覆盖差异字段
```

❌ 严禁逐个手动指定字段 — 会漏配（历史：麒麟配置遗漏 AK 列「是否做拥有检查」）
❌ 严禁 `insert_rows()` 后不复制样式 — 插入行默认无样式

### 规则 2：列定位 = 动态表头映射

**所有 Excel 列的读写必须按表头名称动态定位，严禁硬编码列号。**

```python
def build_header_map(ws, header_row=3):
    hdr = {}
    for cell in ws[header_row]:
        if cell.value is not None:
            hdr[str(cell.value).strip()] = cell.column
    return hdr

hdr = build_header_map(ws)
# 后续用 hdr['FieldName'] 定位，不要写死数字
ws.cell(row=r, column=hdr['Icon']).value = val  # ✅
ws.cell(row=r, column=23).value = val            # ❌
```

❌ 严禁 `col=23`、`col=34` 等硬编码 — Excel 插入新列后全部偏移
📌 历史：ShopGoods 表 6 个字段的列号全部 +1 偏移，Icon 写到 IconInBuyWindow

### 规则 3：复制 N-1 行全部，只覆盖差异

**Drop.xlsx、Draw.xlsx 等有 N-1 期参考行的场景，必须从 N-1 行全量复制所有字段，只覆盖需要变更的列。**

❌ 严禁为 Deduplication/CheckExist/ExcludeExist/MustHave/ReplaceGroup 等字段硬编码 0 或 None
📌 历史：丹青阁 Drop 硬编码 0 导致"多赔了很多 0"且替代掉落组 ID 丢失

### 规则 4：Skill 更新 = MD + 脚本双同步

**更新 skill 时，SKILL.md 和对应的 Python 脚本必须同步更新，缺一不可。**

---

## 🟡 操作规范（必须遵守）

### 规范 1：飞书操作降级链

```
lark-cli (bot 身份) → 失败 → feishu 插件 (用户 OAuth) → 失败 → 告知用户需要授权
```

### 规范 2：信息缺失必须暂停

任意步骤中发现关键信息缺失（名称、ID 等无法在 Wiki/Excel 中找到匹配条目），**必须立即暂停，向用户询问**。严禁自行猜测或根据命名惯例推断。

### 规范 3：备份优先

修改 Excel 前应 git commit 或手动备份。Python 脚本大改前也应备份。

### 规范 4：台词列为文本格式

所有涉及台词 ID 的列必须设为文本格式 (`number_format='@'`)，避免 Excel 把 ID 当成数字处理。

### 规范 5：插入行按 ID 顺序

所有区域内的插入行必须保持 ID 升序，找到比新 ID 小的最大一行下方插入。

### 规范 6：模板复制

如下方无参考行可复制（在区域末尾），则向上查找最近一行作为模板。

### 规范 7：ID 连续性

区域内 ID 必须连续递增，不可跳跃或重复。

---

## 🛠 公共工具函数

所有 Python 脚本应使用以下公共函数：

```python
from copy import copy

def copy_cell_style(src, dst):
    """复制单元格的值+全部样式"""
    dst.value = src.value
    if src.has_style:
        dst.font = copy(src.font); dst.border = copy(src.border)
        dst.fill = copy(src.fill); dst.number_format = copy(src.number_format)
        dst.protection = copy(src.protection); dst.alignment = copy(src.alignment)

def copy_row_style(ws, src_row, dst_row, max_col):
    """从 src_row 全量复制值+样式到 dst_row"""
    for col in range(1, max_col + 1):
        copy_cell_style(ws.cell(row=src_row, column=col),
                        ws.cell(row=dst_row, column=col))

def build_header_map(ws, header_row=3):
    """读取表头行，返回 {字段名: 1-based列号} 的映射"""
    hdr = {}
    for cell in ws[header_row]:
        if cell.value is not None:
            hdr[str(cell.value).strip()] = cell.column
    if not hdr:  # 退至 Row 2
        for cell in ws[header_row - 1]:
            if cell.value is not None:
                hdr[str(cell.value).strip()] = cell.column
    return hdr

def insert_row_below(ws, ref_row, max_col):
    """在 ref_row 下方插入新行，自动复制样式"""
    new_row = ref_row + 1
    safe_insert_rows(ws, new_row, 1)
    copy_row_style(ws, ref_row, new_row, max_col)
    return new_row
```

### Excel 日期转换

```python
from datetime import datetime, timedelta
base = datetime(1899, 12, 30)                        # Excel 1900 日期系统基准
actual_date = base + timedelta(days=int(excel_int))
# 格式: YYYY-MM-DD 05:00:00 (开始) / YYYY-MM-DD 23:59:59 (结束)
```

---

## ✅ 配置完成后检查点

每次配置完成后，逐项确认：

| # | 检查项 | 说明 |
|---|--------|------|
| 1 | **ID 唯一性** | 新增 ID 在区域内只出现一次（预设占位行已删除） |
| 2 | **ID 连续性** | 区域内 ID 递增无跳跃 |
| 3 | **样式继承** | 新增行的字体/边框/填充/对齐与参考行一致 |
| 4 | **列定位** | 所有字段写在正确列（hdr 定位，非硬编码） |
| 5 | **时间格式** | YYYY-MM-DD HH:MM:SS，开始 05:00:00，结束 23:59:59 |
| 6 | **下架时间** | ShopGoods 固定 2060-10-31 23:59:59，Drop.ExpireDate 与活动结束一致 |
| 7 | **Deduplication 等** | Drop 的非差异字段全部从 N-1 行继承，未硬编码 0 |
| 8 | **拼音一致性** | 各处使用的拼音与 Wiki/Hero.xlsx 一致（大小写/格式） |
| 9 | **台词文本格式** | 台词 ID 列已设 `number_format='@'` |
| 10 | **武将名替换** | 描述/Iconfolder 中所有引用点已完成武将名替换 |
| 11 | **复制后覆盖完整性** | 从模板复制的字段中，差异字段已全部覆盖（无遗漏） |

---

## 📊 配置完成后自动记录使用数据

每次完成配置任务后，**必须调用 `skill-usage-tracker`**（`skills/skill-usage-tracker/SKILL.md`），向多维表格写入一条记录：

- **配置任务详情**：从用户原始请求提取核心信息（≤80 字）
- **配置时间**：`YYYY-MM-DD HH:MM` 格式（北京时间）
- **人员**：从当前 SenderId 查飞书姓名
- **任务类型**：当前 skill 对应的任务类型（如「丹青阁配置」「武将皮肤配置」）

记录规则：
- 一条用户消息触发多个 skill（如"配置孙权皮肤+边框"），每个 skill 分别记一条
- 如果是更新 skill 本身（改 SKILL.md / 脚本），不记录
- 写入失败不阻塞主流程

详细步骤见 `skills/skill-usage-tracker/SKILL.md`。

---

## 📋 各 Skill 特殊规则索引

| Skill | 特殊注意事项 |
|-------|-------------|
| configure-hero-item | 信物道具默认不创建；3 皮肤收藏类型不同；形象 ID=102+后四位 |
| configure-hero-skin | 皮肤拼音格式 `{hero}_{skin}`；台词 / 配音 / 原画默认「待配置」 |
| configure-hero-drop | 掉落时间 = 发布时间 + 月数（大将军+1月/战令+4月） |
| configure-hero-companion | 三表联动 (Item形象+道具+Pet)；技能ID 需在 Skill.xlsx 查找 |
| configure-danqingge | DrawSkin 预设占位行必须删除；ShopGoods 配置的是 N-1 期道具 |
| configure-avatar-frame | ItemFrame C-G 列拼音全局替换 |
