package utf8

import (
	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
	"golang.org/x/text/encoding/simplifiedchinese"
)

type Charset string

const (
	UTF8    Charset = "UTF-8"
	utf8    Charset = "utf-8"
	GB18030 Charset = "GB18030"
	gb18030 Charset = "gb18030"
	GBK     Charset = "GBK"
	gbk     Charset = "gbk"
)

// From 将指定编码的字节切片转换为utf8字符串
func From(other []byte, charset Charset) string {
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

// To 将utf8字节切片转换为指定编码的字符串
func To(utf []byte, charset Charset) string {
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
