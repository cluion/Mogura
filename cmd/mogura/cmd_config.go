package main

import (
	"errors"

	"mogura/internal/i18n"
	"mogura/internal/ui"
)

func runConfig(args []string) error {
	if len(args) > 0 {
		usage()
		return errors.New(args[0])
	}
	if !isTTY() {
		return errors.New(i18n.T("config 需要互動終端機"))
	}
	return ui.RunSettings()
}
