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
)

var (
	PersionMap map[string]os.FileInfo = make(map[string]os.FileInfo)
	osType = runtime.GOOS
	dat = "dat"
	xml = "xml"
	PersionPrefix = "Prep_"
	PersionSuffix = fmt.Sprintf(".%s", dat)
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
	if err := filepath.Walk(etc.GetSrcPath(), WalkFunc); err != nil {
		log.Logger.Error(err.Error())
	}
	return
}

func CopyFileToDst()  {
	for k, v := range PersionMap {
		splitName := v.Name()[len(PersionPrefix):strings.LastIndex(v.Name(), PersionSuffix)]
		dirWithoutP := k[:strings.LastIndex(k, GetDcmTypeFilterLeftBySystemSplit("P"))]
		log.Sugar.Debugf("k=%s, name=%s, size=%d, splitName=%s, dirWithoutP=%s", k, v.Name(), v.Size(), splitName, dirWithoutP)
		// dat已经拷贝跳过
		if file.IsFileExist(path.Join(etc.GetDstPath(), dat, splitName), fmt.Sprintf("%s.%s", splitName, dat)) {
			break
		}
		// xml已经拷贝跳过
		if file.IsFileExist(path.Join(etc.GetDstPath(), xml, splitName), fmt.Sprintf("%s.%s", splitName, xml)) {
			break
		}
		// M目录下的对应xml文件不存在跳过
		if !file.IsFileExist(path.Join(dirWithoutP, "M", splitName), fmt.Sprintf("%s.%s", splitName, xml)) {
			break
		}
		// 确保dat目录存在
		if err := EnsureDir(path.Join(etc.GetDstPath(), dat, splitName)); err != nil {
			log.Logger.Error(err.Error())
			break
		}
		// 确保xml目录存在
		if err := EnsureDir(path.Join(etc.GetDstPath(), xml, splitName)); err != nil {
			log.Logger.Error(err.Error())
			break
		}
		if err := EnsureDir(path.Join(etc.GetDstPath(), "xml")); err != nil {
			log.Logger.Error(err.Error())
			break
		}

		srcXml := path.Join(dirWithoutP, "M", splitName, fmt.Sprintf("%s.%s", splitName, xml))
		dstXml := path.Join(etc.GetDstPath(), xml, splitName, fmt.Sprintf("%s.%s", splitName, xml))
		srcDat := path.Join(k, v.Name())
		dstDat := path.Join(etc.GetDstPath(), dat, splitName, fmt.Sprintf("%s.%s", splitName, dat))
		Copy(srcXml, dstXml)
		log.Sugar.Infof("拷贝成功 %s ===> %s", srcXml, dstXml)
		Copy(srcDat, dstDat)
		log.Sugar.Infof("拷贝成功 %s ===> %s", srcDat, dstDat)
	}
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