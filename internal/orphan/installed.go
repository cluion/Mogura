// Package orphan 找出已解除安裝軟體留下的孤兒設定檔。
package orphan

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

// System 是孤兒比對的依據:仍存在的軟體識別名,與 dpkg 已移除但留設定的套件。
type System struct {
	Installed      map[string]bool
	RemovedConfigs []string
}

// Detect 蒐集 dpkg 套件、snap、flatpak app id 與 PATH 執行檔名(全部小寫)。
func Detect() System {
	sys := System{Installed: map[string]bool{}}

	if out, err := exec.Command("dpkg-query", "-W", "-f", "${db:Status-Abbrev} ${Package}\n").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(line)
			if len(fields) != 2 {
				continue
			}
			switch {
			case strings.HasPrefix(fields[0], "ii"):
				sys.Installed[strings.ToLower(fields[1])] = true
			case strings.HasPrefix(fields[0], "rc"):
				sys.RemovedConfigs = append(sys.RemovedConfigs, fields[1])
			}
		}
	}

	if out, err := exec.Command("snap", "list").Output(); err == nil {
		for i, line := range strings.Split(string(out), "\n") {
			if i == 0 {
				continue // 表頭
			}
			if fields := strings.Fields(line); len(fields) > 0 {
				sys.Installed[strings.ToLower(fields[0])] = true
			}
		}
	}

	if out, err := exec.Command("flatpak", "list", "--app", "--columns=application").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			id := strings.ToLower(strings.TrimSpace(line))
			if id == "" {
				continue
			}
			sys.Installed[id] = true
			if parts := strings.Split(id, "."); len(parts) > 1 {
				sys.Installed[parts[len(parts)-1]] = true // org.mozilla.firefox → firefox
			}
		}
	}

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			sys.Installed[strings.ToLower(e.Name())] = true
		}
	}

	// 正在執行的程序是最強的「活著」訊號,涵蓋 tarball/AppImage 這類套件系統外的安裝
	if procs, err := process.Processes(); err == nil {
		for _, p := range procs {
			if name, err := p.Name(); err == nil && name != "" {
				sys.Installed[strings.ToLower(name)] = true
			}
		}
	}
	return sys
}
