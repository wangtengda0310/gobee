package cron

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron"
	"github.com/wangtengda0310/gobee/lvan/pkg/batch"
	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
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
				logger.Error("添加定时任务失败 [%s]: %v", cronExp, err)
				continue // 修改此处从 return 改为 continue
			}
		} else {
			err := c.AddFunc(cronExp, func() {
				logger.Warn("还不支持文件定时任务")
			})
			if err != nil {
				logger.Error("添加文件定时任务失败 [%s]: %v", cronExp, err)
				continue // 修改此处从 return 改为 continue
			}
		}
	}
	c.Run()
}

func validateCronExpress(exp string) bool {
	_, err := cron.Parse(exp)
	return err == nil
}
