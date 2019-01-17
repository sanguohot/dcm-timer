package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now()
	fmt.Println(now.Unix())
	fmt.Println(now.Add(10 * time.Second).Unix())
	fmt.Println(now.Unix())
}
