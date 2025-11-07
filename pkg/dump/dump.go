package dump

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type Record map[string][]byte

func Dump(db *sql.DB, database, table string, columns []string, where string, ids ...string) (records []Record) {
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
		values := make([][]byte, colCount)
		pointers := make([]interface{}, colCount)
		for i := range values {
			pointers[i] = &values[i]
		}
		err := query.Scan(pointers...)
		if err != nil {
			log.Panic(err)
		}
		rowMap := make(map[string][]byte)
		for i, column := range columns {
			rowMap[column] = values[i]
		}
		records = append(records, rowMap)
	}
	if err := query.Err(); err != nil {
		log.Panic(err)
	}

	log.Println("dump", database, table, "where", where, ids)
	return
}
