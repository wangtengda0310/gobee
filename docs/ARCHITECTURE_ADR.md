# 架构决策记录 (ADR)

## 决策背景

**日期**: 2025-01-31
**决策**: 统一 lvan_dumper 项目架构
**状态**: 提议

### 问题描述

当前项目存在以下架构问题：

1. **全局变量依赖**：`internal/acce/acci`、`TransImport/TransExport` 都是全局变量
2. **访问者模式使用不一致**：部分代码未使用此模式
3. **v1 和 v2 数据源并存**：需要统一
4. **go-spring 未完成**：代码中存在未完成的 go-spring 集成

### 需求

1. **父命令解析参数 → 创建对象 → 子命令直接使用**
2. **支持命令分散在不同目录**（Cobra 最佳实践）
3. **避免全局变量**
4. **保持类型安全**

---

## 考虑的方案

### 方案1: Context 传递模式

**核心思想**：使用 `context.Context` 传递已创建的对象

**优点**：
- ✅ Go 标准做法
- ✅ 支持跨目录
- ✅ 无额外依赖

**缺点**：
- 🟡 需要 Cobra 1.2+ 的 `cmd.SetContext()` 支持
- 🟡 类型断言需要运行时检查

### 方案2: 接口抽象模式

**核心思想**：定义清晰的接口，依赖注入

**优点**：
- ✅ 职责清晰
- ✅ 易测试
- ✅ 可扩展

**缺点**：
- 🟡 代码量较多
- 🟡 仍需某种方式传递 context

### 方案3: go-spring 依赖注入

**核心思想**：使用框架管理依赖生命周期

**优点**：
- ✅ 功能强大
- ✅ 自动管理生命周期

**缺点**：
- 🔴 文档不完整
- 🔴 学习曲线陡峭
- 🔴 过度设计（对小型项目）

---

## 决策结果

**选择方案1：Context 传递模式 + 接口抽象**

### 决策理由

1. **符合 Go 惯用法**：使用标准库的 context 包
2. **支持当前项目结构**：不需要大规模重构
3. **向前兼容**：为未来可能的 go-spring 迁移留有余地
4. **简单高效**：代码改动量适中，易于维护

---

## 实施计划

### 阶段1：创建新的 context 管理包

```
cmd/context/
├── context.go      # Context 管理器
├── keys.go         # context key 定义
└── manager.go      # Manager 接口和实现
```

### 阶段2：重构 service 层

```
pkg/dump/service/
├── datasource.go   # DatasourceManager 接口
├── mysql.go        # MySQL 实现
└── dump.go         # DumpService
```

### 阶段3：重构命令层

- 移除 `internal/accept.go` 中的全局变量
- 使用 `cmd.SetContext()` 传递 manager
- 更新所有子命令使用新的 context 获取方式

### 阶段4：回归测试

- 运行现有测试用例
- 确保 dump/import 功能正常

---

## 架构对比

| 特性 | 当前架构 | 新架构 |
|------|---------|--------|
| 依赖传递 | 全局变量 | context.Context |
| 类型安全 | 🟡 运行时 | ✅ 编译期 |
| 跨目录支持 | 🟡 有限 | ✅ 完全支持 |
| 测试性 | 🟡 中 | ✅ 高 |
| 可维护性 | 🟡 中 | ✅ 高 |

---

## 参考资料

- [Cobra Working with Commands](https://cobra.dev/docs/how-to-guides/working-with-commands/)
- [Go Context Package](https://pkg.go.dev/context)
- [go-spring GitHub](https://github.com/go-spring/spring-core)
