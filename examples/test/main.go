package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now().Format("20060102_15_04")
	fmt.Println(now)
}
