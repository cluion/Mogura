package main

import (
	"bufio"
	"strings"
	"testing"

	"mogura/internal/clean"
	"mogura/internal/rules"
)

func result(name, risk string) clean.Result {
	return clean.Result{Rule: rules.Rule{Name: name, Risk: risk}}
}

// TestHighRiskNames 守住二次確認的觸發條件:漏判會讓不可逆的刪除
// 只隔一道 y 就執行,誤判則讓每次清理都多問一次而失去警示意義
func TestHighRiskNames(t *testing.T) {
	for _, tc := range []struct {
		name   string
		picked []clean.Result
		want   []string
	}{
		{
			name:   "沒有選任何項目",
			picked: nil,
			want:   nil,
		},
		{
			name:   "全是低風險",
			picked: []clean.Result{result("建置快取", "low"), result("npm 快取", "low")},
			want:   nil,
		},
		{
			name:   "中風險不觸發",
			picked: []clean.Result{result("未使用映像", "medium"), result("垃圾桶", "medium")},
			want:   nil,
		},
		{
			name:   "混合時只挑出高風險",
			picked: []clean.Result{result("建置快取", "low"), result("未使用資料卷", "high"), result("未使用映像", "medium")},
			want:   []string{"未使用資料卷"},
		},
		{
			name:   "多個高風險全部列出",
			picked: []clean.Result{result("資料卷", "high"), result("另一項", "high")},
			want:   []string{"資料卷", "另一項"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := highRiskNames(tc.picked)
			if len(got) != len(tc.want) {
				t.Fatalf("高風險項目 = %v, 預期 %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("第 %d 項 = %q, 預期 %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// TestConfirmHighRiskSkipsWhenNothingRisky 確保沒有高風險項目時不會多問一道,
// 否則每次清理都要輸入 yes,警示會被當成雜訊而被無視
func TestConfirmHighRiskSkipsWhenNothingRisky(t *testing.T) {
	if !confirmHighRisk([]clean.Result{result("建置快取", "low")}) {
		t.Error("沒有高風險項目時應直接放行,不該等待輸入")
	}
}

// TestConfirmHighRiskRequiresFullYes 是這道機制的重點:
// 第一道確認已經收過 y,若第二道也認 y,連按兩下就穿過去了,等於沒設防
func TestConfirmHighRiskRequiresFullYes(t *testing.T) {
	picked := []clean.Result{result("未使用資料卷", "high")}
	for _, tc := range []struct {
		input string
		want  bool
	}{
		{"yes\n", true},
		{"YES\n", true},
		{"  yes  \n", true},
		{"y\n", false},
		{"\n", false},
		{"no\n", false},
		{"yess\n", false},
	} {
		t.Run(strings.TrimSpace(tc.input), func(t *testing.T) {
			orig := stdin
			t.Cleanup(func() { stdin = orig })
			stdin = bufio.NewReader(strings.NewReader(tc.input))

			if got := confirmHighRisk(picked); got != tc.want {
				t.Errorf("輸入 %q → %v, 預期 %v", tc.input, got, tc.want)
			}
		})
	}
}
