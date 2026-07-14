# 🦡 Mogura

[![CI](https://github.com/cluion/Mogura/actions/workflows/ci.yml/badge.svg)](https://github.com/cluion/Mogura/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/cluion/Mogura)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**繁體中文** | [English](README.md)

像鼴鼠一樣,把 Linux 磁碟裡的垃圾挖出來。

Mogura(もぐら,日語的鼴鼠)是為 Linux 原生打造的系統清理工具。單一靜態執行檔,不需安裝任何 runtime 或函式庫;執行時只用系統標配工具(sh、coreutils),dpkg / snap / flatpak / uv 等則是有裝才用、沒裝自動略過。

介面語言跟隨系統語系:`LANG` 為 `zh` 開頭顯示繁體中文,其他顯示英文(可用 `MOGURA_LANG=en|zh` 或 `mogura config` 覆蓋)。

![mogura clean](demo/clean.gif)

![mogura analyze](demo/analyze.gif)

## 安裝

```bash
curl -fsSL https://raw.githubusercontent.com/cluion/Mogura/main/install.sh | sh
```

其他方式:

- **Debian / Ubuntu**:從 [Releases](https://github.com/cluion/Mogura/releases) 下載 `.deb` 後 `sudo dpkg -i mogura_*.deb`
- **Fedora / openSUSE**:下載 `.rpm` 後 `sudo rpm -i mogura_*.rpm`
- **Arch(AUR)**:`yay -S mogura-bin`
- **原始碼**:`CGO_ENABLED=0 go build -o mogura ./cmd/mogura`

清理規則會依發行版自動適配(apt / pacman / dnf / zypper / snap / flatpak,沒安裝的工具對應規則自動隱藏)。

## 使用

```bash
mogura              # 掃描 + 互動選擇 + 清理
mogura clean --list # 只列出可回收空間,不清理
mogura analyze [路徑] # 磁碟空間分析,互動瀏覽各目錄佔用
mogura dev [路徑]     # 掃描建置產物(node_modules、target、vendor...)
mogura orphan        # 找出已解除安裝軟體留下的孤兒設定檔
mogura monitor       # 即時系統監控(CPU、記憶體、磁碟、網路)
mogura mem           # 記憶體大戶排行;--drop-caches / --swap-reset 釋放
mogura config        # 開啟設定(語言);TUI 內也可按 , 呼出
```

- 預設先掃描、顯示每項可回收大小,勾選並確認後才會動手
- 使用者層項目(快取、垃圾桶)不需要 root;標 🔒 的系統層項目才會要求 sudo
- 數字是誠實的 `du` 口徑:實際磁碟佔用(`st_blocks`)、硬連結只計一次
- 清理規則是宣告式 YAML(`internal/rules/data/`),新增規則不用改程式碼

## 開發

```bash
CGO_ENABLED=0 go build -o mogura ./cmd/mogura
go test -race ./...
```

## 授權

[MIT](LICENSE)
