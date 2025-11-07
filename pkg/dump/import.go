package dump

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"log"
	"strings"
)

// 导入数据到指定表
func Import(db *sql.DB, database string, table string, records []Record) {
	for _, importRecords := range beforeImportRecordsCallback {
		importRecords(records)
	}
	// 1. 获取表结构信息（包含字段类型）
	columnsInfo, err := getTableColumnsWithTypes(db, database, table)
	if err != nil {
		log.Panicf("获取表结构失败: %v", err)
	}

	// 提取列名
	columns := make([]string, len(columnsInfo))
	for i, colInfo := range columnsInfo {
		columns[i] = colInfo.Name
	}

	// 2. 获取主键信息
	pkColumns, err := GetPrimaryKeyColumns(db, database, table)
	if err != nil {
		log.Panicf("获取主键失败: %v", err)
	}
	log.Println("主键", pkColumns)

	// 3. 构建插入语句
	insertSQL, _ := buildInsertSQL(table, columns)

	// 4. 准备插入语句
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		log.Panicf("准备插入语句失败: %v", err)
	}
	defer stmt.Close()

	// 5. 开启事务
	tx, err := db.Begin()
	if err != nil {
		log.Panicf("开启事务失败: %v", err)
	}

	// 6. 导入数据
	for _, record := range records {
		for _, beforeImport := range beforeImportEachRecordsCallback {
			beforeImport(record)
		}

		// 检查主键是否存在（如果主键存在则跳过）
		if hasPrimaryKey(record, pkColumns) {
			log.Println("主键定义存在")
			// 检查记录是否已存在
			exists, err := recordExists(tx, table, pkColumns, record)
			if err != nil {
				tx.Rollback()
				log.Panicf("检查记录存在失败: %v", err)
			}
			if exists {
				log.Println("数据存在")
				continue // 跳过已存在记录
			}
		}

		// 准备参数值，根据字段类型进行处理
		args := make([]interface{}, len(columns))
		for i, colInfo := range columnsInfo {
			if val, ok := record[colInfo.Name]; ok {
				// 根据字段类型处理值
				args[i] = processFieldValue(val, colInfo.Type)
			} else {
				args[i] = nil // 缺失字段设为 NULL
			}
		}

		// 执行插入
		_, err := tx.Stmt(stmt).Exec(args...)
		if err != nil {
			tx.Rollback()
			log.Panicf("插入数据失败: %s %v %v", insertSQL, args, err)
		}

		for _, afterImport := range afterImportEachRecordsCallback {
			afterImport(record)
		}
	}

	// 7. 提交事务
	if err := tx.Commit(); err != nil {
		log.Panicf("提交事务失败: %v", err)
	}
	log.Println("commit")
	for _, afterImportAll := range afterImportAllRecordsCallback {
		afterImportAll(records)
	}
}

// 构建插入SQL语句
func buildInsertSQL(table string, columns []string) (string, string) {
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	columnsClause := strings.Join(columns, ", ")
	valuesClause := strings.Join(placeholders, ", ")

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, columnsClause, valuesClause), valuesClause
}

// 检查记录是否已存在
func recordExists(tx *sql.Tx, table string, pkColumns []string, record map[string][]byte) (bool, error) {
	// 注意：这里我们假设在调用此函数之前已经获取了字段类型信息
	// 为了简化实现，我们暂时保留原有的处理方式
	// 在实际应用中，应该传递字段类型信息进来

	// 构建WHERE条件
	var conditions []string
	var args []interface{}
	for _, col := range pkColumns {
		if col == "uid" {
			b := record[col]
			var dst [8]byte
			copy(dst[:], b[:])
			uidVal := binary.LittleEndian.Uint64(dst[:8]) // 或根据实际类型转换
			args = append(args, uidVal)
		} else {
			args = append(args, record[col])
		}
		conditions = append(conditions, fmt.Sprintf("%s = ?", col))
	}

	query := fmt.Sprintf("SELECT 1 FROM %s WHERE %s", table, strings.Join(conditions, " AND "))

	var exists int
	err := tx.QueryRow(query, args...).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}

	for _, c := range args {
		if b, ok := c.([]byte); ok {
			log.Println(binary.LittleEndian.Uint64(b))
		}
	}
	log.Println(query, args)
	return true, nil
}

// processFieldValue 根据字段类型处理值
func processFieldValue(value []byte, fieldType string) interface{} {
	switch fieldType {
	case "blob", "longblob", "mediumblob", "tinyblob":
		// blob类型直接使用record对应的值
		return value
	case "int", "integer", "smallint", "mediumint", "bigint", "tinyint":
		// 数值类型对record值进行小端解码
		return decodeNumericValue(value, fieldType)
	case "timestamp", "datetime", "date", "time":
		// timestamp类型处理
		return decodeTimestampValue(value)
	case "varchar", "char", "text", "mediumtext", "longtext":
		// varchar类型处理
		return string(value)
	default:
		// 默认情况下直接使用值
		return value
	}
}

// decodeNumericValue 对数值类型进行小端解码
func decodeNumericValue(value []byte, fieldType string) interface{} {
	if len(value) == 0 {
		return nil
	}

	// 根据字段类型确定数值类型
	switch fieldType {
	case "bigint", "int", "integer":
		if len(value) >= 8 {
			return binary.LittleEndian.Uint64(value[:8])
		} else if len(value) >= 4 {
			return binary.LittleEndian.Uint32(value[:4])
		}
	case "smallint", "mediumint":
		if len(value) >= 4 {
			return binary.LittleEndian.Uint32(value[:4])
		} else if len(value) >= 2 {
			return binary.LittleEndian.Uint16(value[:2])
		}
	case "tinyint":
		if len(value) >= 1 {
			return value[0]
		}
	}

	// 如果无法解析，则返回原始值
	return value
}

// decodeTimestampValue 对时间戳类型进行解码
func decodeTimestampValue(value []byte) interface{} {
	// 对于时间戳类型，直接使用字符串表示
	if len(value) == 0 {
		return nil
	}

	// 将字节转换为字符串
	timeStr := string(value)

	// 处理ISO 8601格式的时间戳（如2025-08-19T17:00:52Z）
	// MySQL的datetime类型不接受这种格式，需要转换
	if strings.Contains(timeStr, "T") && strings.HasSuffix(timeStr, "Z") {
		// 移除末尾的'Z'并替换'T'为' '以匹配MySQL datetime格式
		timeStr = strings.Replace(timeStr, "T", " ", 1)
		timeStr = strings.TrimSuffix(timeStr, "Z")
	}

	return timeStr
}
