package cron

import (
	"encoding/base64"
	"github.com/robfig/cron"
	"github.com/wangtengda/gobee/lvan/pkg/batch"
	"github.com/wangtengda/gobee/lvan/pkg/logger"
	"os"
	"path/filepath"
	"time"
)

func WorkDir(workdir string) {
	dirWithCronName, err := os.ReadDir(workdir)
	if err != nil {
		logger.Warn("没有找到定时任务工作目录 %s, 30秒后重试", workdir)
		time.Sleep(30 * time.Second)
		WorkDir(workdir)
		return
	}
	c := cron.New()
	for _, dirr := range dirWithCronName {
		cronExp := dirr.Name()
		if !validateCronExpress(cronExp) {
			decodeString, err := base64.RawStdEncoding.DecodeString(cronExp)
			if err != nil {
				logger.Warn("文件名不符合cron表达式 %s， 跳过", cronExp)
				continue
			}
			cronExp = string(decodeString)
			if !validateCronExpress(cronExp) {
				logger.Warn("文件名不符合cron表达式 %s， 跳过", cronExp)
				continue
			}
		}
		if dirr.IsDir() {
			logger.Info("执行 %s 目录下的定时任务 %s %s", workdir, dirr, cronExp)
			err := c.AddFunc(cronExp, func() {
				err := batch.WithSort(filepath.Join(workdir, dirr.Name()))
				if err != nil {
					logger.Warn("定时任务执行失败 %s", err.Error())
				}
			})
			if err != nil {
				return
			}
		} else {
			err := c.AddFunc(cronExp, func() {
				logger.Warn("还不支持文件定时任务")
			})
			if err != nil {
				return
			}
		}
	}
	c.Run()
}

func validateCronExpress(exp string) bool {
	_, err := cron.ParseStandard(exp)
	return err == nil
}
