---
name: configure-hero-item
description: |
  配置名将杀游戏"新武将道具"的完整流程：从 Hero.xlsx 和飞书 Wiki 获取武将信息，自动写入 Item.xlsx（武将道具+3个皮肤道具）、HeroSkinSpine_英雄皮肤Spine.xlsx（动态原画Spine）、HeroSkinItem_英雄皮肤.xlsx（皮肤详情+台词）、HeroUI_武将表现表.xlsx（武将表现配置）。
  当用户说"配置XXX道具"、"新增XXX武将道具"、"配置XXX武将"时触发，其中XXX为武将名称（如"王元姬"）。
  支持从飞书知识库获取武将品质信息，并自动关联 HeroLines 武将台词表完成全链路配置。
---

# 配置武将道具 (Configure Hero Item)

引用 `skills/common-rules.md` 中的通用铁律、检查点和使用记录规则。

配置一个新武将道具的完整 Excel 配置流程，涉及 5 个 Excel 文件共 9 处修改。

## 前置条件

- 本地 Excel 目录: `/Users/zt-3803045/.openclaw/workspace/名将杀配置/`
- 飞书 Wiki（武将品质）: `https://ztgame.feishu.cn/wiki/XwMDwfskviuqL1kCtDQc7hEEnxe`
- Python + openpyxl: `/Users/zt-3803045/.openclaw/skills/aicconfig/scripts/aicconfig/venv/bin/python`
- openpyxl 通过 `copy()` 保留单元格全部样式（字体/边框/填充/对齐/数字格式）
- 需要 `lark-cli` (bot 身份) 或 feishu 插件读取 Wiki

## 工作流程

### Step 1 — 解析用户需求

从自然语言中提取武将名称。例如:
- "配置王元姬道具" → 武将名: `王元姬`
- "新增王元姬武将道具" → 武将名: `王元姬`
- "配置神吕布道具" → 武将名: `神吕布`

### Step 2 — 读取 Hero.xlsx 获取武将信息

使用 `/Users/zt-3803045/.openclaw/skills/aicconfig/scripts/aicconfig/venv/bin/python` 读取 Hero.xlsx：

```python
import openpyxl
wb = openpyxl.load_workbook('/Users/zt-3803045/.openclaw/workspace/名将杀配置/Hero.xlsx')
ws = wb['武将表|Hero']
# 查找武将名所在行，提取:
# Col1 (武将ID) → hero_id, 如 12103
# Col3 (C列)   → hero_pinyin (PascalCase), 如 "WangyYuanJi"
# Col22        → expansion_pack, 如 "HeroExpansionPack_QinSaoLiuHe"
# Col2         → hero_name 确认
```

**提取字段:**
- **hero_id** — Col1 的值，如 `12103`
- **hero_pinyin** — Col3 的值 (PascalCase)，如 `WangyYuanJi`
- **hero_pinyin_lower** — hero_pinyin 全小写，如 `wangyyanji`
- **hero_item_id** — `int("10" + str(hero_id))`，如 `1012103`
- **expansion_pack** — Col22 的值，如 `HeroExpansionPack_QinSaoLiuHe`

### Step 3 — 获取武将品质

从飞书 Wiki 读取武将品质信息。使用 `lark-cli` 或 feishu 插件：

```bash
# 获取 wiki 节点信息
lark-cli api GET "/open-apis/wiki/v2/spaces/get_node" --params '{"token":"XwMDwfskviuqL1kCtDQc7hEEnxe"}'

# 读取表格
lark-cli sheets +read --url "https://ztgame.feishu.cn/sheets/<obj_token>"
```

如果 bot 无权限，改用 feishu 插件（需用户 OAuth）。

**品质映射:**
| 飞书品质 | 道具品质值 | 分解道具 | 合成消耗 |
|----------|-----------|----------|----------|
| 传说     | 4         | {1000025;400} | {1000025;1600} |
| 史诗     | 3         | {1000025;100} | {1000025;400} |
| 稀有     | 2         | {1000025;20}  | {1000025;80} |
| 普通     | 1         | {1000025;5}   | {1000025;20} |

### Step 4 — 执行配置脚本

```bash
/Users/zt-3803045/.openclaw/skills/aicconfig/scripts/aicconfig/venv/bin/python \
  /Users/zt-3803045/.openclaw/skills/configure-hero-item/scripts/configure_hero_item.py \
  "/Users/zt-3803045/.openclaw/workspace/名将杀配置/" \
  "<武将名>" \
  "<品质>"
```

参数说明:
- `base_dir` — 配置目录
- `hero_name` — 武将名称
- `quality` — 品质 (传说/史诗/稀有/普通)

**脚本自动完成以下所有操作:**

#### 4.1 Hero.xlsx → 获取武将信息
- 查找武将名，获得 HeroID、HeroPinyin (PascalCase)、ExpansionPack
- 计算 HeroItemId = `int("10" + str(HeroID))`

#### 4.2 Item.xlsx → 配置武将道具（#武将道具 区域）
- 在 `#武将道具` 区域中，按道具 ID 顺序找到插入位置（比新道具 ID 小的最大一行下方）
- 复制参考行，在其下方插入新行
- 修改内容:

| 列 | 字段 | 修改规则 |
|----|------|---------|
| A(1) | 道具id | `10` + HeroID |
| B(2) | 道具名称 | 武将名 |
| D(4) | 道具品质 | 传说=4, 史诗=3, 稀有=2, 普通=1 |
| L(12) | 道具参数 | HeroID |
| N(14) | 分解成为的道具 | `{1000025;400/100/20/5}` |
| O(15) | 专用信物道具ID | `11` + HeroID |
| V(22) | 图标icon | `UI/Images/Item/ui_s1_daoju_wujiang_{pinyin_lower}.png` |
| AB(28) | 道具描述 | `{武将名}的武将卡牌，在所有游戏模式中使用武将{武将名}` |
| AD(30) | 是否可以合成 | 1 |
| AE(31) | 合成消耗的道具 | `{1000025;1600/400/80/20}` |

其他列保持参考行模板值不变。

#### 4.3 Item.xlsx → 配置 3 个皮肤道具（#武将皮肤道具 区域）
在 `#武将皮肤道具` 区域中，按皮肤道具 ID 顺序插入 3 行：
- 第1行: 道具名称=`{武将名}-工笔白描`, 道具id=`{HeroID}001`, 图标=`.xiangao`
- 第2行: 道具名称=`{武将名}-动态原画`, 道具id=`{HeroID}002`, 图标无后缀
- 第3行: 道具名称=`{武将名}-兼工带写`, 道具id=`{HeroID}003`, 图标=`.gangman`

每行 道具品质=3, 道具参数=hero_item_id, 道具类型=HeroSkin

#### 4.4 Item.xlsx → 配置武将形象道具（#武将形象道具 区域）

**武将形象ID计算规则:**
- 武将形象ID = `102` + 武将ID后四位
- 示例: 武将ID=12103 → 后四位=2103 → 形象ID=1022103

**插入位置:** 在 `#武将形象道具` 区域中，按道具ID顺序找到插入位置（比新形象ID小的最大一行下方）。

**插入方式:** 从该区域内已有的一行武将形象道具**全量复制**值+样式到新行，然后只覆盖差异字段。

**覆盖字段:**

| 列 | 字段 | 修改规则 |
|----|------|---------|
| A(1) | 道具id | 武将形象ID（`102` + HeroID后四位） |
| B(2) | 道具名称 | `{武将名}-形象` |
| D(4) | 道具品质 | 固定为 3 |
| V(22) | 图标icon | 将模板行图标路径中的拼音部分，替换为 Step 2 中 Hero.xlsx C列的值（PascalCase 拼音） |
| AB(28) | 道具描述 | 将模板行描述中的武将名称替换为该武将名 |

其他所有列保持模板行值不变（全量复制后无需修改）。

#### 4.5 HeroSkinSpine_英雄皮肤Spine.xlsx → 动态原画 Spine
- 找到 `#动态皮肤` 标记行
- 在标记行**下方**（即 `#动态皮肤` 基础区域），按皮肤道具 ID 顺序查找插入位置
- 找到最后一个 col2 包含「动态皮肤」且 ID 比新 ID 小的行，在其下方插入新行
- 修改: 皮肤道具id={HeroID}002, C2=`{武将名}动态皮肤`, C3=True, C10=True, C12=True

#### 4.6 HeroSkinItem_英雄皮肤.xlsx → 3 行皮肤详情
找到 `#武将皮肤` 段落标记，在该区域中按英雄 ID 排序，找到比新 hero_id 小的最大英雄的最后一款皮肤行下方，插入 3 行：

**第1行（线稿皮肤）:**
| 列 | 修改规则 |
|----|---------|
| A(1) | 皮肤道具id = `{HeroID}001` |
| B(2) | 英雄Id = HeroID |
| C(3) | `{武将名}线稿皮肤` |
| D(4) | 品质类型 = 1 |
| E(5) | 皮肤类型 = `SkinLineSkin` |
| F(6) | 皮肤名称 = `工笔白描` |
| H(8) | 皮肤拼音 = `{pinyin_lower}_xiangao` |
| K~O(11-15) | 台词 ID（从 HeroLines 读取） |
| P(16) | 台词配音 = `待配置` |
| Q(17) | 原画绘制人 = `待配置` |

**第2行（动态皮肤）:**
- D(4) = 空（不配置）
- E(5) = `SkinNormalDynamicsSkin`
- F(6) = `动态原画`
- H(8) = `{pinyin_lower}`
- 其他同第1行

**第3行（港漫皮肤）:**

**新增配置要求 (HeroSkinItem G列和R列):**
- 3行皮肤的 获取途径(G列) 都配置为: "六合时邕解锁一定阶数获得"
- 第1行(线稿皮肤) 皮肤所属收藏(R列) 配置为: "ECollitionType_GBBM"
- 第2行(动态皮肤) 皮肤所属收藏(R列) 配置为: "" (空)
- 第3行(港漫皮肤) 皮肤所属收藏(R列) 配置为: "ECollitionType_JGDX"
- D(4) = 2
- E(5) = `SkinHKComicsSkin`
- F(6) = `兼工带写`
- H(8) = `{pinyin_lower}_gangman`
- 其他同第1行

#### 4.7 HeroLines_武将台词表.xlsx → 读取台词
- 搜索 `#{武将名}` 段落标记，提取该段落内所有台词行
- 按台词标签分类: 登场 / 击杀 / 阵亡 / 自选 / 其他技能
- 将台词 ID 填入 HeroSkinItem 的 C11~C15（格式为文本）
- 如果找不到台词，台词列留空，提示用户手动配置
- 台词 ID 按数值排序后用逗号连接

#### 4.8 HeroUI_武将表现表.xlsx → 武将表现配置
- 按武将 ID 顺序找到插入位置
- 复制参考行，在下方插入新行
- 修改:

| 列 | 修改规则 |
|----|---------|
| A(1) | 武将ID = HeroID |
| B(2) | 武将名称 = 武将名 |
| C(3) | 声音类型 = 空 |
| D(4) | 所属扩展包 = expansion_pack |
| E(5) | 武将获得描述 = 空 |
| F(6) | 简介（长）= 空 |
| G(7) | 简介（短）= 空 |
| H(8) | 考据 = 空 |
| I(9) | 评价 = 空 |
| J(10) | 是否新武将 = 0 |
| K(11) | 文案编辑人 = `椰椰` |
| L(12) | 技能设计人 = `待配置` |
| M(13) | 获得方式 = 空 |
| N(14) | 武将定位 = 空 |
| O(15) | 武将专属牌 = 空 |
| P(16) | 技能ID = 空 |
| Q(17) | 武将2v2胜率 = 空 |
| R(18) | 胜率显示优先级 = 空 |

## 品质映射表

| 飞书品质 | Item.xlsx 品质值 | 分解道具 | 合成消耗 |
|----------|-----------------|----------|----------|
| 传说     | 4               | `{1000025;400}` | `{1000025;1600}` |
| 史诗     | 3               | `{1000025;100}` | `{1000025;400}` |
| 稀有     | 2               | `{1000025;20}`  | `{1000025;80}` |
| 普通     | 1               | `{1000025;5}`   | `{1000025;20}` |

## ID 命名规则

| ID 类型 | 格式 | 示例 (王元姬, HeroID=12103) |
|---------|------|---------------------------|
| HeroID | 武将编号 | 12103 |
| HeroItemId | `10` + HeroID | 1012103 |
| 信物道具ID | `11` + HeroID | 1112103 |
| 工笔白描 SkinID | HeroID + `001` | 12103001 |
| 动态原画 SkinID | HeroID + `002` | 12103002 |
| 兼工带写 SkinID | HeroID + `003` | 12103003 |
| 武将形象ID | `102` + HeroID后四位 | 1022103 |

## 图标命名规则

| 皮肤类型 | Icon 路径格式 |
|---------|-------------|
| 武将道具 | `UI/Images/Item/ui_s1_daoju_wujiang_{pinyin_lower}.png` |
| 工笔白描 | `UI/Images/Item/ui_s1_daoju_wujiang_{pinyin_lower}_xiangao.png` |
| 动态原画 | `UI/Images/Item/ui_s1_daoju_wujiang_{pinyin_lower}.png` |
| 兼工带写 | `UI/Images/Item/ui_s1_daoju_wujiang_{pinyin_lower}_gangman.png` |
| 武将形象 | `UI/Images/Item/ui_s1_daoju_wujiang_{pinyin_lower}_xx.png`（将模板图标路径中的拼音部分替换为 Hero.xlsx C列 PascalCase 值） |

## 涉及文件

```
名将杀配置/
├── Hero.xlsx                       ← 武将表 (武将ID/名称/拼音/扩展包)
├── Item.xlsx                       ← 道具表 (武将道具 + 皮肤道具 + 形象道具)
├── HeroSkinItem_英雄皮肤.xlsx      ← 皮肤详情 (台词/配音/原画等)
├── HeroLines_武将台词表.xlsx        ← 台词表 (台词文本/类型/标签)
├── HeroSkinSpine_英雄皮肤Spine.xlsx ← Spine动画 (动态皮肤配置)
└── HeroUI_武将表现表.xlsx           ← 武将表现 (简介/考据/评价等)
```

## 注意事项

0. **⚠️ 暂不配置武器信物** — 默认**不创建**武器信物道具（Upgrade 类型，ID=`11`+HeroID）。脚本当前不包含武器信物的创建步骤，仅会在武将道具行的 `专用信物道具ID` 列中填入信物 ID 引用（值为 `11`+HeroID）。如用户明确要求配置武器信物，需另行处理。
1. **所有 Excel 操作通过 openpyxl + copy() 保留样式**
2. **台词列必须设为文本格式 (`number_format='@'`)**，避免 Excel 把 ID 当成数字
3. **插入行按 ID 顺序**，不要破坏已有的排序
4. **Hero.xlsx 的 C 列 (col3) 是 PascalCase 拼音**，图标路径使用全小写
5. **如果 HeroLines 中找不到武将台词**，台词列留空并提示用户手动添加
6. **模板复制**：如下方无参考行可复制（在区域末尾），则向上查找最近一行作为模板
7. **HeroSkinItem 中品质类型 (RailyType) 列**：线稿皮肤=1，动态皮肤=不配（空），港漫皮肤=2
8. **HeroSkinSpine 中动态原画皮肤插入在 `#动态皮肤` 标记下方基础区域**，按 skin ID 顺序排列，只配置基本字段：皮肤道具id、C2、C3(True)、C10(True)、C12(True)