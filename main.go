package main

import (
	_ "github.com/CodyGuo/godaemon"
	_ "github.com/sanguohot/dcm-timer/pkg/core"
	"os"
)

func main() {
	done := make(chan os.Signal, 1)
	<-done
}
