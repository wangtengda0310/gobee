package pkg

import (
	intern "github.com/wangtengda/gobee/lvan/exporter/internal"
	"github.com/wangtengda/gobee/lvan/exporter/pkg/logger"
	"golang.org/x/text/encoding/simplifiedchinese"
)

const (
	UTF8    intern.Charset = "UTF-8"
	utf8    intern.Charset = "utf-8"
	GB18030 intern.Charset = "GB18030"
	gb18030 intern.Charset = "gb18030"
	GBK     intern.Charset = "GBK"
	gbk     intern.Charset = "gbk"
)

// UtfFrom 将指定编码的字节切片转换为utf8字符串
func UtfFrom(other []byte, charset intern.Charset) string {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn("解码错误:%v", err)
		}
	}()
	var s []byte
	switch charset {
	case GB18030, gb18030:
		decoder := simplifiedchinese.GB18030.NewDecoder()
		var err error
		s, err = decoder.Bytes(other)
		if err != nil {
			return ""
		}
	case GBK, gbk:
		decoder := simplifiedchinese.GBK.NewDecoder()
		var err error
		s, err = decoder.Bytes(other)
		if err != nil {
			panic(err)
		}
	case UTF8, utf8:
		s = other
	default:
		s = other
	}
	return string(s)
}

// UtfTo 将utf8字节切片转换为指定编码的字符串
func UtfTo(utf []byte, charset intern.Charset) string {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn("解码错误:%v", err)
		}
	}()
	var s []byte
	switch charset {
	case GB18030, gb18030:
		encoder := simplifiedchinese.GB18030.NewEncoder()
		var err error
		s, err = encoder.Bytes(utf)
		if err != nil {
			s = utf
		}
	case GBK, gbk:
		encoder := simplifiedchinese.GBK.NewEncoder()
		var err error
		s, err = encoder.Bytes(utf)
		if err != nil {
			s = utf
		}
	case UTF8, utf8:
		s = utf
	default:
		s = utf
	}
	return string(s)
}
