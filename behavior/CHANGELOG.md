# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `.gitignore` for project root
- `Makefile` for common development tasks
- `.golangci.yml` for code quality configuration
- GitHub Actions CI/CD workflow
- Main `README.md` documentation
- `LICENSE` (MIT)
- `CHANGELOG.md`

## [1.0.0] - 2026-03-15

### Added
- **v1 (函数式实现)**
  - 纯函数式设计，节点定义为 `func(Context) Result`
  - Sequence（序列）节点
  - Selector（选择器）节点
  - Parallel（并行）节点
  - Condition（条件）节点
  - Action（动作）节点
  - Inverter（反转器）装饰器
  - Repeater（重复器）装饰器
  - UntilSuccess 装饰器
  - UntilFailure 装饰器
  - 完整的单元测试（覆盖率 98%+）
  - 模糊测试
  - 使用示例和详细文档

- **v2 (面向对象实现)**
  - 面向对象设计，节点定义为 `Node` 接口
  - Sequence（序列）节点
  - Selector（选择器）节点
  - Parallel（并行）节点
  - Condition（条件）节点
  - Action（动作）节点
  - Inverter（反转器）装饰器
  - Repeater（重复器）装饰器，支持 `Reset()` 方法
  - UntilSuccess 装饰器
  - UntilFailure 装饰器
  - 完整的单元测试（覆盖率 100%）
  - 模糊测试
  - 使用示例和详细文档

### Features
- 空子节点边界处理
- v1 Repeater 使用闭包状态，避免 context 冲突
- v2 Repeater 支持 Reset 方法重置状态
- 两个版本 API 风格一致，便于选择

[Unreleased]: https://github.com/wangtengda/gobee/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/wangtengda/gobee/releases/tag/v1.0.0
