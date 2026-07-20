// Package i18n 提供極簡在地化:繁中原文即查表鍵,
// 語系為 zh* 時整層短路,其他語系查英文表、查不到回原文
package i18n

import (
	"fmt"
	"os"
	"strings"
)

var english = detectEnglish()

func detectEnglish() bool {
	if v := os.Getenv("MOGURA_LANG"); v != "" {
		return !strings.HasPrefix(strings.ToLower(v), "zh")
	}
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(k); v != "" {
			return !strings.HasPrefix(strings.ToLower(v), "zh")
		}
	}
	return true // 無語系資訊時預設英文(國際發行)
}

// SetEnglish 強制切換語系,測試用
func SetEnglish(b bool) { english = b }

// Apply 依設定值切換語言:zh / en / auto(auto 重新依環境偵測)
func Apply(lang string) {
	switch lang {
	case "zh":
		english = false
	case "en":
		english = true
	default:
		english = detectEnglish()
	}
}

// Logo 是介面各處的品牌圖示,抽成常數讓查表鍵維持純文字,
// 換圖時不必同步修改對照表的鍵與值兩側
const Logo = "⛏️"

// logoGap 補足圖示與文字的間隔:⛏️(U+26CF + VS16)屬早期符號區,
// 字型畫成兩格寬,終端機卻多半只前進一格,後續文字會貼上圖示右緣;
// 補第二個空格拉開,照兩格前進的終端機下只是間隔略寬,不會重疊
const logoGap = "  "

// Prefix 為已翻譯的字串冠上品牌圖示
func Prefix(s string) string { return Logo + logoGap + s }

// Brand 翻譯字串並冠上品牌圖示,供標題與進度提示使用
func Brand(s string) string { return Prefix(T(s)) }

// T 翻譯一個字串;非英文語系或查無翻譯時原樣返回
func T(s string) string {
	if !english {
		return s
	}
	if t, ok := en[s]; ok {
		return t
	}
	return s
}

// Tf 翻譯格式字串後代入參數
func Tf(format string, args ...any) string {
	return fmt.Sprintf(T(format), args...)
}
