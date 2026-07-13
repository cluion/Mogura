package orphan

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"mogura/internal/clean"
)

// Candidate 是一個找不到對應軟體的設定目錄,由使用者最終判斷是否刪除。
type Candidate struct {
	Path    string
	Name    string
	Size    int64
	ModTime time.Time
}

// skipNames 是桌面環境與系統基礎目錄,永不列為孤兒。
var skipNames = map[string]bool{
	"autostart": true, "systemd": true, "dconf": true, "pulse": true,
	"gtk-2.0": true, "gtk-3.0": true, "gtk-4.0": true, "qt5ct": true, "qt6ct": true,
	"kde": true, "kdedefaults": true, "fontconfig": true, "enchant": true,
	"environment.d": true, "menus": true, "mime": true, "autokey": true,
	"applications": true, "icons": true, "themes": true, "fonts": true,
	"trash": true, "keyrings": true, "sounds": true, "desktop-directories": true,
	"flatpak": true, "gvfs-metadata": true, "backgrounds": true, "session": true,
}

// pkgPrefixes 用於比對發行版的套件命名慣例,如 pip → python3-pip。
var pkgPrefixes = []string{"python3-", "python-", "golang-", "node-", "ruby-", "lib"}

// DefaultBases 回傳預設掃描的設定目錄。
func DefaultBases() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		filepath.Join(home, ".config"),
		filepath.Join(home, ".local", "share"),
	}
}

// ScanBases 找出 bases 下比對不到已安裝軟體的目錄。prog 可為 nil。
func ScanBases(bases []string, installed map[string]bool, prog *clean.Progress) []Candidate {
	var cands []Candidate
	for _, base := range bases {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := strings.ToLower(e.Name())
			if skipNames[name] || isActive(name, installed) {
				continue
			}
			path := filepath.Join(base, e.Name())
			cands = append(cands, Candidate{Path: path, Name: e.Name()})
		}
	}
	for i := range cands {
		cands[i].Size, cands[i].ModTime = clean.Walk(cands[i].Path, prog)
	}
	return cands
}

// isActive 採保守比對:寧可漏報孤兒,不可把活軟體判成孤兒。
func isActive(name string, installed map[string]bool) bool {
	if installed[name] {
		return true
	}
	for _, p := range pkgPrefixes {
		if installed[p+name] {
			return true
		}
	}
	if len(name) < 4 {
		return true // 名稱太短無法可靠比對,一律視為使用中
	}
	for inst := range installed {
		if len(inst) >= 4 && (strings.Contains(inst, name) || strings.Contains(name, inst)) {
			return true
		}
		// 逐 token 比對,涵蓋 BraveSoftware ↔ brave-browser 這類廠商名目錄
		for _, tok := range strings.FieldsFunc(inst, func(r rune) bool {
			return r == '-' || r == '_' || r == '.'
		}) {
			if len(tok) >= 4 && strings.Contains(name, tok) {
				return true
			}
		}
	}
	return false
}

// IdleDays 回傳距離最後修改的天數,無法取得時回傳 -1。
func (c Candidate) IdleDays() int {
	if c.ModTime.IsZero() {
		return -1
	}
	return int(time.Since(c.ModTime).Hours() / 24)
}
