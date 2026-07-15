package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := Load()
	if cfg.Language != "auto" {
		t.Errorf("預設語言 = %q, 預期 auto", cfg.Language)
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := Save(Config{Language: "en"}); err != nil {
		t.Fatal(err)
	}
	if cfg := Load(); cfg.Language != "en" {
		t.Errorf("回讀語言 = %q, 預期 en", cfg.Language)
	}
	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(p) != "config.yaml" {
		t.Errorf("設定檔名 = %s", p)
	}
}

func TestLoadCorruptFileFallsBack(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if err := os.MkdirAll(filepath.Join(dir, "mogura"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mogura", "config.yaml"), []byte("{{{壞掉"), 0o644); err != nil {
		t.Fatal(err)
	}
	if cfg := Load(); cfg.Language != "auto" {
		t.Errorf("損壞設定應回退預設,實際 %q", cfg.Language)
	}
}

func TestLoadJournalDaysValidation(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if err := os.MkdirAll(filepath.Join(dir, "mogura"), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(body string) {
		if err := os.WriteFile(filepath.Join(dir, "mogura", "config.yaml"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		body string
		want int
	}{
		{"journal_days: 14", 14},
		{"journal_days: 0", 7},
		{"journal_days: -3", 7},
		{"journal_days: 9999", 7},
		{"language: en", 7},
	} {
		write(tc.body)
		if got := Load().JournalDays; got != tc.want {
			t.Errorf("%q → JournalDays = %d, 預期 %d", tc.body, got, tc.want)
		}
	}
}

func TestLoadExclude(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if err := os.MkdirAll(filepath.Join(dir, "mogura"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "exclude:\n  - ~/.cache/keep\n  - /opt/data\n"
	if err := os.WriteFile(filepath.Join(dir, "mogura", "config.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Load()
	if len(cfg.Exclude) != 2 || cfg.Exclude[0] != "~/.cache/keep" {
		t.Errorf("Exclude = %v", cfg.Exclude)
	}
}
