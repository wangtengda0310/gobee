package dump

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// isTimeType 列是否为时间类型，需要特殊处理
func isTimeType(dataType string) bool {
	timeTypes := []string{"date", "time", "datetime", "timestamp", "year"}
	for _, t := range timeTypes {
		if dataType == t {
			return true
		}
	}
	return false
}

type Record map[string][]byte

func Dump(db *sql.DB, database, table string, columns []string, columnTypes map[string]string, where string, ids ...string) (records []Record) {
	if len(ids) == 0 {
		log.Println("dump", database, table, where)
		return
	}
	{
		result, err := db.Exec(fmt.Sprintf("use %s;", database))
		if err != nil {
			log.Panic(err)
		}
		log.Println("change database", database, result, err)
	}
	placeholders := strings.Repeat("?, ", len(ids)-1) + "?"
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	sqlQuery := fmt.Sprintf("select %s from %s where %s in (%s)", strings.Join(columns, ","), table, where, placeholders)
	query, err := db.Query(sqlQuery, args...)
	if err != nil {
		log.Panic(sqlQuery, err)
	}

	columns, err = query.Columns()
	colCount := len(columns)
	for query.Next() {
		// 为时间类型字段使用 sql.NullTime（可处理 NULL），其他使用 []byte
		values := make([]interface{}, colCount)
		scanInterfaces := make([]interface{}, colCount)
		for i, col := range columns {
			if isTimeType(columnTypes[col]) {
				var t sql.NullTime
				scanInterfaces[i] = &t
				values[i] = &t
			} else {
				var b []byte
				scanInterfaces[i] = &b
				values[i] = &b
			}
		}

		err := query.Scan(scanInterfaces...)
		if err != nil {
			log.Panic(err)
		}
		rowMap := make(map[string][]byte)
		for i, column := range columns {
			if isTimeType(columnTypes[column]) {
				// 时间类型转换为字符串
				nt := values[i].(*sql.NullTime)
				if nt.Valid {
					rowMap[column] = []byte(nt.Time.Format("2006-01-02 15:04:05"))
				} else {
					rowMap[column] = nil
				}
			} else {
				// 其他类型直接使用 []byte
				rowMap[column] = *values[i].(*[]byte)
			}
		}
		records = append(records, rowMap)
	}
	if err := query.Err(); err != nil {
		log.Panic(err)
	}

	log.Println("dump", database, table, "where", where, ids)
	return
}
