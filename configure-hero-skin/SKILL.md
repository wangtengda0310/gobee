---
name: configure-hero-skin
description: |
  配置名将杀游戏的"武将皮肤"道具，将飞书 Wiki 中"武将皮肤"表的新皮肤信息自动写入 Item.xlsx、HeroSkinItem_英雄皮肤.xlsx、HeroSkinSpine_英雄皮肤Spine.xlsx、ItemHeroSkin_武将皮肤展示表.xlsx。
  当用户说"配置XXX皮肤"、"新增皮肤道具"、"加一个皮肤"时触发。
  支持从飞书知识库的"武将皮肤"页签提取皮肤名称、武将名、拼音、等级，并自动关联 HeroLines 台词表、Hero.xlsx 武将表完成全链路配置。
---

# 配置武将皮肤 (Configure Hero Skin)

引用 `skills/common-rules.md` 中的通用铁律、检查点和使用记录规则。

从飞书 Wiki 的「武将皮肤」表中提取信息，自动写入 4 个 Excel 配置文件。

## 前置条件

- 本地 Excel 目录: `/Users/zt-3803045/.openclaw/workspace/名将杀配置/`
- 飞书 Wiki: https://ztgame.feishu.cn/wiki/MHZLwRAunieeUfkjm2ecNqrCnfe
- Python + openpyxl: 使用 `/Users/zt-3803045/.openclaw/skills/aicconfig/scripts/aicconfig/venv/bin/python`
- openpyxl 通过 `copy(src.font/border/fill/alignment/number_format)` 保留单元格全部样式
- 需要 `lark-cli` (bot 身份) 或 feishu 插件 (需 OAuth) 读取 Wiki

## 工作流程

### Step 1 — 解析用户需求

从自然语言中提取皮肤名称。例如:
- "配置驰猎惊澜皮肤" → 皮肤名: `驰猎惊澜`
- "新增赵云的皮肤" → 武将名: `赵云`, 皮肤名需从 Wiki 获取

### Step 2 — 读取飞书 Wiki 武将皮肤表

使用 `lark-cli` 通过 OpenAPI 查找 Wiki 节点:

```bash
# 获取 wiki 节点信息
lark-cli api GET "/open-apis/wiki/v2/spaces/get_node" --params '{"token":"MHZLwRAunieeUfkjm2ecNqrCnfe"}'

# 从返回的 obj_token 读取表格
lark-cli sheets +info --url "https://ztgame.feishu.cn/sheets/<obj_token>"

# 读取武将皮肤 sheet
lark-cli sheets +read --url "https://ztgame.feishu.cn/sheets/<obj_token>" --sheet-id "<sheet_id>"
```

如果 bot 无权限（permission denied），改用 feishu 插件（需用户 OAuth 授权）。

在「武将皮肤」sheet 中，找到「皮肤」列包含 Step 1 皮肤名的行，提取:
- **皮肤名称** — 如 "驰猎惊澜"
- **武将名** — 如 "孙权"
- **命名拼音** — 如 "chiliejinglan"
- **等级** — "精良" / "卓越" / "至臻"

### Step 3 — 执行配置脚本

```bash
/Users/zt-3803045/.openclaw/skills/aicconfig/scripts/aicconfig/venv/bin/python \
  /Users/zt-3803045/.openclaw/skills/configure-hero-skin/scripts/configure_hero_skin.py \
  "/Users/zt-3803045/.openclaw/workspace/名将杀配置/" \
  "<皮肤名称>" \
  "<武将名>" \
  "<皮肤拼音>" \
  "<等级>"
```

**脚本自动完成以下操作:**

#### 3.1 Hero.xlsx → 获取武将信息
- 查找武将名，获得 HeroID 和 HeroPinyin (PascalCase)
- 计算 HeroItemId = `int("10" + str(HeroID))`

#### 3.2 Item.xlsx → 配置皮肤道具
- 查找该武将的 HeroSkin 类型最大 ID
- 新 SkinID = max + 1 (如 10202004)
- 插入新行，复制上行样式，修改:
  - 道具id = 新 SkinID
  - 道具名称 = 皮肤名称
  - 道具品质 = 精良=2 / 卓越=3 / 至臻=4
  - 道具参数(ItemParam) = HeroItemId
  - 图标 icon = `UI/Images/Item/ui_s1_daoju_wujiang_{heroPinyinLower}_{skinPinyin}.png`
  - 道具描述 = `{武将名}的武将皮肤，获得后可以在游戏中改变武将{武将名}的卡牌形象`

#### 3.3 HeroSkinItem_英雄皮肤.xlsx → 配置皮肤详情
- 插入新行，配置:
  - 皮肤道具id = 新 SkinID
  - 英雄Id = HeroID
  - 皮肤类型 = SkinOtherSkin
  - 品质类型(RailyType) = 精良=1 / 卓越=2 / 至臻=3
  - 皮肤名称 / 获取途径 / 皮肤拼音
  - 台词配音 / 原画绘制人 / 皮肤收藏所属 = "待配置"
  - 是否开放 = 0

#### 3.4 HeroLines_武将台词表.xlsx → 提取台词
- 根据皮肤名称查找台词行
- 按标签分类: 登场 / 击杀 / 阵亡 / 自选 / 其他技能
- 将台词 ID 写入 HeroSkinItem 对应的列（文本格式）
- 台词格式规则:
  - col11(武将台词) = 所有非自选台词 ID（排序后逗号分隔）
  - col12(登场) / col13(击杀) / col14(阵亡) / col15(自选) 各自填写

#### 3.5 HeroSkinSpine_英雄皮肤Spine.xlsx → Spine 动画配置
- 新行插入，替换 HeroPascalCase 到音频配置:
  - SpineAnimAudio: `{2;Skill_Character_{HeroPascal}_Show},...`
  - KillAudio: `Skill_Character_{HeroPascal}_KillAttack`
  - SpineAnimBattleAudio: `{2;InGame_Skill_Character_{HeroPascal}_Show},...`

#### 3.6 ItemHeroSkin_武将皮肤展示表.xlsx → 展示配置
- 新行插入，路径中替换 hero pinyin 为对应武将拼音（小写）
- 坐标和音效留空不配置

## 等级映射对照

| 飞书等级 | Item.xlsx 品质值 | HeroSkinItem RailyType |
|----------|-----------------|----------------------|
| 精良     | 2               | 1                    |
| 卓越     | 3               | 2                    |
| 至臻     | 4               | 3                    |

## ID 命名规则

| ID 类型 | 格式 | 示例 (孙权 10202) |
|---------|------|-------------------|
| HeroID | 武将编号 | 10202 |
| HeroItemId | `10` + HeroID | 1010202 |
| SkinID | HeroID + 皮肤序号(3位) | 10202004 |
| SkinPinYin | `{heroPinyinLower}_{skinPinyin}` | sunquan_chiliejinglan |

## 涉及文件

```
名将杀配置/
├── Hero.xlsx                       ← 武将表 (武将ID/名称/拼音)
├── Item.xlsx                       ← 道具表 (皮肤道具ID/名称/品质/图标)
├── HeroSkinItem_英雄皮肤.xlsx      ← 皮肤详情 (台词/配音/原画等)
├── HeroLines_武将台词表.xlsx        ← 台词表 (台词文本/类型/标签)
├── HeroSkinSpine_英雄皮肤Spine.xlsx ← Spine动画 (动画音效配置)
└── ItemHeroSkin_武将皮肤展示表.xlsx ← 展示 (立绘路径/结算图)
```

## 注意事项

- 所有 Excel 操作通过 openpyxl + copy() 保留样式
- 台词列必须设为文本格式 (`number_format='@'`)，避免 Excel 把 ID 当成数字
- 如果 Wiki 中找不到皮肤或武将，公开提示用户确认
- ItemHeroSkin 路径中的 hero pinyin 必须小写
- HeroSkinSpine 音频配置中的 hero 名必须 PascalCase