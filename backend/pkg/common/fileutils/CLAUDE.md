# fileutils

通用文件工具包，提供跨模块复用的文件操作功能。

## 功能

### OpenExcel

打开指定 Sheet 对应的 Excel 文件，根据操作系统自动选择打开方式。

```go
// 通过 Sheet 名称打开（从目录中查找对应文件）
fileutils.OpenExcel("活动表|Activity", "D:/work/config/excel")

// 直接打开文件路径
fileutils.OpenExcel("", "D:/work/config/excel/Activity.xlsx")
```

| 参数 | 说明 |
|------|------|
| sheetName | Sheet 名称（如"活动表|Activity"），为空时直接打开 filePathOrDir |
| filePathOrDir | Excel 配置目录路径，或当 sheetName 为空时为文件完整路径 |
