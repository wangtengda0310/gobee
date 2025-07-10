package output

import (
	"log"
	"os"
	"strings"
	"sync"
)

var (
	DefaultWriter writeFunc = newFileWriter("log.txt", 1)
	mu            sync.RWMutex
)

// SetDefaultWriter 允许外部设置默认写入器
func SetDefaultWriter(factory func() func([]string)) {
	mu.Lock()
	defer mu.Unlock()
	DefaultWriter = factory()
}

// newFileWriter 仅包内可见，支持batch写入
func newFileWriter(filePath string, batchSize int) writeFunc {
	return func(contents []string) {
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println("open file error:", err)
			return
		}
		defer f.Close()
		for i := 0; i < len(contents); i += batchSize {
			end := i + batchSize
			if end > len(contents) {
				end = len(contents)
			}
			batch := contents[i:end]
			f.WriteString(strings.Join(batch, "\n") + "\n")
		}
	}
}
