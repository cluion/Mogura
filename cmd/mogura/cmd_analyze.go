package main

import (
	"errors"
	"fmt"
	"os"

	"mogura/internal/analyze"
	"mogura/internal/i18n"
	"mogura/internal/rules"
)

func runAnalyze(args []string) error {
	root, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf(i18n.T("無法取得家目錄: %w"), err)
	}
	if len(args) > 0 {
		root = rules.ExpandHome(args[0])
	}
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf(i18n.T("路徑不存在: %s"), root)
	}
	if !info.IsDir() {
		return fmt.Errorf(i18n.T("%s 不是目錄"), root)
	}
	if !isTTY() {
		return errors.New(i18n.T("analyze 需要互動終端機"))
	}
	return analyze.Browse(root)
}
