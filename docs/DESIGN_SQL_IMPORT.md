# MySQL Dump SQL 文件导入功能设计

## 功能概述

支持从 `mysql-dump` 导出的 SQL 文件导入数据，转换为 LVAN Dumper 的标准 ZIP/DIR 格式。

**采用方案**: Hybrid 方案（快速解析 + MySQL 回退）

---

## 架构设计

### 命令结构

```
connect mysql import-sql <dump.sql> [options]
```

### 数据流程

```
┌─────────────────┐
│  dump.sql       │
│  (mysql-dump)   │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  快速解析尝试                        │
│  - 检测文件格式                      │
│  - 尝试解析标准 INSERT               │
└────────┬────────────────────────────┘
         │
    ┌────┴────┐
    │         │
   成功      失败
    │         │
    │         ▼
    │    ┌─────────────────────────────┐
    │    │  MySQL/Dolt 回退方案         │
    │    │  - 创建临时数据库            │
    │    │  - 导入 SQL                 │
    │    │  - 使用现有 dump 功能       │
    │    └────────┬────────────────────┘
    │             │
    └──────┬──────┘
           ▼
┌─────────────────────────────────────┐
│  标准化输出                         │
│  - ZIP 格式                         │
│  - DIR 格式                         │
│  - 数据验证                          │
└─────────────────────────────────────┘
```

---

## 代码设计

### 1. 命令定义

**文件**: `lvan/cmd/dumper/cmd/import_sql.go`

```go
package cmd

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "time"

    "github.com/spf13/cobra"
    "github.com/wangtengda0310/gobee/lvan/cmd/dumper/cmd/internal"
    "github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

var importSqlCmd = &cobra.Command{
    Use:     "import-sql <dump.sql>",
    Aliases: []string{"import-sql", "sql-import"},
    Short:   "从 mysql-dump 导出的 SQL 文件导入",
    Long: `从 mysql-dump 导出的 SQL 文件导入数据，支持转换为 ZIP 或 DIR 格式。

支持的 SQL 格式:
  - 标准 INSERT 语句
  - 扩展插入 (INSERT ... VALUES (...),(...),(...))
  - mysql-dump 默认输出格式

示例:
  # 导入 SQL 文件为 ZIP
  connect mysql import-sql dump.sql

  # 指定输出格式
  connect mysql import-sql dump.sql --in dir

  # 指定 MySQL 连接参数（用于回退方案）
  connect mysql import-sql dump.sql -h localhost -P 3307 -u root`,
    Args: cobra.ExactArgs(1),
    PreRunE: func(cmd *cobra.Command, args []string) error {
        // 验证 SQL 文件存在
        if _, err := os.Stat(args[0]); os.IsNotExist(err) {
            return fmt.Errorf("SQL 文件不存在: %s", args[0])
        }
        return nil
    },
    Run: func(cmd *cobra.Command, args []string) {
        sqlFile := args[0]

        // 获取输出格式
        outputFormat := viper.GetString("in")

        // 获取数据库连接参数（用于回退方案）
        config := dbParams(cmd)

        // 执行导入
        if err := importSQLFile(sqlFile, config, outputFormat); err != nil {
            log.Fatalf("导入失败: %v", err)
        }
    },
}

func init() {
    mysqlCmd.AddCommand(importSqlCmd)

    // 继承 mysql 命令的所有标志
    persistentFlags(importSqlCmd)
    flags(importSqlCmd)

    // 添加特定标志
    importSqlCmd.Flags().String("temp-db", "", "临时数据库名（默认自动生成）")
    importSqlCmd.Flags().Bool("keep-temp", false, "保留临时数据库（用于调试）")
    importSqlCmd.Flags().Duration("timeout", 5*time.Minute, "导入超时时间")
}

// importSQLFile 导入 SQL 文件
func importSQLFile(sqlFile string, config dump.Config, outputFormat string) error {
    log.Printf("开始导入 SQL 文件: %s", sqlFile)

    // 方案 1: 尝试快速解析
    records, tableName, err := quickParseSQL(sqlFile)
    if err == nil && len(records) > 0 {
        log.Printf("快速解析成功: %d 条记录, 表: %s", len(records), tableName)
        return exportRecords(records, tableName, outputFormat)
    }

    // 方案 2: 回退到 MySQL/Dolt 导入
    log.Printf("快速解析失败 (%v)，使用 MySQL 回退方案", err)
    return importViaMySQL(sqlFile, config, outputFormat)
}

// quickParseSQL 快速解析 SQL 文件
func quickParseSQL(sqlFile string) ([]dump.Record, string, error) {
    parser := NewSQLInsertParser()
    return parser.Parse(sqlFile)
}

// importViaMySQL 通过 MySQL/Dolt 导入
func importViaMySQL(sqlFile string, config dump.Config, outputFormat string) error {
    tempDB := fmt.Sprintf("temp_import_%d", time.Now().Unix())
    if tempDBName := viper.GetString("temp-db"); tempDBName != "" {
        tempDB = tempDBName
    }

    log.Printf("创建临时数据库: %s", tempDB)

    // 1. 创建临时数据库
    if err := execSQL(config, fmt.Sprintf("CREATE DATABASE `%s`;", tempDB)); err != nil {
        return fmt.Errorf("创建临时数据库失败: %w", err)
    }
    defer func() {
        if !viper.GetBool("keep-temp") {
            execSQL(config, fmt.Sprintf("DROP DATABASE `%s`;", tempDB))
            log.Printf("已清理临时数据库: %s", tempDB)
        }
    }()

    // 2. 导入 SQL 文件
    absPath, _ := filepath.Abs(sqlFile)
    if err := importSQLFileToDB(config, tempDB, absPath); err != nil {
        return fmt.Errorf("导入 SQL 失败: %w", err)
    }

    // 3. 获取表名
    tables, err := getTables(config, tempDB)
    if err != nil {
        return fmt.Errorf("获取表名失败: %w", err)
    }

    log.Printf("发现 %d 个表: %v", len(tables), tables)

    // 4. 对每个表执行 dump
    for _, table := range tables {
        log.Printf("导出表: %s", table)

        // 使用现有的 dump 功能
        internal.VisitExport(func(db dump.Datasource) {
            columns := dump.Columns(db.DB, tempDB, table)
            records := dump.Dump(db.DB, tempDB, table, columns, "", "")

            pkColumns, _ := dump.GetPrimaryKeyColumns(db.DB, tempDB, table)
            log.Printf("主键: %v, 记录数: %d", pkColumns, len(records))

            // 导出
            export := internal.TransExport(records, pkColumns...)
            log.Printf("导出完成: %s", export)
        }, "", "")
    }

    return nil
}

// execSQL 执行 SQL 命令
func execSQL(config dump.Config, sql string) error {
    cmd := exec.Command("mysql",
        "-h", config.Host,
        fmt.Sprintf("-P%d", config.Port),
        "-u", config.User,
        "-p"+config.Password,
        "-e", sql,
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("执行 SQL 失败: %w, 输出: %s", err, string(output))
    }
    return nil
}

// importSQLFileToDB 导入 SQL 文件到数据库
func importSQLFileToDB(config dump.Config, database, sqlFile string) error {
    cmd := exec.Command("mysql",
        "-h", config.Host,
        fmt.Sprintf("-P%d", config.Port),
        "-u", config.User,
        "-p"+config.Password,
        database,
    )

    // 从文件读取输入
    inputFile, err := os.Open(sqlFile)
    if err != nil {
        return err
    }
    defer inputFile.Close()

    cmd.Stdin = inputFile

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("导入失败: %w, 输出: %s", err, string(output))
    }

    return nil
}

// getTables 获取数据库中的所有表
func getTables(config dump.Config, database string) ([]string, error) {
    cmd := exec.Command("mysql",
        "-h", config.Host,
        fmt.Sprintf("-P%d", config.Port),
        "-u", config.User,
        "-p"+config.Password,
        "-N", // 跳过表头
        "-e", fmt.Sprintf("SHOW TABLES FROM `%s`;", database),
    )

    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    return parseTableList(string(output)), nil
}

func parseTableList(output string) []string {
    var tables []string
    for _, line := range strings.Split(output, "\n") {
        line = strings.TrimSpace(line)
        if line != "" {
            tables = append(tables, line)
        }
    }
    return tables
}
```

---

### 2. SQL 解析器

**文件**: `lvan/pkg/dump/load/sql.go`

```go
package load

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
    "strconv"
   "strings"

    "github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

// SQLInsertParser SQL INSERT 语句解析器
type SQLInsertParser struct {
    insertRegex  *regexp.Regexp
    valuesRegex  *regexp.Regexp
    stringRegex  *regexp.Regexp
}

// NewSQLInsertParser 创建 SQL 解析器
func NewSQLInsertParser() *SQLInsertParser {
    // 匹配: INSERT INTO `table` (col1,col2) VALUES (...),(...);
    insertRe := regexp.MustCompile(
        `(?i)INSERT\s+INTO\s+` +
            "`?([^`\s]+)`?" + // 表名
            `\s*(?:\(([^)]+)\)\s*)?` + // 列名（可选）
            `VALUES\s+(.+);`, // 值列表
    )

    // 匹配值组: (val1,'str',NULL,_binary'...')
    valuesRe := regexp.MustCompile(`\(([^)]+)\)`)

    return &SQLInsertParser{
        insertRegex: insertRe,
        valuesRegex: valuesRe,
    }
}

// Parse 解析 SQL 文件
func (p *SQLInsertParser) Parse(sqlFile string) ([]dump.Record, string, error) {
    file, err := os.Open(sqlFile)
    if err != nil {
        return nil, "", err
    }
    defer file.Close()

    var allRecords []dump.Record
    var tableName string
    var columns []string

    scanner := bufio.NewScanner(file)
    lineNum := 0

    for scanner.Scan() {
        lineNum++
        line := strings.TrimSpace(scanner.Text())

        // 跳过注释和空行
        if p.shouldSkipLine(line) {
            continue
        }

        // 合并多行语句
        for !strings.HasSuffix(line, ";") && scanner.Scan() {
            lineNum++
            nextLine := strings.TrimSpace(scanner.Text())
            if p.shouldSkipLine(nextLine) {
                continue
            }
            line += " " + nextLine
        }

        // 匹配 INSERT 语句
        matches := p.insertRegex.FindStringSubmatch(line)
        if len(matches) < 3 {
            continue
        }

        // 解析表名和列名
        if tableName == "" {
            tableName = matches[1]
            if matches[2] != "" {
                columns = strings.Split(matches[2], ",")
                for i, col := range columns {
                    columns[i] = strings.Trim(strings.TrimSpace(col), "`\"")
                }
            }
            log.Printf("检测到表: %s, 列: %v", tableName, columns)
        } else if matches[1] != tableName {
            return nil, tableName, fmt.Errorf("多表导入暂不支持: %s vs %s", tableName, matches[1])
        }

        // 解析值
        valuesStr := matches[3]
        records, err := p.parseValues(valuesStr, columns)
        if err != nil {
            return nil, tableName, fmt.Errorf("解析值失败 (行 %d): %w", lineNum, err)
        }

        allRecords = append(allRecords, records...)
    }

    if len(allRecords) == 0 {
        return nil, "", fmt.Errorf("未找到有效的 INSERT 语句")
    }

    return allRecords, tableName, nil
}

// shouldSkipLine 判断是否跳过该行
func (p *SQLInsertParser) shouldSkipLine(line string) bool {
    return line == "" ||
        strings.HasPrefix(line, "--") ||
        strings.HasPrefix(line, "/*") ||
        strings.HasPrefix(line, "DROP") ||
        strings.HasPrefix(line, "CREATE") ||
        strings.HasPrefix(line, "LOCK") ||
        strings.HasPrefix(line, "UNLOCK") ||
        strings.HasPrefix(line, "/*!")
}

// parseValues 解析 VALUES 部分
func (p *SQLInsertParser) parseValues(valuesStr string, columns []string) ([]dump.Record, error) {
    // 匹配所有值组: (val1,val2),(val3,val4)
    valueGroups := p.valuesRegex.FindAllStringSubmatch(valuesStr, -1)

    var records []dump.Record
    for _, group := range valueGroups {
        if len(group) < 2 {
            continue
        }

        // 解析单个值组
        values, err := p.parseValueGroup(group[1])
        if err != nil {
            return nil, err
        }

        // 构建 Record
        record := make(dump.Record)
        for i, val := range values {
            var colName string
            if i < len(columns) {
                colName = columns[i]
            } else {
                colName = fmt.Sprintf("col%d", i+1)
            }
            record[colName] = val
        }

        records = append(records, record)
    }

    return records, nil
}

// parseValueGroup 解析单个值组
func (p *SQLInsertParser) parseValueGroup(group string) ([][]byte, error) {
    var values [][]byte
    var current strings.Builder
    inString := false
    escapeNext := false

    for i := 0; i < len(group); i++ {
        ch := group[i]

        if escapeNext {
            current.WriteByte(ch)
            escapeNext = false
            continue
        }

        switch ch {
        case '\\':
            escapeNext = true
            current.WriteByte(ch)

        case '\'':
            current.WriteByte(ch)
            inString = !inString

        case ',':
            if !inString {
                values = append(values, []byte(current.String()))
                current.Reset()
            } else {
                current.WriteByte(ch)
            }

        default:
            current.WriteByte(ch)
        }
    }

    // 最后一个值
    if current.Len() > 0 || len(values) > 0 {
        // 处理 NULL
        val := strings.TrimSpace(current.String())
        if strings.ToUpper(val) == "NULL" {
            values = append(values, nil)
        } else {
            values = append(values, []byte(val))
        }
    }

    return values, nil
}

// SQL 文件格式检测
func DetectSQLFormat(sqlFile string) (format string, confidence float64, err error) {
    file, _ := os.Open(sqlFile)
    defer file.Close()

    scanner := bufio.NewScanner(file)
    lineCount := 0
    insertCount := 0
    extendedInsertCount := 0

    for scanner.Scan() && lineCount < 100 {
        lineCount++
        line := strings.ToUpper(strings.TrimSpace(scanner.Text()))

        if strings.HasPrefix(line, "INSERT INTO") {
            insertCount++
            if strings.Contains(line, "),(") {
                extendedInsertCount++
            }
        }
    }

    if insertCount == 0 {
        return "unknown", 0, nil
    }

    if extendedInsertCount > insertCount/2 {
        return "extended-insert", 0.9, nil
    }

    return "standard-insert", 0.7, nil
}
```

---

### 3. 导出辅助函数

```go
// exportRecords 导出记录到指定格式
func exportRecords(records []dump.Record, tableName string, outputFormat string) error {
    pkColumns := []string{"id"} // 尝试检测主键

    // 获取导出函数
    var exporter func([]dump.Record, ...string) string
    switch outputFormat {
    case "zip":
        exporter = _type.Zip(fmt.Sprintf("%s.zip", tableName))
    case "dir":
        exporter = _type.Dir(tableName)
    case "-":
        exporter = write.Console
    default:
        return fmt.Errorf("不支持的输出格式: %s", outputFormat)
    }

    // 执行导出
    result := exporter(records, pkColumns...)
    log.Printf("导出完成: %s", result)

    return nil
}
```

---

## 接口设计

### 命令行接口

```bash
# 基础用法
connect mysql import-sql dump.sql

# 指定输出格式
connect mysql import-sql dump.sql --in dir

# 指定数据库连接（回退方案）
connect mysql import-sql dump.sql -h localhost -P 3307 -u root -d gforge

# 保留临时数据库（调试用）
connect mysql import-sql dump.sql --keep-temp

# 指定临时数据库名
connect mysql import-sql dump.sql --temp-db my_temp

# 设置超时
connect mysql import-sql dump.sql --timeout 10m
```

### 配置文件支持

```yaml
# ~/.dumper.yaml
mysql:
  host: localhost
  port: 3307
  user: root
  password: ""

import-sql:
  output-format: zip
  timeout: 5m
  keep-temp: false
```

---

## 错误处理

### 错误类型

| 错误 | 处理方式 |
|------|---------|
| SQL 文件不存在 | 立即返回错误 |
| 解析失败 | 自动回退到 MySQL 导入 |
| MySQL 连接失败 | 返回错误，提示检查连接 |
| 表结构不一致 | 记录警告，继续处理 |
| 数据类型转换失败 | 记录错误，跳过该记录 |
| 超时 | 返回错误，提示增加超时时间 |

### 日志输出

```
[INFO] 开始导入 SQL 文件: dump.sql
[INFO] 检测 SQL 格式: extended-insert (置信度: 0.9)
[INFO] 快速解析失败: 扩展插入包含复杂转义
[INFO] 使用 MySQL 回退方案
[INFO] 创建临时数据库: temp_import_1738300000
[INFO] 导入 SQL 文件到临时数据库... (这可能需要几分钟)
[INFO] 发现 3 个表: [user, profile, logs]
[INFO] 导出表: user, 记录数: 1000
[INFO] 主键: [uid], 导出格式: zip
[INFO] 导出完成: user.zip
[INFO] 已清理临时数据库: temp_import_1738300000
[INFO] 导入完成
```

---

## 性能考虑

### 大文件处理

| 文件大小 | 预期时间 | 内存使用 |
|---------|---------|---------|
| < 10MB | < 10s | < 50MB |
| 10-100MB | 10s-2min | < 200MB |
| 100MB-1GB | 2min-10min | < 500MB |
| > 1GB | > 10min | 取决于记录数 |

### 优化策略

1. **流式处理**: 大文件逐行读取，不全部加载到内存
2. **并行导出**: 多表可以并行处理
3. **进度显示**: 显示处理进度（已处理记录数/总记录数）
4. **批处理**: 扩展插入按批次解析

---

*设计文档版本: v1.0*
*创建日期: 2025-01-31*
