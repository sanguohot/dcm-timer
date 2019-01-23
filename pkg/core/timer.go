package core

import (
	"github.com/sanguohot/dcm-timer/etc"
	"github.com/sanguohot/dcm-timer/pkg/common/log"
	"time"
)

func timerTask() {
	go func() {
		for {
			exeTaskAndCalcTime("拷贝", NewFinder().FindAndCopy)
			// 任务执行完毕后，计算下一次执行的时间
			now := time.Now()
			next := now.Add(time.Duration(etc.Config.Interval) * time.Second)
			<-time.NewTimer(next.Sub(now)).C
		}
	}()
}

func exeTaskAndCalcTime(task string, f func()) {
	now := time.Now()
	f()
	log.Sugar.Infof("%s任务执行完毕, 耗时 ===> %f 秒", task, time.Since(now).Seconds())
}

func cleanTask() {
	go func() {
		for {
			exeTaskAndCalcTime("清除", NewCleaner().Clean)
			// 任务执行完毕后，计算下一次执行的时间
			now := time.Now()
			next := now.Add(time.Hour * 24)
			// 每天0点定时执行
			next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
			<-time.NewTimer(next.Sub(now)).C
		}
	}()
}

func init() {
	cleanTask()
	timerTask()
}
