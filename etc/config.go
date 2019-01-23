package etc

import (
	"github.com/spf13/viper"
	"os"
	"path"
)

// auto generate struct
// https://mholt.github.io/json-to-go/
// use mapstructure to replace json for '_' key words, e.g. rpc_port,big_data
type ConfigStruct struct {
	Source       string `json:"source"`
	Output       string `json:"output"`
	Interval     int    `json:"interval"`
	Since        string `json:"since"`
	HoldDays     int    `mapstructure:"hold_days"`
	MaxWorker    int    `mapstructure:"max_worker"`
	CopyWaitTime int    `mapstructure:"copy_wait_time"`
	Log          struct {
		Path string `json:"path"`
		Host struct {
			Address string `json:"address"`
			Port    int    `json:"port"`
		} `json:"host"`
	} `json:"log"`
}

var (
	defaultFilePath = "/etc/config.json"
	ViperConfig     *viper.Viper
	Config          *ConfigStruct
	serverPath      = os.Getenv("DCM_TIMER_PATH")
	serverType      = os.Getenv("DCM_TIMER_TYPE")
	serverTypeProd  = "production"
)

func init() {
	if serverPath == "" {
		serverPath = "./"
	}
	InitConfig(path.Join(GetServerDir(), defaultFilePath))
}
func InitConfig(filePath string) {
	ViperConfig = viper.New()
	if filePath == "" {
		ViperConfig.SetConfigFile(defaultFilePath)
	} else {
		ViperConfig.SetConfigFile(filePath)
	}

	err := ViperConfig.ReadInConfig()
	if err != nil {
		if filePath != defaultFilePath {
			panic(err)
		}
	}
	err = ViperConfig.Unmarshal(&Config)
	if err != nil {
		panic(err)
	}
}
func GetServerDir() string {
	//return GetViperConfig().GetString("server.dir")
	return serverPath
}

func GetDstPath() string {
	return path.Join(Config.Output)
}

func GetSrcPath() string {
	return path.Join(Config.Source)
}

func ServerTypeIsProd() bool {
	if serverType == serverTypeProd {
		return true
	}
	return false
}

func GetLogPath() string {
	return path.Join(GetServerDir(), Config.Log.Path)
}

func GetLogHostAddress() string {
	return Config.Log.Host.Address
}

func GetLogHostPort() int {
	return Config.Log.Host.Port
}
