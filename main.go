package main

import (
	_ "github.com/sanguohot/dcm-timer/pkg/finder"
	"os"
)

func main()  {
	done := make(chan os.Signal, 1)
	<-done
}