package pkg

import (
	intern "github.com/wangtengda/gobee/lvan/exporter/internal"
	"github.com/wangtengda/gobee/lvan/exporter/pkg/logger"
	"golang.org/x/text/encoding/simplifiedchinese"
)

var (
	completed = TaskStatus{Status: "completed", ExitCode: 0}
	failed    = TaskStatus{Status: "failed", ExitCode: 1}
	running   = TaskStatus{Status: "running", ExitCode: 2}
	blocking  = TaskStatus{Status: "blocking", ExitCode: 3}
)

const (
	UTF8    intern.Charset = "UTF-8"
	utf8    intern.Charset = "utf-8"
	GB18030 intern.Charset = "GB18030"
	gb18030 intern.Charset = "gb18030"
	GBK     intern.Charset = "GBK"
	gbk     intern.Charset = "gbk"
)

// ByteToString 将字节切片转换为指定编码的字符串
func ByteToString(byte []byte, charset intern.Charset) string {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn("解码错误:%v", err)
		}
	}()
	var str string
	switch charset {
	case GB18030, gb18030:
		decoder := simplifiedchinese.GB18030.NewDecoder()
		var err error
		str, err = decoder.String(string(byte))
		if err != nil {
			return ""
		}
	case GBK, gbk:
		decoder := simplifiedchinese.GBK.NewDecoder()
		var err error
		str, err = decoder.String(string(byte))
		if err != nil {
			panic(err)
		}
	case UTF8, utf8:
		str = string(byte)
	default:
		str = string(byte)
	}
	return str
}
