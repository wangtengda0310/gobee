package _type

import (
	"archive/zip"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

func Zip(file string) func(records []dump.Record, pks ...string) string {
	return func(records []dump.Record, pks ...string) string {
		zipFile(records, file, pks...)
		return file
	}
}

func zipFile(records []dump.Record, file string, pks ...string) {
	// 创建 ZIP 文件
	zipFile, err := os.Create(file)
	if err != nil {
		log.Panic(err)
	}
	defer zipFile.Close()

	// 创建 ZIP writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 遍历所有记录
	for i, record := range records {
		var r any = i
		if len(pks) > 0 {
			var pkColumns []string
			for _, pk := range pks {
				pkColumns = append(pkColumns, string(record[pk]))
			}
			r = strings.Join(pkColumns, "_")
		}

		// 为每个记录创建文件夹
		recordDir := fmt.Sprintf("%v", r)

		// 为每个字段创建文件
		for column, valueByte := range record {
			// 创建文件路径
			filePath := filepath.Join(recordDir, column)

			// 创建 ZIP 中的文件
			f, err := zipWriter.Create(filePath)
			if err != nil {
				log.Panic(err)
			}

			// 写入数据
			_, err = f.Write(valueByte)
			if err != nil {
				log.Panic(err)
			}
		}
	}

	log.Println(len(records), "records zipped to file", file)
}
