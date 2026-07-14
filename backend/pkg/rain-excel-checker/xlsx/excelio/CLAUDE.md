# excelio 包 — Excel 文件读取和解析的核心工具包

Excel 文件读取、解析和数据结构定义，为整个 rain-excel-checker 项目提供基础的 Excel 文件处理能力。支持单文件和目录批量读取，提供并发读取模式。

## 文件结构

```
excelio/
├── const.go     # 配表格式常量定义
├── types.go     # 核心数据结构（Sheet、列属性、工作表类型）
├── loader.go    # Excel 文件加载器（单文件/目录/并发读取）
├── reader.go    # 高级读取接口（GetSheetMap/GetSheetMapFromBytes）
├── parser.go    # Excel 文件过滤器
└── colutil.go   # 列数据处理工具（时间解析、列值读取、布尔值解析）
```

## 核心组件

### const.go — 配表格式常量

| 常量 | 值 | 说明 |
|------|---|------|
| `MJS_FIXED_ROWS_CHS` | 0 | 普通配表第1行（中文名称行）索引 |
| `MJS_FIXED_ROWS_TYPE` | 1 | 普通配表第2行（类型定义行）索引 |
| `MJS_FIXED_ROWS_NAME` | 2 | 普通配表第3行（属性名称行）索引 |
| `MJS_FIXED_ROWS_CAS` | 3 | 普通配表第4行（导出标识行）索引 |
| `MJS_FIXED_ROWS_NUM` | 4 | 普通配表固定表头行数 |
| `MJS_FIXED_ENUM_ROWS_CHS` | 0 | 枚举配表第1行（列定义行）索引 |
| `MJS_FIXED_ENUM_ROWS_NUM` | 1 | 枚举配表固定表头行数 |

### types.go — 数据结构定义

| 导出类型 | 说明 |
|----------|------|
| `Sheet` | Excel 工作表数据结构（名称、类型、表头、错误信息） |
| `SheetHeader` | 工作表表头信息（所有列属性定义） |
| `ColAttr` | 列属性定义（类型、名称、状态等） |
| `EColType` | 列状态枚举（NORMAL/COMMENT/ENUM/EMPTY/ERROR） |
| `SheetType` | 工作表类型枚举（NONE/MING_JIANG_SHA/MING_JIANG_SHA_ENUM） |

### loader.go — 文件加载器

| 导出函数 | 说明 |
|----------|------|
| `ReadXlsx()` | 读取单个 Excel 文件 |
| `ReadFileOrDir()` | 读取文件或目录（单文件/目录递归） |
| `ReadFileOrDirConcurrent()` | 并发读取文件或目录（工作池模式） |

### reader.go — 高级读取接口

| 导出函数 | 说明 |
|----------|------|
| `GetSheetMap()` | 从目录读取所有 Excel，构建 Sheet 名称到文件对象的映射 |
| `GetSheetMapFromBytes()` | 从字节数据构建 Sheet 映射（用于 Git 历史版本读取） |

### parser.go — Excel 过滤器

| 导出类型/函数 | 说明 |
|-------------|------|
| `FilteredExcel` | 过滤后的 Excel 文件映射类型 |
| `ExcelFilter()` | 过滤 Excel 文件，识别符合项目规范的 Sheet |
| `IsValidBusinessSheetName()` | 检查 Sheet 名称是否符合"中文|英文"格式 |

### colutil.go — 列数据处理工具

| 导出函数 | 说明 |
|----------|------|
| `ParseDate()` | 解析日期字符串（多种格式） |
| `TimeEquals()` | 判断两个时间是否精确相等 |
| `FormatDate()` | 格式化日期为 "2006-01-02" |
| `FormatDateTime()` | 格式化日期时间为 "2006-01-02 15:04:05" |
| `GetColIndexByName()` | 根据列名获取列索引 |
| `GetColValue()` | 获取指定列指定行的值 |
| `ParseBool()` | 解析布尔值 |

## 包依赖

### 依赖
- `github.com/xuri/excelize/v2` — Excel 文件读取库

### 被依赖
- `engine` — 使用 `GetSheetMap`/`GetSheetMapFromBytes` 读取 Excel
- `diff` — 使用 `GetSheetMapFromBytes` 从 Git 历史读取
- `coded_rules` — 通过 `Sheet` 对象获取列数据和元信息
- `helpers` — 使用常量和数据结构
- `workflow` — 使用读取和解析功能

## 关键行为

- **Sheet 名称业务格式**：必须包含 `|` 分隔符（中文|英文），英文部分以字母或下划线开头
- **并发读取配置**：Worker 数 = max(8, CPU 核心数)，文件通道缓冲 100
- **导出标识有效值**：`""` / `"server"` / `"client"` / `"server/client"` / `"client/server"`
- **日期格式支持**：`2006-01-02 15:04:05`、`2006/01/02 15:04:05`、`2006-01-02`、`2006/01/02`
