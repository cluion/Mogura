package i18n

import "testing"

func TestT(t *testing.T) {
	SetEnglish(true)
	if got := T("已取消。"); got != "Cancelled." {
		t.Errorf("英文模式 T = %q", got)
	}
	if got := T("這串沒有翻譯"); got != "這串沒有翻譯" {
		t.Errorf("查無翻譯應原樣返回,實際 %q", got)
	}
	if got := Tf("閒置 %d 天", 5); got != "idle 5 days" {
		t.Errorf("英文 Tf = %q", got)
	}

	SetEnglish(false)
	if got := T("已取消。"); got != "已取消。" {
		t.Errorf("中文模式應短路原樣返回,實際 %q", got)
	}
	if got := Tf("閒置 %d 天", 5); got != "閒置 5 天" {
		t.Errorf("中文 Tf = %q", got)
	}
}

func TestEnTableFormatVerbsMatch(t *testing.T) {
	// 每條翻譯的格式動詞數量必須與原文一致,否則執行期會印出 %!s(MISSING)
	for zh, english := range en {
		if countVerbs(zh) != countVerbs(english) {
			t.Errorf("格式動詞數量不一致:\n  zh: %q\n  en: %q", zh, english)
		}
	}
}

func countVerbs(s string) int {
	n := 0
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '%' {
			if s[i+1] == '%' {
				i++
				continue
			}
			n++
		}
	}
	return n
}
