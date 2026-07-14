package main

import (
	"errors"
	"fmt"

	"mogura/internal/i18n"
	"mogura/internal/monitor"
)

func runMonitor(args []string) error {
	if len(args) > 0 {
		usage()
		return fmt.Errorf(i18n.T("未知選項: %s"), args[0])
	}
	if !isTTY() {
		return errors.New(i18n.T("monitor 需要互動終端機"))
	}
	return monitor.Run()
}
