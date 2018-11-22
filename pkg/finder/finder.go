package finder

import (
	"fmt"
	"github.com/sanguohot/dcm-timer/etc"
	"github.com/sanguohot/dcm-timer/pkg/common/file"
	"github.com/sanguohot/dcm-timer/pkg/common/log"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	PersionMap map[string]os.FileInfo = make(map[string]os.FileInfo)
	osType = runtime.GOOS
	dat = "dat"
	xml = "xml"
	PersionPrefix = "Prep_"
	PersionSuffix = fmt.Sprintf(".%s", dat)
	defaultWorkers = 10
	layout = "2006-01-02 15:04:05"
)

func GetSplitBySystem() string {
	if osType == "windows"{
		return "\\"
	}
	return "/"
}

func GetDcmTypeFilterBySystemSplit(filter string) string {
	return fmt.Sprintf("%s%s%s", GetSplitBySystem(), filter, GetSplitBySystem())
}

func GetDcmTypeFilterLeftBySystemSplit(filter string) string {
	return fmt.Sprintf("%s%s", GetSplitBySystem(), filter)
}

func WalkFunc(srcPath string, info os.FileInfo, err error) error {
	if info == nil {
		log.Sugar.Infof("找不到路径 %s", srcPath)
		return nil
	}

	if info.IsDir() {
		return nil
	} else {
		parentDir := srcPath[:strings.LastIndex(srcPath, GetSplitBySystem())]
		if !strings.Contains(srcPath, GetDcmTypeFilterBySystemSplit("P")) {
			return nil
		}
		if since, err := time.ParseInLocation(layout, etc.Config.Since, time.Local); err != nil {
			return err
		}else if info.ModTime().Before(since) {
			return nil
		}
		if strings.HasSuffix(srcPath,PersionSuffix) {
			fileInfo, ok := PersionMap[parentDir]
			if !ok {
				PersionMap[parentDir] = info
			}else if fileInfo.Size() < info.Size() {
				PersionMap[parentDir] = info
			}
		}
		return nil
	}
}

func ShowFileList() {
	log.Sugar.Infof("检索目录 ===> %s", etc.GetSrcPath())
	if err := filepath.Walk(etc.GetSrcPath(), WalkFunc); err != nil {
		log.Logger.Error(err.Error())
	}
	return
}

func CopyWorker(id int, jobs <-chan string, results chan<- bool)  {
	for j := range jobs {
		k := j
		v := PersionMap[j]
		splitName := v.Name()[len(PersionPrefix):strings.LastIndex(v.Name(), PersionSuffix)]
		dirWithoutP := k[:strings.LastIndex(k, GetDcmTypeFilterLeftBySystemSplit("P"))]
		log.Sugar.Debugf("k=%s, name=%s, size=%d, splitName=%s, dirWithoutP=%s", k, v.Name(), v.Size(), splitName, dirWithoutP)
		// dat已经拷贝跳过
		dstDir := path.Join(etc.GetDstPath(), splitName)
		if file.IsFileExist(dstDir, fmt.Sprintf("%s.%s", splitName, dat)) {
			log.Sugar.Infof("文件 %s 已拷贝, 跳过", fmt.Sprintf("%s.%s", splitName, dat))
			results <- true
			continue
		}
		// xml已经拷贝跳过
		if file.IsFileExist(dstDir, fmt.Sprintf("%s.%s", splitName, xml)) {
			log.Sugar.Infof("文件 %s 已拷贝, 跳过", fmt.Sprintf("%s.%s", splitName, xml))
			results <- true
			continue
		}
		// M目录下的对应xml文件不存在跳过
		if !file.IsFileExist(path.Join(dirWithoutP, "M", splitName), fmt.Sprintf("%s.%s", splitName, xml)) {
			results <- true
			continue
		}
		// 确保目录存在
		if err := EnsureDir(dstDir); err != nil {
			log.Logger.Error(err.Error())
			results <- true
			continue
		}

		srcXml := path.Join(dirWithoutP, "M", splitName, fmt.Sprintf("%s.%s", splitName, xml))
		dstXml := path.Join(dstDir, fmt.Sprintf("%s.%s", splitName, xml))
		srcDat := path.Join(k, v.Name())
		dstDat := path.Join(dstDir, fmt.Sprintf("%s.%s", splitName, dat))
		if err := Copy(srcXml, dstXml); err != nil {
			log.Logger.Error(err.Error())
			results <- true
			continue
		}
		log.Sugar.Infof("搬砖者:%d 拷贝成功 %s ===> %s", id, srcXml, dstXml)
		if err := Copy(srcDat, dstDat); err != nil {
			log.Logger.Error(err.Error())
			results <- true
			continue
		}
		log.Sugar.Infof("搬砖者:%d 拷贝成功 %s ===> %s", id, srcDat, dstDat)
		results <- true
	}
}

func CopyFileToDst()  {
	workers := defaultWorkers
	if len(PersionMap) < defaultWorkers {
		workers = len(PersionMap)
	}
	log.Sugar.Infof("待处理数 ===> %d, 搬砖者数 ===> %d", len(PersionMap), workers)
	jobs := make(chan string, len(PersionMap))
	results := make(chan bool, len(PersionMap))
	for w := 1; w <= workers; w++ {
		go CopyWorker(w, jobs, results)
	}
	for k, _ := range PersionMap {
		jobs <- k
	}
	close(jobs)
	// 这里收集所有结果
	for a := 1; a <= len(PersionMap); a++ {
		<-results
	}
	close(results)
}

func EnsureDir(dir string) error {
	if !file.FilePathExist(dir) {
		return os.MkdirAll(dir, os.ModePerm)
	}
	return nil
}

func Copy(src, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, input, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}