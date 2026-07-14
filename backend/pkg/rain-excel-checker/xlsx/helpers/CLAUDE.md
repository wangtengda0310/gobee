# helpers 包 — Excel 校验规则的通用辅助工具集

提供校验规则的通用辅助工具，包括列操作、参数解析、正则表达式、拼音转换、路径规范化、武将数据处理等功能。

## 文件结构

```
helpers/
├── utils.go              # 通用工具函数（参数解析、Sheet查找、数据结束检测）
├── hero_rule_helper.go   # 武将数据处理（时间解析、武将查询、掉落库、战令/大将军）
├── path_normalizer.go    # 路径规范化工具
├── pinyin_helper.go      # 拼音转换工具
├── regex_helper.go       # 正则表达式工具
└── rule_helpers.go       # 规则辅助函数（参数解析、日期解析）
```

## 核心组件

### utils.go — 通用工具函数

| 导出函数 | 说明 |
|----------|------|
| `FindSheetBySuffix()` | 根据英文名后缀查找 Sheet（支持"中文|英文"格式） |
| `MatchSheetBySuffix()` | 检查 Sheet 名是否匹配过滤器 |
| `AutoDetectEndIndex()` | 自动检测数据结束位置（连续N行空单元格） |
| `GetColEndIndex()` | 获取列的数据结束位置 |
| `GetDataEndIndex()` | 从列数据中自动检测数据结束位置 |
| `SolveEmptyAndCommit()` | 处理空值和注释（返回是否跳过检查） |
| `ParseBreakLine()` | 解析 breakLine 参数 |
| `ParseAllowEmpty()` | 解析 allowEmpty 参数 |
| `ParseAllowCommit()` | 解析 allowCommit 参数 |

### hero_rule_helper.go — 武将数据处理核心

| 导出类型 | 说明 |
|----------|------|
| `ItemConfig` | 物品配置（ItemId、Count） |
| `HeroRow` | 武将行数据（Id、Name、HeroType、IsOpen、OpenDate 等） |
| `SeasonPassHero` | 战令武将信息 |
| `GeneralHero` | 大将军武将信息 |
| `DropItemInfo` | 掉落项信息 |
| `SkillMeltInfo` | 技能熔炼信息 |
| `HeroDropPoolStatus` | 武将掉落库状态枚举 |

| 导出函数 | 说明 |
|----------|------|
| `FindHeroById()` | 根据武将 ID 查找武将行 |
| `IsHeroOpened()` | 判断武将是否已开放 |
| `FindSeasonPassHeroes()` | 查找所有战令武将 |
| `FindArenaGeneralHeroes()` | 查找所有大将军武将 |
| `GetHeroDropPoolStatus()` | 获取武将在掉落库中的状态 |
| `IsHeroSynthesisEnabled()` | 判断武将是否可合成 |
| `IsHeroMeltEnabled()` | 判断武将是否可熔炼 |
| `ParseItemCfg()` | 解析物品配置字符串 |
| `ParseDate()` / `TimeEquals()` / `TimeIsInRange()` | 时间处理函数 |

### path_normalizer.go — 路径规范化工具

| 导出类型/函数 | 说明 |
|-------------|------|
| `PathNormalizer` | 路径规范化器 |
| `NewPathNormalizer()` | 创建规范化器 |
| `Normalize()` | 转为绝对路径+正斜杠格式 |
| `Equal()` | 比较两个路径是否等价 |

### pinyin_helper.go — 拼音转换工具

| 导出类型/函数 | 说明 |
|-------------|------|
| `PinyinFormat` | 拼音输出格式枚举（PinyinCamel/PinyinLower/PinyinSnake） |
| `ConvertToPinyin()` | 中文字符串转拼音（驼峰格式） |
| `ConvertToPinyinWithFormat()` | 转为指定格式的拼音 |
| `GetPinyinVariants()` | 获取所有可能拼音组合（处理多音字） |

### regex_helper.go — 正则表达式工具

| 导出函数 | 说明 |
|----------|------|
| `ExtractValuesByRegex()` | 使用正则提取捕获组值 |
| `ParseCaptureGroups()` | 解析捕获组索引字符串 |

### rule_helpers.go — 规则辅助函数

| 导出函数 | 说明 |
|----------|------|
| `ParseIntParamWithError()` | 解析整数参数（返回 error） |
| `ParseIntParam()` | 解析整数参数（空值返回默认值） |
| `ParseIntWithError()` | 解析整数（支持十六进制） |
| `ParseBool()` | 解析布尔值 |
| `ParseSkillList()` | 解析技能列表 |

## 包依赖

### 依赖
- `excelio` — Excel 读写工具（常量、数据结构）
- `json_rule` — 规则类型定义
- `github.com/mozillazg/go-pinyin` — 拼音转换库

### 被依赖
- `coded_rules/` — 所有校验规则实现直接导入 helpers 包
- `engine` — 检查执行时使用辅助函数
- `diff` — 差异检测中使用辅助工具
