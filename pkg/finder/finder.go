package finder

import (
	"fmt"
	"github.com/sanguohot/dcm-timer/etc"
	"github.com/sanguohot/dcm-timer/pkg/common/file"
	"github.com/sanguohot/dcm-timer/pkg/common/log"
	"go.uber.org/zap"
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
	rawDataRecordXml = "RawdataRecord.xml"
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
		if !strings.Contains(srcPath, GetDcmTypeFilterBySystemSplit("P")) {
			return nil
		}
		if since, err := time.ParseInLocation(layout, etc.Config.Since, time.Local); err != nil {
			return err
		}else if info.ModTime().Before(since) {
			return nil
		}
		// 不是Prep_s2018102922221914708.dat形式的文件跳过
		if !strings.HasPrefix(info.Name(), PersionPrefix) || !strings.HasSuffix(info.Name(), PersionSuffix) {
			return nil
		}
		parentDir := srcPath[:strings.LastIndex(srcPath, GetSplitBySystem())]
		splitName := info.Name()[len(PersionPrefix):strings.LastIndex(info.Name(), PersionSuffix)]
		dirWithoutP := parentDir[:strings.LastIndex(parentDir, GetDcmTypeFilterLeftBySystemSplit("P"))]
		// P目录下的对应hdr文件不存在跳过
		if !file.IsFileExist(parentDir, fmt.Sprintf("%s%s.%s", PersionPrefix, splitName, hdr)) {
			log.Sugar.Warnf("%s不存在, 跳过", fmt.Sprintf("%s%s.%s", PersionPrefix, splitName, hdr))
			return nil
		}
		// M目录下的对应xml文件不存在跳过
		if !file.IsFileExist(path.Join(dirWithoutP, "M", splitName), fmt.Sprintf("%s.%s", splitName, xml)) {
			log.Sugar.Warnf("%s不存在, 跳过", fmt.Sprintf("%s.%s", splitName, xml))
			return nil
		}
		fileInfo, ok := PersionMap[parentDir]
		if !ok {
			PersionMap[parentDir] = info
		}else if fileInfo.Size() < info.Size() {
			PersionMap[parentDir] = info
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
		// 确保目录存在
		if err := EnsureDir(dstDir); err != nil {
			log.Logger.Error(err.Error())
			results <- false
			continue
		}

		srcXml := path.Join(dirWithoutP, "M", splitName, fmt.Sprintf("%s.%s", splitName, xml))
		dstXml := path.Join(dstDir, fmt.Sprintf("%s.%s", splitName, xml))
		srcRawDataRecordXml := path.Join(dirWithoutP, "M", splitName, rawDataRecordXml)
		dstRawDataRecordXml := path.Join(dstDir, rawDataRecordXml)
		srcDat := path.Join(k, v.Name())
		dstDat := path.Join(dstDir, fmt.Sprintf("%s.%s", splitName, dat))
		srcHdr := path.Join(k, fmt.Sprintf("%s%s.%s", PersionPrefix, splitName, hdr))
		dstHdr := path.Join(dstDir, fmt.Sprintf("%s.%s", splitName, hdr))
		m := make([]map[string]string, 4)
		m[0] = map[string]string{"src": srcXml, "dst": dstXml}
		m[1] = map[string]string{"src": srcRawDataRecordXml, "dst": dstRawDataRecordXml}
		m[2] = map[string]string{"src": srcDat, "dst": dstDat}
		m[3] = map[string]string{"src": srcHdr, "dst": dstHdr}
		for _, item := range m {
			if bl, err := copyWorkerCore(id, item["src"], item["dst"]); err != nil {
				goto loopFail
			}else if !bl {
				goto loopFail
			}
		}
		results <- true
		continue
		loopFail:
			results <- false
	}
}

func copyWorkerCore(id int, srcFile, dstFile string) (bool, error) {
	if !file.FilePathExist(srcFile) {
		log.Sugar.Infof("拷贝者:%d %s不存在, 跳过", id, srcFile)
		return true, nil
	}else if file.FilePathExist(dstFile) {
		log.Sugar.Debugf("拷贝者:%d %s已拷贝, 跳过", id, dstFile)
		return true, nil
	}else if size, err := file.StandardCopy(srcFile, dstFile); err != nil {
		log.Logger.Error(err.Error(), zap.String("src", srcFile), zap.String("dst", dstFile))
		return false, err
	}else {
		log.Sugar.Infof("拷贝者:%d 拷贝成功 %s ===> %s, 约 %d KB", id, srcFile, dstFile, size/1024)
		return true, nil
	}
}

func CopyFileToDst()  {
	if len(PersionMap) <= 0 {
		log.Sugar.Info("需拷贝数 ===> 0")
		return
	}
	workers := defaultWorkers
	if len(PersionMap) < defaultWorkers {
		workers = len(PersionMap)
	}
	log.Sugar.Infof("需拷贝数 ===> %d, 拷贝者数 ===> %d", len(PersionMap), workers)
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
	log.Sugar.Infof("需拷贝数 ===> %d, 已拷贝数 ===> %d", len(PersionMap), cnt)
}

func EnsureDir(dir string) error {
	if !file.FilePathExist(dir) {
		return os.MkdirAll(dir, os.ModePerm)
	}
	return nil
}

