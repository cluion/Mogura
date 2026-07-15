// Package trash 實作「移至垃圾桶」:優先用 gio,退回 XDG Trash 規範。
package trash

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"mogura/internal/i18n"
)

// Put 把路徑移入垃圾桶。gio 能正確處理跨分割區與各掛載點的
// 垃圾桶,所以優先;沒有 gio 時用內建的 XDG Trash 實作。
func Put(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if gio, err := exec.LookPath("gio"); err == nil {
		out, err := exec.Command(gio, "trash", "--", abs).CombinedOutput()
		if err != nil {
			return fmt.Errorf("gio trash: %s", strings.TrimSpace(string(out)))
		}
		return nil
	}
	return xdgPut(abs)
}

// xdgPut 依 XDG Trash 規範搬移:先以 O_EXCL 建立 .trashinfo 佔名,
// 再 rename 進 files/。已在垃圾桶內的路徑直接刪除,避免循環。
func xdgPut(abs string) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if abs == dir || strings.HasPrefix(abs, dir+string(os.PathSeparator)) {
		return os.RemoveAll(abs)
	}
	filesDir := filepath.Join(dir, "files")
	infoDir := filepath.Join(dir, "info")
	for _, d := range []string{filesDir, infoDir} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			return err
		}
	}

	name, infoFile, err := reserveName(infoDir, filepath.Base(abs), abs)
	if err != nil {
		return err
	}
	if err := os.Rename(abs, filepath.Join(filesDir, name)); err != nil {
		os.Remove(infoFile)
		if errors.Is(err, syscall.EXDEV) {
			return errors.New(i18n.T("與垃圾桶不在同一分割區,無法移入(可在設定改回直接刪除)"))
		}
		return err
	}
	return nil
}

// reserveName 以 O_EXCL 建立 .trashinfo 搶佔唯一名稱並寫入內容。
func reserveName(infoDir, base, abs string) (name, infoFile string, err error) {
	for i := 1; ; i++ {
		name = base
		if i > 1 {
			name = fmt.Sprintf("%s.%d", base, i)
		}
		infoFile = filepath.Join(infoDir, name+".trashinfo")
		f, err := os.OpenFile(infoFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		if err != nil {
			return "", "", err
		}
		content := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
			(&url.URL{Path: abs}).EscapedPath(), time.Now().Format("2006-01-02T15:04:05"))
		if _, err := f.WriteString(content); err != nil {
			f.Close()
			os.Remove(infoFile)
			return "", "", err
		}
		return name, infoFile, f.Close()
	}
}

// Dir 回傳家目錄垃圾桶位置(依 XDG 慣例)。
func Dir() (string, error) {
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		return filepath.Join(x, "Trash"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "Trash"), nil
}
