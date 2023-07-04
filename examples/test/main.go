package main

import (
	"bufio"
	"fmt"
	"k8s.io/klog/v2"
	"os"
)

func main() {
	file, err := os.OpenFile("/var/log/message1", os.O_RDONLY, 0222)
	if err != nil {
		klog.Errorf("Failed to open file. err:%s", err)
		os.Exit(1)
	}
	reader := bufio.NewScanner(file)
	for reader.Scan() {
		fmt.Println(reader.Text())
	}

}
