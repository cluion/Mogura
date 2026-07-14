package clean

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"mogura/internal/i18n"
)

// ItemOutcome 是單一規則的執行結果,供上層逐項回報。
type ItemOutcome struct {
	Result Result
	Err    error
}

// Execute 依序執行選定規則的清理,回傳釋放量與逐項結果。
func Execute(selected []Result) (freed int64, outcomes []ItemOutcome) {
	for _, res := range selected {
		err := executeOne(res)
		if err == nil && res.Known {
			freed += res.Size
		}
		outcomes = append(outcomes, ItemOutcome{Result: res, Err: err})
	}
	return freed, outcomes
}

func executeOne(res Result) error {
	r := res.Rule
	if r.Action != "" {
		// Action 來自 go:embed 的內建規則,非使用者輸入;若未來支援
		// 使用者自訂規則檔,須重新審視 sh -c 的注入風險。
		cmd := exec.Command("sh", "-c", r.Action)
		if r.Root && os.Geteuid() != 0 {
			cmd = exec.Command("sudo", "sh", "-c", r.Action)
		}
		cmd.Stdin = os.Stdin // sudo 需要從終端機讀密碼
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
		}
		return nil
	}

	var failed []string
	for _, t := range res.Targets {
		if err := guardPath(t, r.Root); err != nil {
			failed = append(failed, fmt.Sprintf("%s(%s)", t, err))
			continue
		}
		if err := os.RemoveAll(t); err != nil {
			failed = append(failed, t)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf(i18n.T("%d 個目標刪除失敗: %s"), len(failed), strings.Join(failed, ", "))
	}
	return nil
}

// guardPath 是刪除前的最後防線:擋下可疑的過短路徑,
// 且非 root 規則只允許刪除家目錄內的路徑。
func guardPath(path string, root bool) error {
	if !strings.HasPrefix(path, "/") {
		return errors.New(i18n.T("非絕對路徑"))
	}
	if strings.Count(path, "/") < 3 {
		return errors.New(i18n.T("路徑層級過淺,拒絕刪除"))
	}
	if !root {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf(i18n.T("無法取得家目錄: %w"), err)
		}
		if !strings.HasPrefix(path, home+string(os.PathSeparator)) {
			return errors.New(i18n.T("使用者層規則不可刪除家目錄以外的路徑"))
		}
	}
	return nil
}
