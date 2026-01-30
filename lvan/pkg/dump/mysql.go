package dump

import (
	"database/sql"
	"fmt"
	"log"
)

type Datasource struct {
	*sql.DB
	*Config
}

type Config struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Database string
	Table    string
}

func ConnC(c Config) func(func(Datasource)) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", c.User, c.Password, c.Host, c.Port, c.Database)
	mysqldb, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Panic(dsn, err)
	}
	log.Println("connect", dsn)

	log.Println("ping error", mysqldb.Ping())

	db := Datasource{mysqldb, &c}

	return func(f func(Datasource)) {
		defer func(db Datasource) {
			err := db.Close()
			if err != nil {
				log.Panic(err)
			}
			log.Println("db closed")
		}(db)
		f(db)
	}
}
func Conn(host, database string, port uint16, user, password string) func(func(Datasource)) {
	return ConnC(Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	})
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
func GetPrimaryKeyColumns(db *sql.DB, database, table string) ([]string, error) {
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

// GetColumnTypes 获取表中所有列的数据类型
// 返回 map[column_name]data_type，例如: {"uid": "int", "ctime": "timestamp"}
func GetColumnTypes(db *sql.DB, database, table string) map[string]string {
	query := `
		SELECT COLUMN_NAME, DATA_TYPE
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`

	rows, err := db.Query(query, database, table)
	if err != nil {
		log.Panic("获取列类型失败:", err)
	}
	defer rows.Close()

	columnTypes := make(map[string]string)
	for rows.Next() {
		var colName, dataType string
		if err := rows.Scan(&colName, &dataType); err != nil {
			log.Panic("扫描列类型失败:", err)
		}
		columnTypes[colName] = dataType
	}

	return columnTypes
}
