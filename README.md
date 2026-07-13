# 🦡 Mogura

像鼴鼠一樣,把 Linux 磁碟裡的垃圾挖出來。

Mogura(もぐら,日語的鼴鼠)是為 Linux 原生打造的系統清理工具。單一 Go 執行檔、無任何依賴。

## 使用

```bash
mogura              # 掃描 + 互動選擇 + 清理
mogura clean --list # 只列出可回收空間,不清理
mogura analyze [路徑] # 磁碟空間分析,互動瀏覽各目錄佔用
mogura dev [路徑]     # 掃描建置產物(node_modules、target、vendor...)
mogura orphan        # 找出已解除安裝軟體留下的孤兒設定檔
```

- 預設先掃描、顯示每項可回收大小,勾選並確認後才會動手
- 使用者層項目(快取、垃圾桶)不需要 root;標 🔒 的系統層項目才會要求 sudo
- 清理規則是宣告式 YAML(`internal/rules/data/`),新增規則不用改程式碼

## 開發

```bash
go build -o mogura ./cmd/mogura
go test ./...
```

## 路線圖

- [x] Phase 1:清理引擎 + Debian/Ubuntu 規則集 + 互動 TUI
- [x] Phase 2:磁碟分析 + 開發垃圾掃描(node_modules、target、__pycache__)
- [x] Phase 3:孤兒設定檔掃描(已移除套件 ↔ ~/.config 殘留比對)
- [ ] Phase 4:系統監控、記憶體釋放、GoReleaser 發布 + 一鍵安裝
