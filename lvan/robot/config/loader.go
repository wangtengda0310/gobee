package main

import (
	"flag"
	"fmt"
)

var all = []loader{

{"log",cLogger.NewConf},

{ "robot", loadConfig},

}

type loader struct {

	key string

	loaderFunc func() cConfig.Conf

}

func loadConfig() cconfig.Conf { return GetConfig()}
func InitConfig() error {
	initParse()
	flag.Parse()

	var chain lconfig.SourceInheritanceChain
	fileSource, err := file.NewSource(file.WithPaths(confFile), file.WithTryLoad(b: true))
	if err != nil {
		return err
	}
	if fileSource != nil {
		chain = chain.Add(fileSource)
	}

	config, err := lconfig.NewConfig(chain)
	if err != nil {
		return fmt.Errorf("init config failed, error:%w", err)
	}

	for _, l := range all {
		c := l.loaderFunc()

		// 数据加载
		if err = config.ScanKey(l.key, c); err != nil {
			return fmt.Errorf("d2Config Scan failed, error:%v", err)
		}

		// 加载后处理
		if err = c.Process(); err != nil {
			return fmt.Errorf("process config failed, key:%v error:%v", l.key, err)
		}
		logger.Info("load config", "key", l.key,"config", c)
	}

	return nil
}
