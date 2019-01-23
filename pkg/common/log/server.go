package log

import (
	"fmt"
	"github.com/sanguohot/dcm-timer/etc"
	"net/http"
)

// 可以通过设置Atom.SetLevel(zap.DebugLevel)动态调节日志级别
// 可以通过http接口动态设置日志级别和查看当前日志级别http://localhost:9000/level
func InitLogServer() {
	http.HandleFunc("/level", Atom.ServeHTTP)
	address := fmt.Sprintf("%s:%d", etc.GetLogHostAddress(), etc.GetLogHostPort())
	Logger.Fatal(http.ListenAndServe(address, nil).Error())
}
