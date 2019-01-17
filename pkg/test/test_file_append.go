package main

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sanguohot/dcm-timer/pkg/common/log"
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
	"time"
)

func ApendToFile(filePath string) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Logger.Fatal(err.Error())
	}
	n, err := file.WriteString(fmt.Sprintf("%v\n", uuid.New()))
	if err != nil {
		log.Logger.Fatal(err.Error())
	}
	log.Sugar.Debugf("append %s, size => %d, content => %s", filePath, n, uuid.New().String())
}

func CheckDirIsStillWriting(k string) bool {
	ticker := time.NewTicker(500 * time.Millisecond)
	isWriting := make(chan bool, 1)
	go func() {
		now := time.Now()
		var (
			maxDirSize int64 = 0
			curDirSize int64 = 0
		)
		for t := range ticker.C {
			log.Sugar.Debugf("%v 开始检查目录 %s", t, k)
			dirInfo, err := os.Stat(k)
			if err != nil {
				log.Logger.Error(err.Error(), zap.String("k", k))
				continue
			}
			modTime := dirInfo.ModTime()
			log.Sugar.Debugf("目录 %s 最近修改时间 %v", k, modTime)
			if modTime.After(now) {
				log.Sugar.Infof("目录 %s 有写入, 修改时间更新 %d => %d,跳过拷贝", k, now.Unix(), modTime.Unix())
				isWriting <- true
				break
			}
			curDirSize = 0
			err = filepath.Walk(k, func(path string, info os.FileInfo, err error) error {
				if info.ModTime().After(now) {
					log.Sugar.Infof("文件 %s => %s 有写入, 跳过拷贝", path, info.Name())
					return errors.New("FILE_WRITING")
				}
				curDirSize += info.Size()
				return nil
			})
			if err != nil {
				isWriting <- true
				break
			}
			if maxDirSize == 0 {
				maxDirSize = curDirSize
			}
			if maxDirSize < curDirSize {
				log.Sugar.Infof("目录 %s 有写入, 占用空间增大 %d => %d, 跳过拷贝", k, maxDirSize, curDirSize)
				isWriting <- true
				break
			}
			tenSecLater := now.Add(time.Duration(20) * time.Second)
			if tenSecLater.Before(t) {
				log.Sugar.Infof("%d秒内目录 %s 没有任何修改, 拟进行拷贝, %d vs %d", 20, k, tenSecLater.Unix(), t.Unix())
				isWriting <- false
			}
		}
	}()
	result := <-isWriting
	ticker.Stop()
	return result
}

func CheckDir(dirPath string) {
	go CheckDirIsStillWriting(dirPath)
}

func main() {
	dirPath := path.Join("/opt/syncthing-default/movie")
	CheckDir(dirPath)
	quit := make(chan os.Signal, 1)
	<-quit
	//finder.CheckDirIsStillWriting()
}
