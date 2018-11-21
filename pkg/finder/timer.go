package finder

import (
	"github.com/sanguohot/dcm-timer/etc"
	"time"
)

func timerTask()  {
	ShowFileList()
	CopyFileToDst()
}

func init() {
	// 一分钟写一百条
	ticks := time.NewTicker(time.Duration(etc.Config.Interval) * time.Second)
	tick := ticks.C
	go func() {
		timerTask()
		for _ = range tick {
			timerTask()
		}
	}()
}