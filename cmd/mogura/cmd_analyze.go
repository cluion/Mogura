package main

import (
	"fmt"
	"os"

	"mogura/internal/analyze"
	"mogura/internal/rules"
)

func runAnalyze(args []string) error {
	root, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("無法取得家目錄: %w", err)
	}
	if len(args) > 0 {
		root = rules.ExpandHome(args[0])
	}
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("路徑不存在: %s", root)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s 不是目錄", root)
	}
	if !isTTY() {
		return fmt.Errorf("analyze 需要互動終端機")
	}
	return analyze.Browse(root)
}
