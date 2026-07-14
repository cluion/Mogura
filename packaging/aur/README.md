# AUR 發布步驟(mogura-bin)

首次發布需要 AUR 帳號與已註冊的 SSH 公鑰(https://aur.archlinux.org 註冊後在帳號設定加入)。

```bash
# 1. 取得 AUR 套件 repo(首次 clone 空 repo 即代表佔名)
git clone ssh://aur@aur.archlinux.org/mogura-bin.git
cd mogura-bin

# 2. 複製本目錄的 PKGBUILD,更新 pkgver 後補上真實 checksum
updpkgsums            # 需要 pacman-contrib;會自動下載 release 並填 sha256

# 3. 產生 .SRCINFO(AUR 必要)
makepkg --printsrcinfo > .SRCINFO

# 4. 本機驗證能裝
makepkg -si

# 5. 推上 AUR
git add PKGBUILD .SRCINFO
git commit -m "update to 0.7.0"
git push
```

之後每次發版:改 `pkgver`、`updpkgsums`、重生 `.SRCINFO`、commit push 即可。
