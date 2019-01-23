package main

import (
	"fmt"
	"github.com/sanguohot/dcm-timer/etc"
	"github.com/sanguohot/dcm-timer/pkg/common/log"
	"time"
)

func main() {
	now := time.Now()
	fmt.Println(now.Unix())
	fmt.Println(now.Add(10 * time.Second).Unix())
	fmt.Println(now.Unix())
	t := time.Now().Add(-time.Duration(etc.Config.HoldDays) * 24 * time.Hour)
	log.Sugar.Info(t)
	log.Sugar.Info(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()))
}
