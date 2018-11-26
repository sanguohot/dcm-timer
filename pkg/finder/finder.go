package finder

import (
	"fmt"
	"github.com/sanguohot/dcm-timer/etc"
	"github.com/sanguohot/dcm-timer/pkg/common/file"
	"github.com/sanguohot/dcm-timer/pkg/common/log"
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
	hdr = "hdr"
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
		if strings.HasPrefix(info.Name(), PersionPrefix) && strings.HasSuffix(info.Name(), PersionSuffix) {
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
		// P目录下的对应hdr文件不存在跳过
		if !file.IsFileExist(k, fmt.Sprintf("%s%s.%s", PersionPrefix, splitName, hdr)) {
			log.Sugar.Infof("%s不存在, 跳过", fmt.Sprintf("%s%s.%s", PersionPrefix, splitName, hdr))
			results <- false
			continue
		}
		// M目录下的对应xml文件不存在跳过
		if !file.IsFileExist(path.Join(dirWithoutP, "M", splitName), fmt.Sprintf("%s.%s", splitName, xml)) {
			log.Sugar.Infof("%s不存在, 跳过", fmt.Sprintf("%s.%s", splitName, xml))
			results <- false
			continue
		}
		// 确保目录存在
		if err := EnsureDir(dstDir); err != nil {
			log.Logger.Error(err.Error())
			results <- false
			continue
		}

		srcXml := path.Join(dirWithoutP, "M", splitName, fmt.Sprintf("%s.%s", splitName, xml))
		dstXml := path.Join(dstDir, fmt.Sprintf("%s.%s", splitName, xml))
		srcDat := path.Join(k, v.Name())
		dstDat := path.Join(dstDir, fmt.Sprintf("%s.%s", splitName, dat))
		srcHdr := path.Join(k, fmt.Sprintf("%s%s.%s", PersionPrefix, splitName, hdr))
		dstHdr := path.Join(dstDir, fmt.Sprintf("%s.%s", splitName, hdr))
		realCpCnt := 0
		// xml已经拷贝跳过
		if file.IsFileExist(dstDir, fmt.Sprintf("%s.%s", splitName, xml)) {
			log.Sugar.Infof("搬砖者:%d %s已拷贝, 跳过", id, dstXml)
		}else if size, err := file.StandardCopy(srcXml, dstXml); err != nil {
			log.Logger.Error(err.Error())
			results <- false
			continue
		}else {
			realCpCnt++
			log.Sugar.Infof("搬砖者:%d 拷贝成功 %s ===> %s, 约 %d KB", id, srcXml, dstXml, size/1024)
		}
		if file.IsFileExist(dstDir, fmt.Sprintf("%s.%s", splitName, dat)) {
			log.Sugar.Infof("搬砖者:%d %s已拷贝, 跳过", id, dstDat)
		}else if size, err := file.StandardCopy(srcDat, dstDat); err != nil {
			log.Logger.Error(err.Error())
			results <- false
			continue
		}else {
			realCpCnt++
			log.Sugar.Infof("搬砖者:%d 拷贝成功 %s ===> %s, 约 %d KB", id, srcDat, dstDat, size/1024)
		}
		if file.IsFileExist(dstDir, fmt.Sprintf("%s.%s", splitName, hdr)) {
			log.Sugar.Infof("搬砖者:%d %s已拷贝, 跳过", id, dstHdr)
		}else if size, err := file.StandardCopy(srcHdr, dstHdr); err != nil {
			log.Logger.Error(err.Error())
			results <- false
			continue
		}else {
			realCpCnt++
			log.Sugar.Infof("搬砖者:%d 拷贝成功 %s ===> %s, 约 %d KB", id, srcHdr, dstHdr, size/1024)
		}
		if realCpCnt > 0 {
			results <- true
		}else {
			results <- false
		}
	}
}

func CopyFileToDst()  {
	workers := defaultWorkers
	if len(PersionMap) < defaultWorkers {
		workers = len(PersionMap)
	}
	log.Sugar.Infof("预处理数 ===> %d, 搬砖者数 ===> %d", len(PersionMap), workers)
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
	cnt := 0
	for a := 1; a <= len(PersionMap); a++ {
		if <-results {
			cnt++
		}
	}
	close(results)
	log.Sugar.Infof("预处理数 ===> %d, 实处理数 ===> %d", len(PersionMap), cnt)
}

func EnsureDir(dir string) error {
	if !file.FilePathExist(dir) {
		return os.MkdirAll(dir, os.ModePerm)
	}
	return nil
}

