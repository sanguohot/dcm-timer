package file

import (
	"github.com/sanguohot/dcm-timer/pkg/common/log"
	"io/ioutil"
	"os"
	"path"
)

func IsFileExist(output, name string) bool {
	if _, err := os.Stat(path.Join(output, name)); err == nil {
		// path/to/whatever exists
		return true
	}
	return false
}
func SaveToLocal(output, name string, data []byte) error {
	filePath := path.Join(output, name)
	if !FilePathExist(output) {
		err := os.Mkdir(output, os.ModePerm)
		if err != nil {
			log.Logger.Fatal(err.Error())
		}
	}
	if IsFileExist(output, name) {
		// path/to/whatever exists
		return nil
	}
	return ioutil.WriteFile(filePath, data, 0644)
}

func AppendUrlToLocal(output, name string, data []byte) error {
	filePath := path.Join(output, name)
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func FilePathExist(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}