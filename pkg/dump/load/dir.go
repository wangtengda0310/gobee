package load

import (
	"log"
	"os"
	"path/filepath"

	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

func Dir(dir string) []dump.Record {
	if !IsDirExist(dir) {
		panic("dir not exist: " + dir)
	}
	readDir, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	var result []dump.Record
	func() {
		log.Println(readDir)
		var record = make(dump.Record)
		for _, entry := range readDir {
			if !entry.IsDir() {
				file, err := os.ReadFile(filepath.Join(dir, entry.Name()))
				if err != nil {
					panic(err)
				}
				record[entry.Name()] = file
			}
		}
		result = append(result, record)
	}()
	return result
}
func IsDirExist(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false // 路径不存在
		}
		return false // 其他错误（如权限问题）
	}
	return info.IsDir() // 存在且是目录
}
