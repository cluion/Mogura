package main

import (
	"fmt"

	"mogura/internal/monitor"
)

func runMonitor(args []string) error {
	if len(args) > 0 {
		usage()
		return fmt.Errorf("未知選項: %s", args[0])
	}
	if !isTTY() {
		return fmt.Errorf("monitor 需要互動終端機")
	}
	return monitor.Run()
}
