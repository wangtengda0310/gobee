# rain-excel-checker

策划配表规则检查工具，用于 CI/CD 流水线自动检查 Excel 配表是否符合预设规则。

## 功能特性

- **列级规则检查**：唯一性、非空、枚举、数值范围、日期格式等
- **表级规则检查**：跨表关联、时间预警、业务逻辑验证
- **变更检测**：新增行/列、删除行/列、字段值变化通知
- **增量检查**：支持只检查 git diff 变更的文件

## 快速开始

```bash
# 全量检查
go run main.go -excelPath=/path/to/excel -casePath=/path/to/rules

# 增量检查(只检查变更文件)
go run main.go -excelPath=/path/to/excel -casePath=/path/to/rules -diffFiles="Hero.xlsx\nArenaSeason.xlsx"
```

## 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| -feishuRobot | LJH | 飞书机器人 |
| -excelPath | - | Excel 目录 |
| -casePath | - | 检查规则目录 |
| -diffFiles | "" | 变更文件列表(\n 分隔)|

## 文档

开发者文档请参考 [CLAUDE.md](./CLAUDE.md)

## 模块依赖

```
rain-excel-checker
├── feishu-lib          # 飞书通知
└── rain-qa-func        # QA 工具(调用方)
```
