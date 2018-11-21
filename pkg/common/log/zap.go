package log

import (
	"go.uber.org/zap"
	"log"
)
var (
	Sugar *zap.SugaredLogger
	Logger *zap.Logger
)

func init() {
	var err error
	Logger, err = zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer Logger.Sync() // flushes buffer, if any
	Sugar = Logger.Sugar()
}