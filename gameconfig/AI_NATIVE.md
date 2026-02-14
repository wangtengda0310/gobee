# gameconfig AI 原生开发工具规划

## 概述

将 gameconfig 从单纯的配置加载工具升级为 AI 原生开发工具，使得 AI 能够：
1. 🎯 主动配置配置表（而非被动加载）
2. 🔍 审查和优化配置（而非报错）
3. 🧪 自动生成测试数据（而非手动编写）
4. 📊 生成文档和示例（而非过时）
5. 🌐 跨语言复用（而非单一语言）

---

## 一、核心能力矩阵

### 1.1 当前状态

| 维度 | 当前状态 | AI 原生机会 |
|------|----------|-------------|
| **读取** | ✅ Excel/CSV 加载 | 🤖 自动识别格式、推断字段类型、生成结构体 |
| **映射** | ✅ 结构体反射映射 | 🤖 根据数据推断结构体、自动匹配字段名 |
| **验证** | ✅ 类型检查 + 必填验证 | 🤖 业务规则验证、数据质量检查、一致性校验 |
| **测试** | ✅ Mock 数据模式 | 🤖 🤖 AI 生成全面测试数据（边界、异常、业务逻辑） |
| **导出** | ✅ Excel 导出 | 🤖 AI 优化格式、生成文档、自动排版 |
| **版本** | ✅ Schema 迁移 | 🤖 AI 分析变更、生成迁徜脚本、影响分析 |
| **审查** | ❌ 无 | 🤖 🔍 AI 审查配置表、发现问题和优化点 |
| **分析** | ❌ 无 | 🤖 📊 AI 分析数据分布、发现业务洞察 |

### 1.2 能力分层

```
┌─────────────────────────────────────────────────┐
│            AI 原生能力金字塔                │
├─────────────────────────────────────────────────┐
│                                            │
│                    ┌─────────┐      │
│                    │  🤖 主动  │      │
│                    │  配置  │      │
│                    └─────────┘      │
│                            │                │
│            ┌───────────────┐              │
│            │     🔍 审查  │              │
│            │    分析  │              │
│            └───────────────┘              │
│                            │                │
│        ┌─────────────────────────┐         │
│        │       📊 生成  │         │
│        │   测试/文档  │         │
│        └─────────────────────────┘         │
│                            │                │
│  ┌─────────────────────────────────┐    │
│  │       ✅ 基础  │    │
│  │    加载/映射  │    │
│  └─────────────────────────────────┘    │
│                            │                │
└─────────────────────────────────────────┘
```

---

## 二、AI 原生功能清单

### 2.1 🎯 配置模板生成器

**功能描述**：根据业务需求自动生成配置表模板和结构体定义

**使用场景**：
```go
// 策划输入业务需求
需求 := "武器系统：装备有 ID、名称、攻击力、类型（普通/魔法/传说）"

// AI 自动生成
1. Excel 模板文件（包含字段定义、类型、示例）
2. Go 结构体定义
3. 字段说明文档
4. 测试数据模板
```

**技术实现**：
```go
type ConfigGenerator struct {
    LLMClient *anthropic.Client
}

func (g *ConfigGenerator) GenerateFromRequirement(req string) (*ConfigTemplate, error) {
    // 1. LLM 分析需求，提取字段信息
    fields, err := g.extractFields(req)

    // 2. 生成结构体定义
    structDef := g.generateStruct(fields)

    // 3. 生成 Excel 模板
    excelPath := g.generateExcel(fields)

    // 4. 生成文档
    docs := g.generateDocs(fields)

    return &ConfigTemplate{
        StructDef:  structDef,
        ExcelPath:  excelPath,
        Docs:      docs,
    }, nil
}
```

**用户价值**：
- 🚀 加速策划配置表创建（从小时到分钟）
- 📏 减少 manual 错误（类型错误、字段遗漏）
- 🎓 统一团队规范（自动遵循命名约定）

---

### 2.2 🔍 配置审查器

**功能描述**：AI 审查现有配置表，发现潜在问题和优化建议

**使用场景**：
```bash
$ gameconfig review config/装备表.xlsx

✅ 配置表结构正常
⚠️ 发现 5 处潜在问题：
  - 'attack' 字段建议添加范围检查 (0-10000)
  - 'type' 字段建议使用枚举注释 [普通/魔法/传说]
  - 缺少 'description' 字段（建议添加）
  - 发现 3 条记录的 type 字段超出范围 [0,2]
  - 发现 12 条记录的 attack < 0（建议检查）

💡 优化建议：
  - 添加 type 字段的枚举验证
  - attack 建议添加范围 [0,10000]
  - 考虑添加 description 字段提升可读性
```

**技术实现**：
```go
type ConfigReviewer struct {
    Loader    *config.Loader[any]
    LLMClient *anthropic.Client
}

func (r *ConfigReviewer) Review() (*ReviewReport, error) {
    // 1. 加载配置数据
    data, err := r.Loader.Load()

    // 2. 结构分析（字段类型、必填项）
    structIssues := r.analyzeStructure(data)

    // 3. 数据质量分析（范围、分布、异常）
    dataIssues := r.analyzeDataQuality(data)

    // 4. LLM 业务规则审查
    businessIssues := r.reviewBusinessRules(data)

    // 5. 生成优化建议
    suggestions := r.generateSuggestions(structIssues, dataIssues, businessIssues)

    return &ReviewReport{
        Issues:      append(structIssues, dataIssues...),
        Suggestions: suggestions,
        Score:        r.calculateScore(),
    }, nil
}
```

**用户价值**：
- 🛡️️ 提升配置质量（提前发现错误）
- 📋 最佳实践推荐（基于行业规范）
- 🎓 知识传承（新策划快速上手）

---

### 2.3 🧪 测试数据生成器

**功能描述**：根据结构体定义自动生成全面的测试数据

**使用场景**：
```go
type Equipment struct {
    ID      int    `excel:"id"`
    Type    int    `excel:"type"`     // 0:普通 1:武器 2:盔甲
    Attack  int    `excel:"attack,when:type=1"`
    Defense int    `excel:"defense,when:type=2"`
}

// AI 生成测试数据
mockData := GenerateMockData[Equipment](
    WithCoverage(),      // 覆盖所有字段
    WithEdgeCases(),     // 边界情况（空值、极值）
    WithBusinessLogic(), // 业务逻辑（type=0 时无 attack）
    WithRealisticDistribution(), // 真实数据分布
)

// 输出
loader.SetMockData(mockData)
```

**生成策略**：
| 测试类型 | 覆盖内容 | 示例 |
|----------|----------|------|
| **基本覆盖** | 所有字段有值 | 正常范围值 |
| **边界测试** | 最小值、最大值 | 0, 999999 |
| **空值测试** | 空值和默认值 | "", 0 |
| **异常测试** | 超出范围值 | -1, 999999 |
| **业务逻辑** | 符合条件字段 | type=1 时 attack>0 |
| **真实分布** | 符合实际使用 | 70% 普通, 20% 武器, 10% 盔甲 |

**用户价值**：
- 🎯 自动化测试数据生成（节省手动编写时间）
- 🧪 覆盖率高（发现边界 bug）
- 🐛 发现边界 bug（异常值处理）

---

### 2.4 📊 文档生成器

**功能描述**：根据配置表自动生成文档

**生成内容**：
```markdown
# 装备表配置说明

## 表结构
| 字段名 | 类型 | 必填 | 说明 | 范围 |
|--------|------|------|------|------|
| id      | int  | ✅  | 装备 ID | 1000-9999 |
| name    | string | ✅  | 装备名称 | - |
| type    | int  | ✅  | 装备类型 | 0-2 |
| attack  | int  | ❌  | 攻击力 | 0-10000（type=1） |

## 枚举值说明
### type (装备类型)
| 值 | 说明 |
|----|------|
| 0  | 普通道具 |
| 1  | 武器 |
| 2  | 盔甲 |

## 业务规则
- type=1 时，attack 必须 > 0
- type=2 时，defense 必须 > 0
- attack 和 defense 不能同时 > 0

## 代码示例
\`\`\`go
loader := config.NewLoader[Equipment]("config/装备表.xlsx", "武器")
items, err := loader.Load()
for _, item := range items {
    fmt.Printf("%s: 攻击=%d\n", item.Name, item.GetAttack())
}
\`\`\`
```

**用户价值**：
- 📚 保持文档同步（配置变更自动更新）
- 🆘 新成员快速上手（文档完整）
- 📖 API 文档自动更新（无需手动维护）

---

### 2.5 🔄 配置迁徜助手

**功能描述**：配置表结构变更时，AI 自动生成迁徜脚本

**使用场景**：
```go
// v1: 老装备定义
type EquipmentV1 struct {
    ID      int    `excel:"id"`
    Name    string `excel:"name"`
    Power   int    `excel:"power"`      // 旧字段：总战力
}

// v2: 新装备定义
type EquipmentV2 struct {
    ID      int    `excel:"id"`
    Name    string `excel:"name"`
    Attack  int    `excel:"attack"`     // 新字段：分离攻击力
    Defense int    `excel:"defense"`    // 新字段：防御力
}

// AI 分析变更
migration := AutoGenerateMigration(EquipmentV1{}, EquipmentV2{})
// AI 生成迁徜规则：
// v1.Power = v2.Attack + v2.Defense
```

**迁徜类型**：
| 变更类型 | 说明 | 示例 |
|----------|------|------|
| **新增列** | 添加新字段，使用默认值 | 添加 quality 字段，默认 0 |
| **删除列** | 移除字段，数据丢失 | 移除 power 字段 |
| **重命名** | 字段名调整 | name -> display_name |
| **拆分列** | 一列拆多列 | power → attack + defense |
| **合并列** | 多列合一列 | min_atk + max_atk → atk_range |
| **类型变更** | 字段类型转换 | price: int → float |

**用户价值**：
- 🛡️️ 安全的数据升级（减少数据丢失）
- 🚀 减少手动迁徜错误
- 📊 版本管理自动化

---

### 2.6 🎨 CLI 智能工具

**功能描述**：交互式配置管理命令行工具

**交互流程**：
```bash
$ gameconfig wizard

🤖 gameconfig 智能向导
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. 创建新配置表
2. 审查现有配置
3. 生成测试数据
4. 导出为 CSV
5. 查看配置差异
6. 配置迁徜助手

请选择操作 (1-6): 1

📋 请输入业务需求：
> 武器系统，包含 ID、名称、攻击力、类型

🎯 AI 正在分析需求...
✅ 推断：需要 4 个字段
📋 建议结构体定义：

type Weapon struct {
    ID      int    `excel:"id"`
    Name    string `excel:"name,required"`
    Attack  int    `excel:"attack,default:0"`
    Type    int    `excel:"type,default:0"`  // 0:普通 1:魔法
}

是否生成 Excel 模板？ (y/n): y
✅ 已生成：config/武器表.xlsx

下一步？
1. 添加更多字段
2. 生成测试数据
3. 查看文档
> 2

✅ 已生成测试数据（100 条）
是否继续？ (y/n):
```

**用户价值**：
- 🎯 降低使用门槛（无需学习 API）
- 🤖 智能交互（向导式操作）
- 🚀 快速原型开发（分钟级创建配置）

---

### 2.7 🔗 配置分析器

**功能描述**：AI 分析配置数据质量和业务洞察

**分析报告**：
```markdown
# 装备表数据分析报告

## 基础统计
- 总记录数：1,234 条
- 字段数：6 个
- 最后更新：2025-02-14

## 数据完整性
✅ 必填字段：100% 完整
⚠️ 缺失值：0 条记录
⚠️ 格式错误：0 条记录

## 数据质量
### 数值范围
| 字段 | 最小 | 最大 | 平均 | 异常值 |
|------|------|------|------|--------|
| attack | 0 | 9999 | 125.3 | 3 条 > 10000 |
| defense | 0 | 5000 | 89.7 | 1 条 < 0 |

### 数据分布
| type | 数量 | 占比 |
|------|------|------|
| 0 (普通) | 823 | 66.7% |
| 1 (武器) | 298 | 24.1% |
| 2 (盔甲) | 113 | 9.2% |

💡 业务洞察：
- 普通道具占比过高（67%），建议平衡游戏经济
- 盔甲数量较少（9%），可能是防御属性不受欢迎
- 建议调整掉落概率或属性平衡

## 优化建议
1. attack 建议添加范围 [0,10000]
2. type 字段建议使用枚举注释
3. 考虑添加 description 字段提升可读性
```

**用户价值**：
- 🔍 数据质量监控（持续跟踪配置健康度）
- 📈 业务洞察（数据驱动的游戏平衡建议）
- 🛡️️ 优化建议（发现潜在问题）

---

## 三、多语言适配架构

### 3.1 为什么需要多语言支持？

**业务痛点**：
- 🏢 Java 服务器 + Unity 客户 + Python 工具，各用各的配置工具
- 📚 策划不懂技术栈，只想用 Excel
- 🔄 不同团队配置规范不统一，难以协作

**解决方案**：
```
┌─────────────────────────────────────────────────┐
│          统一配置管理平台                │
├─────────────────────────────────────────────────┐
│                                            │
│  ┌──────────────┐                     │
│  │   Excel/CSV   │                     │
│  │   统一引擎   │  (Go)               │
│  └──────────────┘                     │
│         │                               │
│         ├──────┐                       │
│         │ FFI  │                       │
│         │ RPC  │                       │
│         └──────┘                       │
│         │                               │
│  ┌───────────────────────────────────┐     │
│  │         语言绑定层         │     │
│ ├───────────────────────────────────┤     │
│  │                                  │     │
│  │  ┌─────────┐  ┌─────────┐       │     │
│  │  │ Java  │  │Python │  │ ... │     │
│  │  │Binding│  │Binding │       │     │
│  │  └─────────┘  └─────────┘       │     │
│  │                                  │     │
│  └───────────────────────────────────┘     │
│         │                                  │
│         └──────────┐                      │
│         │  AI 工具链 │                     │
│         │  (审查/分析/生成)                 │
│         └──────────┘                      │
│                                            │
└─────────────────────────────────────────────────┘
```

### 3.2 技术架构

**核心引擎**：当前 Go 实现
- Excel/CSV 读取
- 结构体映射
- Schema 迁移
- 条件字段

**统一接口层**：
```go
// 所有语言实现的统一接口
package gameconfig

// Loader 接口
type Loader[T any] interface {
    Load() ([]T, error)
    Reload() error
    GetSheetNames() ([]string, error)
}

// Writer 接口
type Writer[T any] interface {
    Write(data []T) error
    WriteSheet(sheet string, data []T) error
}

// Validator 接口
type Validator interface {
    Validate(data interface{}) ([]ValidationIssue, error)
}
```

**语言绑定优先级**：
| 语言 | 优先级 | 复杂度 | 原因 |
|------|--------|--------|------|
| **Python** | ⭐⭐⭐⭐⭐ | ⭐ 简单 | 数据类、数据科学工具、快速原型 |
| **TypeScript** | ⭐⭐⭐⭐ | ⭐⭐ 中等 | Web 前端、Node.js 工具链 |
| **Java** | ⭐⭐⭐ | ⭐⭐⭐ 较杂 | 企业服务端、反射、类型系统 |
| **C#** | ⭐⭐ | ⭐⭐⭐ 最复杂 | Unity、Godot 游戏引擎 |
| **C++** | ⭐ | ⭐⭐⭐⭐⭐ 最复杂 | 高性能游戏引擎 |

### 3.3 实现路径

**Phase 1：抽象接口层**（2-3 周）
- 设计统一接口（Loader、Writer、Validator）
- 定义错误处理规范
- 文档化接口契约

**Phase 2：FFI/RPC 服务**（4-6 周）
- 将 Go 引擎暴露为 FFI 动态库
- 或实现 gRPC/HTTP 服务
- 跨语言调用测试

**Phase 3：Python 绑定**（优先级最高）
- Python dataclass 绑定
- 类型提示支持
- pip 包发布

**Phase 4：TypeScript/Java 绑定**
- TypeScript 接口定义
- Java 反射集成

---

## 四、Agent Skills 能力注入

### 4.1 什么是 Agent Skills？

**概念**：将 gameconfig 的 AI 能力封装成 Claude Code Skills，使得引入 gameconfig 后自动获得配置管理能力。

**价值**：
- 🎯 新项目开箱即用（引入 skill → 自动获得配置审查、测试生成等能力）
- 📚 工具链效应（gameconfig 能力可被其他工具复用）
- 🤖 渐进增强（从基础 skill 开始，逐步添加高级能 as）

### 4.2 Skill Manifest

```json
{
  "name": "gameconfig-skills",
  "version": "1.0.0",
  "description": "游戏配置管理 AI 能力集",
  "skills": [
    {
      "name": "review-config",
      "description": "审查游戏配置表，发现潜在问题和优化点",
      "parameters": [
        {"name": "configPath", "type": "string", "description": "配置文件路径"}
      ],
      "examples": [
        "审查装备表配置",
        "检查 config/装备表.xlsx 有什么问题"
      ],
      "implementation": "internal/skills/review_config.go"
    },
    {
      "name": "generate-mock-data",
      "description": "根据结构体定义生成测试数据",
      "parameters": [
        {"name": "structType", "type": "string", "description": "结构体名称"},
        {"name": "coverage", "type": "enum", "options": ["basic", "comprehensive", "edge-cases"]}
      ],
      "examples": [
        "为 Weapon 结构体生成测试数据",
        "生成装备表的完整测试用例"
      ],
      "implementation": "internal/skills/generate_mock.go"
    },
    {
      "name": "analyze-schema",
      "description": "分析配置表结构变更并生成迁徜建议",
      "parameters": [
        {"name": "oldVersion", "type": "string"},
        {"name": "newVersion", "type": "string"}
      ],
      "implementation": "internal/skills/analyze_schema.go"
    }
  ]
}
```

### 4.3 Skill 实现示例

```go
// internal/skills/review_config.go
package skills

// ReviewConfigSkill 审查配置表 skill
type ReviewConfigSkill struct {
    configPath string
}

// Execute skill执逻辑
func (s *ReviewConfigSkill) Execute(ctx *SkillContext) (*SkillResult, error) {
    // 1. 读取配置
    loader := config.NewLoader(/* ... */)
    data, err := loader.Load()

    // 2. AI 分析（调 LLM）
    issues := s.analyzeWithAI(data)

    // 3. 生成报告
    return &SkillResult{
        Summary: fmt.Sprintf("发现 %d 个潜在问题", len(issues)),
        Details: issues,
        Suggestions: s.generateSuggestions(issues),
    }, nil
}

// analyzeWithAI 调 LLM 分析配置
func (s *ReviewConfigSkill) analyzeWithAI(data interface{}) []Issue {
    prompt := fmt.Sprintf(`
        审查以下游戏配置数据，发现潜在问题：
        %+#v

        检查项：
        1. 必填字段是否缺失
        2. 数值范围是否合理
        3. 枚举值分布是否异常
        4. 是否有重复数据
        5. 字段命名是否规范
        `, data)

    // 调 Claude 分析
    response, err := callLLM(prompt)
    return parseIssues(response)
}
```

### 4.4 能力矩阵

| Skill | 状态 | 复杂度 | 用户价值 |
|-------|------|--------|----------|
| 审查配置表 | ✅ 可实现 | ⭐⭐ 简单 | 🛡️️ 提升质量 |
| 生成测试数据 | ✅ 可实现 | ⭐⭐⭐ 中等 | 🚀 加速开发 |
| 分析 Schema | ✅ 可实现 | ⭐⭐⭐⭐ 较杂 | 🔄 安全升级 |
| 生成文档 | 🚧 规划中 | ⭐⭐⭐ 中等 | 📚 知识传承 |
| 配置向导 | 🚧 规划中 | ⭐⭐⭐ 较杂 | 🎯 降低门槛 |
| 数据分析 | 🚧 规划中 | ⭐⭐⭐⭐ 最复杂 | 📈 业务洞察 |

---

## 五、发展路线图

### 5.1 阶段规划（3 个月）

**月份 1：Agent Skills + Python 绑定**
- Week 1-2：实现 3 个核心 skills（审查、测试生成、文档）
- Week 3-4：Python 绑定 + FFI 接口
- Week 5-6：测试和文档

**月份 2：AI 能力扩展**
- Week 7-8：配置模板生成器
- Week 9-10：配置迁徜助手
- Week 11-12：CLI 智能工具

**月份 3：生态扩展**
- Week 13-16：TypeScript/Java 绑定
- Week 17-20：更多 AI skills
- Week 21-24：开发者工具集成（IDE 扩展）

### 5.2 里程碑

| 里程碑 | 时间 | 交付物 |
|--------|------|--------|
| **M1** | Week 4 | Python 绑定 + 3 个 skills |
| **M2** | Week 8 | 配置模板生成 + 测试生成 |
| **M3** | Week 12 | CLI 智能工具 |
| **M4** | Week 16 | TypeScript 绑定 |
| **M5** | Week 20 | Java 绑定 |
| **M6** | Week 24 | IDE 扩展预览版 |

---

## 六、技术选型

### 6.1 LLM 集成

**推荐**：Claude API（Anthropic）
- 原因：工具本身基于 Claude Code，生态统一
- 能力：强、推理能力适合配置分析和生成

**备选**：
- OpenAI GPT-4：成本优势
- 本地模型（Llama 3）：隐私敏感场景

### 6.2 跨语言通信

**FFI vs RPC vs MCP**

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| **FFI** | 性能最好、类型安全 | 需要编译绑定 | 高性能场景（C++ 游戏） |
| **gRPC** | 跨语言、生态成熟 | 需要网络开销 | 微服务架构 |
| **HTTP REST** | 简单、通用 | 性能较差 | 简单场景 |
| **MCP** | AI 原生、与 Claude Code 深度集成 | 新技术、生态小 | AI 工具链 |

**推荐**：提供多种接口
- FFI：高性能场景（C++/Unity）
- gRPC：企业服务（Java/Go）
- MCP：AI 工具集成（优先实现）

### 6.3 存储

**配置存储**：
- 本地文件系统（当前）
- Git 仓库（推荐：版本控制）
- S3/OSS（云端协作）

---

## 七、成功指标

### 7.1 开发者效能

| 指标 | 当前 | 目标 | 测量方式 |
|------|------|------|----------|
| 配置加载时间 | 10 分钟 | 1 分钟 | ⬇️️ 手动操作时间 |
| 测试数据生成 | 30 分钟 | 1 分钟 | ⬇️️ 手动编写时间 |
| 配置审查 | 2 小时 | 5 分钟 | ⬇️️ AI 审查时间 |
| 文档更新 | 1 小时 | 自动 | ✅ 完全自动化 |

### 7.2 配置质量

| 指标 | 当前 | 目标 | 测量方式 |
|------|------|------|----------|
| 配置错误率 | 5-10% | <1% | 📉 上线 bug 数量 |
| 数据完整性 | 90% | >99% | 📈 必填字段缺失率 |
| 文档覆盖率 | 60% | 100% | 📚 文档与代码同步率 |

---

## 八、风险与挑战

### 8.1 技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| **跨语言性能** | FFI 调用可能有性能开销 | 提供本地缓存、批量处理 |
| **LLM 成本** | API 调用增加成本 | 智能缓存、批量处理 |
| **类型系统差异** | 不同语言类型系统差异大 | 抽象接口、统一类型映射 |

### 8.2 业务风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| **学习曲线** | 团队需要学习新工具 | 详细文档、培训材料 |
| **迁移成本** | 现有项目迁移需要时间 | 提供迁移工具、渐进式迁移 |

---

## 九、下一步

### 立即行动（本周）

1. ✅ **创建 Golang 开发 skill**
   - 引导 AI 快速使用 gameconfig
   - 包含完整示例和最佳实践

2. ✅ **设计 FFI 接口**
   - 定义跨语言调用接口
   - 编写技术设计文档

3. ✅ **实现配置审查 skill**
   - 第一个可用的 AI skill
   - 验证技术可行性

### 短期规划（本月）

1. 🚧 Python 绑定实现
2. 🚧 配置测试数据生成器
3. 🚧 CLI 智能工具原型

---

**文档版本**：v1.0
**最后更新**：2025-02-14
**维护者**：gameconfig 团队
