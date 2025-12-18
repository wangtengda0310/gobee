package metrics

import (
	"time"
)

type Conf struct {
	PushGateWayAddr    string        // prometheus PushGateWay 地址
	PushJobName        string        // prometheus PushGateWay job名称
	PushTickerDuration time.Duration // prometheus PushGateWay 推送间隔秒
}

func NewConf() *Conf {
	return &Conf{
		PushGateWayAddr:    "",
		PushJobName:        "pushgateway",
		PushTickerDuration: time.Second,
	}
}

var conf *Conf

// GetConfig 获取配置信息
func GetConfig() *Conf {
	if conf != nil {
		return conf
	}
	conf = NewConf()
	if conf == nil {
		return nil
	}
	return conf
}
