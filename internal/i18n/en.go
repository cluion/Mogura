package i18n

// en 是繁中原文 → 英文的對照表;新增使用者可見字串時同步補這裡
var en = map[string]string{
	// 共用
	"錯誤:":      "Error:",
	"已取消。":     "Cancelled.",
	"未知選項: %s": "unknown option: %s",
	"未知":       "unknown",
	"今天有動":     "active today",
	"閒置 %d 天":  "idle %d days",

	// usage
	`用法: mogura [指令] [選項]

無指令時開啟總覽選單(終端機環境)。

指令:
  clean      掃描並清理系統垃圾
  analyze    磁碟空間分析,互動瀏覽各目錄佔用
  dev        掃描開發專案的建置產物(node_modules、target、vendor...)
  orphan     找出已解除安裝軟體留下的孤兒設定檔
  monitor    即時系統監控(CPU、記憶體、磁碟、網路)
  mem        記憶體大戶排行;--drop-caches / --swap-reset 釋放
  config     開啟設定
  completion 輸出 shell 補全腳本(bash|zsh|fish)
  version    顯示版本

選項:
  --list         只列出結果,不進入互動清理(clean、dev、orphan)
  --json         以 JSON 輸出結果(clean、dev、orphan、mem)
  [路徑]         analyze 與 dev 的起始目錄,預設為家目錄`: `Usage: mogura [command] [options]

Run without a command to open the overview menu (in a terminal).

Commands:
  clean      scan and clean system junk
  analyze    disk usage analysis, browse interactively
  dev        scan project build artifacts (node_modules, target, vendor...)
  orphan     find config files left by uninstalled software
  monitor    live system monitor (CPU, memory, disk, network)
  mem        top memory consumers; --drop-caches / --swap-reset
  config     open settings
  completion print shell completion script (bash|zsh|fish)
  version    show version

Options:
  --list         list results only, skip interactive cleaning (clean, dev, orphan)
  --json         output results as JSON (clean, dev, orphan, mem)
  [path]         start directory for analyze and dev, defaults to home`,

	// 確認與執行
	"\n將清理以下項目:": "\nAbout to clean:",
	"預估釋放: %s\n": "Estimated space to free: %s\n",
	"部分項目需要 sudo,執行時可能要求輸入密碼。": "Some items need sudo; you may be asked for your password.",
	"確定執行?[y/N] ":          "Proceed? [y/N] ",
	"\n✨ 完成,共釋放約 %s\n":     "\n✨ Done, freed about %s\n",
	"\n✨ 完成,約 %s 已移至垃圾桶\n": "\n✨ Done, moved about %s to trash\n",
	"🗑 垃圾桶模式:項目會移至垃圾桶,可還原(action 型項目除外)。": "🗑 Trash mode: items go to the trash and can be restored (except action-based items).",
	"已掃描 %s · %s 個檔案": "scanned %s · %s files",

	// clean
	"掃描系統垃圾中...":         "Scanning system junk...",
	"Mogura — 選擇要清理的項目":  "Mogura — select items to clean",
	"未選擇任何項目,結束。":        "Nothing selected, exiting.",
	"合計可回收(可估算項目): %s\n": "Total reclaimable (measurable items): %s\n",
	"無法取得家目錄: %w":        "cannot locate home directory: %w",

	// dev
	"掃描 %s 的建置產物中...":      "Scanning %s for build artifacts...",
	"沒有找到建置產物,很乾淨!":        "No build artifacts found. Squeaky clean!",
	"Mogura — 選擇要刪除的建置產物":  "Mogura — select build artifacts to delete",
	"合計可回收: %s\n":          "Total reclaimable: %s\n",
	"dev 掃描僅支援家目錄內的路徑: %s": "dev only scans paths inside your home directory: %s",

	// orphan
	"蒐集已安裝軟體清單...":                            "Collecting installed software list...",
	"比對設定目錄中...":                              "Matching config directories...",
	"沒有找到孤兒設定,很乾淨!":                           "No orphaned configs found. Squeaky clean!",
	"dpkg 殘留設定(%d 個已移除套件)":                    "dpkg leftover configs (%d removed packages)",
	"dpkg 殘留設定(%d 個套件)":                       "dpkg leftover configs (%d packages)",
	"找不到對應的已安裝軟體 · %s":                        "no matching installed software · %s",
	"Mogura — 孤兒設定檔(啟發式判斷,刪前請確認)":             "Mogura — orphaned configs (heuristic, review before deleting)",
	"\ndpkg 殘留設定(可用 sudo dpkg --purge 清除):\n": "\ndpkg leftover configs (clean with sudo dpkg --purge):\n",
	"\n找不到對應軟體的設定目錄(啟發式,刪前請確認):":              "\nConfig directories with no matching software (heuristic, review before deleting):",
	"\n合計: %s\n": "\nTotal: %s\n",

	// mem
	"記憶體  %s / %s(可用 %s · cache %s)\n":                                   "Memory  %s / %s (available %s · cache %s)\n",
	"swap    %s / %s\n":                                                  "swap    %s / %s\n",
	"「可用」才是真實可用量,cache 由 kernel 自動回收。":                                   "\"Available\" is what you can actually use; cache is reclaimed by the kernel automatically.",
	"\n記憶體佔用前 %d 名:\n":                                                   "\nTop %d memory consumers:\n",
	"\n釋放操作(需要 sudo):mogura mem --drop-caches · mogura mem --swap-reset": "\nRelease actions (need sudo): mogura mem --drop-caches · mogura mem --swap-reset",
	"\n提醒:page cache 平常會自動回收,清除通常不必要,且會讓系統短暫變慢。":                         "\nNote: the page cache reclaims itself; dropping it is rarely necessary and slows the system briefly.",
	"\nswap 未被使用,不需要重置。":                                                 "\nSwap is not in use, no reset needed.",
	"可用記憶體不足以安全收回 swap(需要約 %s)":                                          "not enough available memory to safely reclaim swap (about %s needed)",
	"\n將把 swap 內容搬回 RAM,期間系統可能短暫變慢。":                                     "\nSwap contents will be moved back into RAM; the system may slow down briefly.",
	"清除 page cache": "Drop page cache",
	"重置 swap":       "Reset swap",
	"%s 需要 sudo,執行時可能要求輸入密碼。":          "%s needs sudo; you may be asked for your password.",
	"執行中,資料量大時可能需要數十秒...":              "Working... large data sets can take tens of seconds.",
	"\n✨ 完成(耗時 %s)\n":                  "\n✨ Done (took %s)\n",
	"  可用記憶體 %s → %s(%s)\n":            "  Available memory %s → %s (%s)\n",
	"  swap 使用   %s → %s(內容已搬回 RAM)\n": "  Swap usage   %s → %s (contents moved back to RAM)\n",

	// analyze
	"路徑不存在: %s":                             "path does not exist: %s",
	"%s 不是目錄":                               "%s is not a directory",
	"analyze 需要互動終端機":                       "analyze needs an interactive terminal",
	"monitor 需要互動終端機":                       "monitor needs an interactive terminal",
	"config 需要互動終端機":                        "config needs an interactive terminal",
	"completion 需要指定 shell:bash、zsh 或 fish": "completion needs a shell: bash, zsh or fish",
	"不支援的 shell: %s(支援 bash、zsh、fish)":      "unsupported shell: %s (bash, zsh and fish are supported)",
	"Mogura 磁碟分析":                           "Mogura Disk Analyzer",
	"  排序:":                                 "  sort: ",
	"大小":                                    "size",
	"名稱":                                    "name",
	"修改時間":                                  "modified",
	"掃描中...  已掃描 %s · %s 檔\n":               "Scanning...  %s · %s files so far\n",
	"讀取失敗: ":                                "read failed: ",
	"(空目錄)":                                 "(empty directory)",
	"backspace 返回上層 · q 離開":                 "backspace go up · q quit",
	"刪除 %s(%s)?此操作無法復原  y 確認 · 其他鍵取消": "Delete %s (%s)? This cannot be undone.  y confirm · any other key cancels",
	"刪除中...":          "Deleting...",
	"已刪除 %s,釋放 %s":    "Deleted %s, freed %s",
	"已將 %s 移至垃圾桶(%s)": "Moved %s to trash (%s)",
	"將 %s(%s)移至垃圾桶?  y 確認 · 其他鍵取消": "Move %s (%s) to trash?  y confirm · any other key cancels",
	"刪除失敗:":                     "delete failed: ",
	"已取消刪除。":                    "Deletion cancelled.",
	"計算中 %d/%d · 已掃描 %s · %s 檔": "computing %d/%d · scanned %s · %s files",
	"\nenter 進入 · backspace 上層 · s 排序 · d 刪除 · , 設定 · q 離開": "\nenter open · backspace up · s sort · d delete · , settings · q quit",
	"剛剛":          "just now",
	"%d 小時前":      "%dh ago",
	"%d 天前":       "%dd ago",
	" 檔":          " files",
	"拒絕刪除":        "deletion refused",
	"拒絕刪除第一層系統目錄": "refusing to delete a top-level system directory",
	"拒絕刪除家目錄":     "refusing to delete the home directory",

	// dashboard
	"Mogura — 總覽":                      "Mogura — Overview",
	"可回收空間 ":                           "Reclaimable ",
	"(可估算項目)":                          " (measurable items)",
	"可回收空間 掃描中... %s · %s 檔":           "Reclaimable: scanning... %s · %s files",
	"\n↑↓ 移動 · enter 進入 · , 設定 · q 離開": "\n↑↓ move · enter open · , settings · q quit",
	"清理系統垃圾":                           "Clean system junk",
	"快取、垃圾桶、套件快取與日誌":                   "caches, trash, package caches and logs",
	"磁碟空間分析":                           "Disk usage analyzer",
	"互動瀏覽各目錄佔用":                        "browse directory usage interactively",
	"開發垃圾":                             "Dev junk",
	"node_modules、target 等建置產物":        "build artifacts like node_modules and target",
	"孤兒設定檔":                            "Orphaned configs",
	"已解除安裝軟體留下的殘留設定":                   "configs left behind by uninstalled software",
	"系統監控":                             "System monitor",
	"CPU、記憶體、磁碟、網路":                    "CPU, memory, disk, network",
	"大戶排行與釋放操作":                        "top consumers and release actions",
	"設定":                               "Settings",
	"離開":                               "Quit",
	"語言、刪除方式、journal 保留":               "language, delete mode, journal retention",
	"\n按 Enter 返回總覽...":                "\nPress Enter to return to the overview...",

	// ui / settings
	"Mogura — 設定": "Mogura — Settings",
	"語言":          "Language",
	"自動(跟隨系統)":    "Auto (follow system)",
	"設定儲存失敗:":     "failed to save settings: ",
	"設定檔:%s":      "Config file: %s",
	"排除清單(exclude)等進階設定請直接編輯設定檔": "Advanced options like the exclude list are edited in the config file directly",
	"\n↑↓ 選擇 · ←→ 切換 · enter 確定": "\n↑↓ select · ←→ change · enter done",
	"刪除方式":       "Delete mode",
	"直接刪除":       "Delete directly",
	"移至垃圾桶":      "Move to trash",
	"journal 保留": "journal retention",
	"%d 天":       "%d days",
	"與垃圾桶不在同一分割區,無法移入(可在設定改回直接刪除)": "not on the same partition as the trash; cannot move (switch back to direct delete in settings)",
	"風險低":      "low risk",
	"風險中":      "med risk",
	"風險高":      "high risk",
	"已選擇可回收: ": "Selected reclaimable: ",
	"\n空白鍵 勾選 · a 全選 · n 全不選 · enter 執行 · x 排除 · , 設定 · q 離開 · 🔒 需要 sudo": "\nspace select · a all · n none · enter run · x exclude · , settings · q quit · 🔒 needs sudo",
	"此項目無法排除(非單一路徑)":          "this item can't be excluded (not a single path)",
	"已排除 %s,之後掃描不再顯示(設定檔可移除)": "Excluded %s; future scans will skip it (remove it from the config file to undo)",
	"互動介面啟動失敗: %w":            "failed to start interactive UI: %w",

	// monitor
	"取樣中...\n":    "Sampling...\n",
	"Mogura 系統監控": "Mogura System Monitor",
	"%s · 開機 %d 天 %s · 負載 %.2f %.2f %.2f": "%s · up %dd %s · load %.2f %.2f %.2f",
	"記憶體":               "Memory",
	"  %s / %s · 可用 %s": "  %s / %s · available %s",
	"磁碟":                "Disk",
	"網路":                "Network",
	"\n每 2 秒更新 · q 離開":  "\nrefreshes every 2s · q quit",

	// 清理引擎
	"其餘 %d 項":            "%d others",
	"%d 個目標刪除失敗: %s":     "%d targets failed to delete: %s",
	"非絕對路徑":              "not an absolute path",
	"路徑層級過淺,拒絕刪除":        "path too shallow, refusing to delete",
	"使用者層規則不可刪除家目錄以外的路徑": "user-level rules may not delete outside the home directory",

	"\n⚠ 以下項目無法還原,垃圾桶模式也不適用:": "\n⚠ These items cannot be restored, and trash mode does not apply:",
	"確認請完整輸入 yes: ":           "Type yes in full to confirm: ",

	// 規則(name / description)
	"Docker 建置快取": "Docker build cache",
	"建置過程留下的中介層,下次建置時會重新產生": "Intermediate layers from builds, regenerated on the next build",
	"Docker 已停止容器": "Stopped Docker containers",
	"已結束但未移除的容器,連同其可寫層一併刪除": "Exited containers not yet removed, deleted along with their writable layers",
	"Docker 未使用映像":                    "Unused Docker images",
	"沒有任何容器在用的映像,需要時得重新拉取":            "Images not used by any container; they must be pulled again when needed",
	"Docker 匿名資料卷":                    "Anonymous Docker volumes",
	"沒有容器掛載的匿名資料卷,具名資料卷不受影響":          "Anonymous volumes not mounted by any container; named volumes are untouched",
	"使用者快取":                           "User cache",
	"~/.cache 下的應用程式快取,刪除後會自動重建":      "App caches under ~/.cache, rebuilt automatically after deletion",
	"uv 套件快取":                         "uv package cache",
	"uv 的 Python 套件快取,只清除未被任何環境使用的部分": "uv's Python package cache; only prunes entries unused by any environment",
	"垃圾桶": "Trash",
	"垃圾桶中的檔案,清空後無法復原": "Files in the trash; emptying cannot be undone",
	"npm 快取": "npm cache",
	"npm 下載快取,需要時會重新下載":           "npm download cache, re-downloaded when needed",
	"Flatpak 未使用的 runtime":        "Unused Flatpak runtimes",
	"已無應用程式依賴的 Flatpak runtime":   "Flatpak runtimes no longer required by any app",
	"APT 套件快取":                    "APT package cache",
	"已下載的 .deb 安裝檔,需要時會重新下載":      "Downloaded .deb files, re-downloaded when needed",
	"Pacman 套件快取":                 "Pacman package cache",
	"已下載的套件安裝檔,保留已安裝版本,其餘清除":      "Downloaded packages; keeps installed versions, clears the rest",
	"DNF 套件快取":                    "DNF package cache",
	"dnf 的套件與中繼資料快取,需要時會重新下載":     "dnf package and metadata cache, re-downloaded when needed",
	"Zypper 套件快取":                 "Zypper package cache",
	"zypper 的套件與中繼資料快取":           "zypper package and metadata cache",
	"systemd 日誌":                  "systemd journal",
	"清除 {days} 天以前的 journal 日誌":   "Clears journal logs older than {days} days",
	"當機報告":                        "Crash reports",
	"/var/crash 中 apport 產生的當機傾印": "Apport crash dumps in /var/crash",
	"Snap 舊版本":                    "Old snap revisions",
	"snap 保留的已停用舊版本,目前使用的版本不受影響":  "Disabled old snap revisions; current versions are untouched",
}
