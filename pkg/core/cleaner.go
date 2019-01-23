package core

import (
	"github.com/sanguohot/dcm-timer/etc"
	"github.com/sanguohot/dcm-timer/pkg/common/log"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	prefix = "s"
)

type Cleaner struct {
	Hold time.Time
	Map  map[string]os.FileInfo
}

func NewCleaner() *Cleaner {
	if etc.Config.HoldDays <= 0 {
		log.Logger.Fatal("非法的配置项：保留天数", zap.Int("HoldDays", etc.Config.HoldDays))
	}
	t := time.Now().Add(-time.Duration(etc.Config.HoldDays) * 24 * time.Hour)
	return &Cleaner{
		Hold: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()),
		Map:  make(map[string]os.FileInfo),
	}
}

func (c *Cleaner) cleanerWalkFun(path string, info os.FileInfo, err error) error {
	if info == nil {
		log.Sugar.Infof("找不到路径 %s", path)
		return nil
	}
	if info.IsDir() && strings.HasPrefix(info.Name(), prefix) {
		year, err := strconv.ParseInt(info.Name()[len(prefix):len(prefix)+4], 10, 32)
		if err != nil {
			return err
		}
		month, err := strconv.ParseInt(info.Name()[len(prefix)+4:len(prefix)+6], 10, 32)
		if err != nil {
			return err
		}
		day, err := strconv.ParseInt(info.Name()[len(prefix)+6:len(prefix)+8], 10, 32)
		if err != nil {
			return err
		}
		da := time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.Now().Location())
		if da.Before(c.Hold) {
			c.Map[path] = info
		}
		return nil
	}
	return nil
}

func (c *Cleaner) Clean() {
	log.Sugar.Infof("目录 => %s, 清除%d天(%v)前的数据", etc.GetDstPath(), etc.Config.HoldDays, c.Hold)
	if err := filepath.Walk(etc.GetDstPath(), c.cleanerWalkFun); err != nil {
		log.Logger.Error(err.Error())
	}
	for k, _ := range c.Map {
		if err := os.RemoveAll(k); err != nil {
			log.Logger.Error(err.Error())
		}
		log.Sugar.Infof("删除目录 => %s 成功", k)
	}
	log.Sugar.Infof("目录 => %s, 清除数据完毕, 清理数量 => %d", etc.GetDstPath(), len(c.Map))
	return
}
