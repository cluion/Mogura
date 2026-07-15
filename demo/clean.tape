# mogura 示範:總覽選單(背景掃描) → 進清理(沿用掃描) → 勾選 → x 排除 → 回總覽
Output demo/clean.gif

Set FontSize 15
Set Width 1100
Set Height 640
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms
Set Padding 12

Env MOGURA_LANG "en"
Env XDG_CONFIG_HOME "/tmp/mogura-demo-cfg"

Type "mogura"
Sleep 500ms
Enter
Sleep 1.5s

Down@300ms 3
Sleep 800ms
Up@300ms 3
Sleep 2s

Enter
Sleep 1.5s

Down@250ms 2
Space
Sleep 800ms
Down@250ms 1
Space
Sleep 1s
Up@250ms 2
Type "x"
Sleep 2.5s

Type "q"
Sleep 1.5s
Enter
Sleep 4.5s

Type "q"
Sleep 1s
