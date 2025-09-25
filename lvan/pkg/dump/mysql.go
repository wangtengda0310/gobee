package dump

import (
	"database/sql"
	"fmt"
	"log"
)

func Conn(host, database string, port uint16, user, password string) func(func(*sql.DB)) {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, host, port, database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Panic(dsn, err)
	}
	log.Println("connect", dsn)
	return func(f func(*sql.DB)) {
		defer func(db *sql.DB) {
			err := db.Close()
			if err != nil {
				log.Panic(err)
			}
			log.Println("db closed")
		}(db)
		f(db)
	}
}

func Columns(db *sql.DB, database, table string) (result []string) {
	query, err := db.Query(fmt.Sprintf(`
-- 查询表的所有字段信息
SELECT COLUMN_NAME
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_NAME = '%s'
  AND TABLE_SCHEMA = '%s'; -- 可选，指定数据库名
`, table, database))
	if err != nil {
		panic(err)
	}

	for query.Next() {
		var column string
		err := query.Scan(&column)
		if err != nil {
			panic(err)
		}
		result = append(result, column)
	}
	return
}

// ColumnInfo 表示数据库列的信息
type ColumnInfo struct {
	Name string
	Type string
}

// 获取表的所有列信息
func getTableColumnsWithTypes(db *sql.DB, database, table string) ([]ColumnInfo, error) {
	query := `
		SELECT COLUMN_NAME, DATA_TYPE
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`

	rows, err := db.Query(query, database, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Name, &col.Type); err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}

	return columns, nil
}

// 获取表的所有列名
func getTableColumns(db *sql.DB, database, table string) ([]string, error) {
	columnsInfo, err := getTableColumnsWithTypes(db, database, table)
	if err != nil {
		return nil, err
	}

	var columns []string
	for _, col := range columnsInfo {
		columns = append(columns, col.Name)
	}

	return columns, nil
}

// 获取表的主键列
func PetPrimaryKeyColumns(db *sql.DB, database, table string) ([]string, error) {
	query := `
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = ?
		  AND TABLE_NAME = ?
		  AND CONSTRAINT_NAME = 'PRIMARY'
		ORDER BY ORDINAL_POSITION`

	rows, err := db.Query(query, database, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pkColumns []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, err
		}
		pkColumns = append(pkColumns, col)
	}

	return pkColumns, nil
}

// 检查记录中是否包含主键值
func hasPrimaryKey(record map[string][]byte, pkColumns []string) bool {
	for _, pk := range pkColumns {
		if _, ok := record[pk]; !ok || record[pk] == nil {
			return false
		}
	}
	return true
}
