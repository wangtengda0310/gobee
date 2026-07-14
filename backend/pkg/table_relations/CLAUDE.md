# table_relations

表格关系解析和验证，用于活动配置表的跨表数据校验。

## 核心函数
| 函数 | 文件位置 | 职责 |
|------|----------|------|
| LoadSheetData | loader.go:28 | 加载指定表的列数据 |
| GetStartRowIndex | loader.go:62 | 获取数据起始行索引（固定 4） |
| GetDataEndIndex | loader.go:86 | 获取数据结束行索引 |
| ParseDrawSkinData | parser.go:36 | 解析皮肤抽奖表数据 |
| ParseDropRuleData | parser.go:184 | 解析掉落规则表数据 |
| ValidateRelation | validator.go:34 | 验证表关联关系有效性 |
| ValidateTableExists | validator.go:57 | 验证表是否存在 |
| ValidateFieldExists | validator.go:82 | 验证字段是否存在于指定表中 |

## 开发注意事项
- 支持 MJS 格式的 Excel 表格（固定 4 行表头）
- 自动检测数据结束位置（连续 3 行空单元格）
- 支持数组字段索引语法（如 `CustomParma[0]`）
- 道具消耗格式支持 `{itemId;count}` 多值格式

## 依赖关系
- 依赖：rain-excel-checker/xlsx/excel_internal、pkg/rain-resources-checker/mjs_excel
