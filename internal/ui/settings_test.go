package ui

import (
	"strings"
	"testing"

	"mogura/internal/config"
	"mogura/internal/i18n"
)

// pressSettings 依序送入按鍵並回傳最後的面板狀態
func pressSettings(s Settings, keys ...string) Settings {
	for _, k := range keys {
		s, _ = s.HandleKey(keyMsg(k))
	}
	return s
}

// settingsAt 開一個游標停在指定列的面板;0 語言、1 刪除方式、2 journal 保留
func settingsAt(row int) Settings {
	s := NewSettings()
	s.row = row
	return s
}

func TestSettingsCloseKeys(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, key := range []string{"enter", "q", "esc", ",", "ctrl+c"} {
		t.Run(key, func(t *testing.T) {
			if _, closed := NewSettings().HandleKey(keyMsg(key)); !closed {
				t.Errorf("%s 應關閉面板", key)
			}
		})
	}
}

func TestSettingsNavigationWraps(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, tc := range []struct {
		name string
		keys []string
		want int
	}{
		{"下移一列", []string{"down"}, 1},
		{"末列下移繞回首列", []string{"down", "down", "down"}, 0},
		{"首列上移繞到末列", []string{"up"}, settingsRows - 1},
		{"vim 鍵同樣有效", []string{"j", "j"}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := pressSettings(NewSettings(), tc.keys...).row; got != tc.want {
				t.Errorf("列 = %d, 預期 %d", got, tc.want)
			}
		})
	}
}

func TestCycleLanguage(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Cleanup(func() { i18n.SetEnglish(false) }) // cycle 會即時套用語系,還原以免污染其他測試

	s := settingsAt(0)
	for _, want := range []string{"zh", "en", "auto"} {
		if s = pressSettings(s, "right"); s.cfg.Language != want {
			t.Fatalf("向右循環語言 = %q, 預期 %q", s.cfg.Language, want)
		}
	}
	if s = pressSettings(s, "left"); s.cfg.Language != "en" {
		t.Errorf("向左應反向循環,實際 %q", s.cfg.Language)
	}
}

func TestCycleDeleteMode(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := settingsAt(1)

	if s = pressSettings(s, "right"); !s.cfg.UseTrash() {
		t.Errorf("刪除方式應切到垃圾桶,實際 %q", s.cfg.Delete)
	}
	if s = pressSettings(s, "right"); s.cfg.UseTrash() {
		t.Errorf("再切一次應回到直接刪除,實際 %q", s.cfg.Delete)
	}
}

func TestCycleJournalDays(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := settingsAt(2)

	for _, want := range []int{14, 30, 3, 7} {
		if s = pressSettings(s, "right"); s.cfg.JournalDays != want {
			t.Fatalf("向右循環天數 = %d, 預期 %d", s.cfg.JournalDays, want)
		}
	}
	if s = pressSettings(s, "left"); s.cfg.JournalDays != 3 {
		t.Errorf("向左應反向循環,實際 %d", s.cfg.JournalDays)
	}
}

func TestSpaceCyclesForward(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if s := pressSettings(settingsAt(1), " "); !s.cfg.UseTrash() {
		t.Error("空白鍵應與向右同義,方便單手操作")
	}
}

func TestSettingsSaveImmediately(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := pressSettings(settingsAt(1), "right")

	if s.saveErr != nil {
		t.Fatalf("存檔失敗: %v", s.saveErr)
	}
	if !config.Load().UseTrash() {
		t.Error("面板變更應立刻寫入設定檔,不必等關閉")
	}
}

// TestChangedOnlyForDisplayAffectingSettings 守住 Changed 的分寸:
// 它決定宿主要不要整個重建清單,誤報會讓使用者的勾選白白消失
func TestChangedOnlyForDisplayAffectingSettings(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Cleanup(func() { i18n.SetEnglish(false) })

	if NewSettings().Changed() {
		t.Error("沒動過任何設定不該回報有變")
	}
	if s := pressSettings(settingsAt(1), "right"); s.Changed() {
		t.Error("刪除方式不影響已顯示的規則文字,不該觸發重建")
	}
	if s := pressSettings(settingsAt(0), "right"); !s.Changed() {
		t.Error("語言變了,規則名稱與說明都要重譯,應觸發重建")
	}
	if s := pressSettings(settingsAt(2), "right"); !s.Changed() {
		t.Error("journal 天數會出現在規則說明的 {days},應觸發重建")
	}
}

func TestCycleValueUnknownStartsAtFirst(t *testing.T) {
	for _, tc := range []struct {
		current string
		delta   int
		want    string
	}{
		{"auto", 1, "zh"},
		{"auto", -1, "en"},
		{"不存在的值", 1, "zh"}, // 找不到時視為索引 0
		{"不存在的值", -1, "en"},
	} {
		if got := cycleValue(langValues, tc.current, tc.delta); got != tc.want {
			t.Errorf("cycleValue(%q, %d) = %q, 預期 %q", tc.current, tc.delta, got, tc.want)
		}
	}
}

func TestSettingsViewRendersAllRows(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	out := NewSettings().View()
	for _, want := range []string{"設定", "語言", "刪除方式", "journal 保留", "7 天", "設定檔:"} {
		if !strings.Contains(out, want) {
			t.Errorf("畫面應包含 %q", want)
		}
	}
}

// TestCycleJournalDaysFromCustomValue 涵蓋設定檔被手動改成清單外天數的情況
//
// journalDayValues 是 [3 7 14 30]。使用者可以在 config.yaml 寫任何合法天數,
// 此時循環以大小接續而非索引位移:向右一定變大、向左一定變小,
// 兩端才繞回。使用者按下的方向永遠等於數值變動的方向
func TestCycleJournalDaysFromCustomValue(t *testing.T) {
	for _, tc := range []struct {
		name        string
		custom      int
		right, left int
	}{
		{"落在 7 與 14 之間", 10, 14, 7},
		{"落在 14 與 30 之間", 25, 30, 14},
		{"小於最小值,向左繞到最大", 1, 3, 30},
		{"大於最大值,向右繞到最小", 100, 3, 30},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			if err := config.Save(config.Config{Language: "auto", Delete: "direct", JournalDays: tc.custom}); err != nil {
				t.Fatal(err)
			}
			s := settingsAt(2)
			if s.cfg.JournalDays != tc.custom {
				t.Fatalf("前置條件失敗:自訂天數應被讀入,實際 %d", s.cfg.JournalDays)
			}

			if got := pressSettings(s, "right").cfg.JournalDays; got != tc.right {
				t.Errorf("%d 按右 = %d, 預期下一個更大的 %d", tc.custom, got, tc.right)
			}
			if got := pressSettings(s, "left").cfg.JournalDays; got != tc.left {
				t.Errorf("%d 按左 = %d, 預期下一個更小的 %d", tc.custom, got, tc.left)
			}
		})
	}
}
