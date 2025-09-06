package _type

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

func Dir(dir string) func(records []dump.Record, pks ...string) string {
	return func(records []dump.Record, pks ...string) string {

		for i, record := range records {
			var r any = i
			if len(pks) > 0 {
				var pkColumns []string
				for _, pk := range pks {
					pkColumns = append(pkColumns, string(record[pk]))
				}
				r = strings.Join(pkColumns, "_")
			}
			dir := filepath.Join(dir, fmt.Sprintf("%v", r))
			err := os.MkdirAll(dir, 0777)
			if err != nil {
				log.Panic(err)
			}
			for column, valueByte := range record {
				columnFile := filepath.Join(dir, column)
				err := os.WriteFile(columnFile, valueByte, 0777)
				if err != nil {
					log.Panic(err)
				}
			}

		}
		return dir
	}
}
