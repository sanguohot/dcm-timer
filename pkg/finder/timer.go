package finder

import (
	"github.com/sanguohot/dcm-timer/etc"
	"github.com/sanguohot/dcm-timer/pkg/common/log"
	"time"
)

func timerTask()  {
	now := time.Now()
	ShowFileList()
	CopyFileToDst()
	log.Sugar.Infof("任务执行完毕, 耗时 ===> %f 秒", time.Since(now).Seconds())
}

func init() {
	timerTask()
	// 一分钟写一百条
	ticks := time.NewTicker(time.Duration(etc.Config.Interval) * time.Second)
	tick := ticks.C
	go func() {
		for _ = range tick {
			timerTask()
		}
	}()
}