package etc

import (
	"github.com/sanguohot/dcm-timer/pkg/common/log"
	"github.com/spf13/viper"
	"os"
	"path"
)
// auto generate struct
// https://mholt.github.io/json-to-go/
// use mapstructure to replace json for '_' key words, e.g. rpc_port,big_data
type ConfigStruct struct {
	Source string    `json:"source"`
	Output   string `json:"output"`
	Interval int    `json:"interval"`
}

var (
	defaultFilePath  = "/etc/config.json"
	ViperConfig *viper.Viper
	Config *ConfigStruct
	serverPath = os.Getenv("DCM_TIMER_PATH")
)

func init()  {
	if serverPath == "" {
		serverPath = "./"
		log.Sugar.Warn("DCM_TIMER_PATH env not set, use ./ as default")
	}
	log.Sugar.Infof("DCM_TIMER_PATH ===> %s", serverPath)
	InitConfig(path.Join(GetServerDir(), defaultFilePath))
}
func InitConfig(filePath string) {
	log.Sugar.Infof("config: init config path %s", filePath)
	ViperConfig = viper.New()
	if filePath == "" {
		ViperConfig.SetConfigFile(defaultFilePath)
	}else {
		ViperConfig.SetConfigFile(filePath)
	}

	err := ViperConfig.ReadInConfig()
	if err != nil {
		if filePath != defaultFilePath {
			log.Logger.Fatal(err.Error())
		}
	}
	err = ViperConfig.Unmarshal(&Config)
	if err != nil {
		log.Logger.Fatal(err.Error())
	}
}
func GetServerDir() string {
	//return GetViperConfig().GetString("server.dir")
	return serverPath
}

func GetDstPath() string {
	return path.Join(serverPath, Config.Output)
}

func GetSrcPath() string {
	return path.Join(serverPath, Config.Source)
}