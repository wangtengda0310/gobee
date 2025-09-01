package execute

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
	"github.com/wangtengda0310/gobee/lvan/pkg/utf8"
	"gopkg.in/yaml.v3"
)

var a atomic.Int32

type CommandMeta struct {
	Encoding  utf8.Charset `yaml:"encoding"`
	Shell     []string     `yaml:"shell"`
	Resources []string     `yaml:"resources"`
	Timeout   int          `yaml:"timeout"`
}

func CatchStderr(std io.ReadCloser, encodingFunc func([]byte) string, log func(string)) {
	// 读取标准错误
	go func() {
		scanner := bufio.NewScanner(std)
		buf := make([]byte, 1024*1024)
		scanner.Buffer(buf, cap(buf))

		for scanner.Scan() {
			var s string
			if encodingFunc != nil {
				s = encodingFunc(scanner.Bytes())
			} else {
				s = scanner.Text()
			}
			log(s)
		}

		if err := scanner.Err(); err != nil && err != io.EOF {
			logger.Error("标准错误扫描错误: %v", err)
			log(fmt.Sprintf("\n[SYSTEM ERROR] stderr扫描失败: %v\n", err))
		} else if err == io.EOF {
			logger.Info("标准错误流已结束")
		}
	}()
}

func CatchStdout(std io.ReadCloser, encodingFunc func([]byte) string, log func(string)) {
	// 读取标准输出
	go func() {
		scanner := bufio.NewScanner(std)
		buf := make([]byte, 1024*1024)
		scanner.Buffer(buf, cap(buf))

		for scanner.Scan() {
			var s string
			if encodingFunc != nil {
				s = encodingFunc(scanner.Bytes())
			} else {
				s = scanner.Text()
			}
			log(s)
		}

		if err := scanner.Err(); err != nil && err != io.EOF {
			logger.Error("标准输出扫描错误: %v", err)
			log(fmt.Sprintf("\n[SYSTEM ERROR] stdout扫描失败: %v\n", err))
		} else if err == io.EOF {
			logger.Info("标准输出流已结束")
		}
	}()
}

func TryMeta(metaFile string) *CommandMeta {
	file, err := os.ReadFile(metaFile)
	if err != nil {
		return nil
	}
	meta := &CommandMeta{}
	err = yaml.Unmarshal(file, meta)
	if err != nil {
		logger.Warn("解析 meta.yaml 失败: %v %s", err, file)
		return nil
	}
	return meta
}
