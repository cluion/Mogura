# 🦡 Mogura

[![CI](https://github.com/cluion/Mogura/actions/workflows/ci.yml/badge.svg)](https://github.com/cluion/Mogura/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/cluion/Mogura)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

[繁體中文](README.zh-TW.md) | **English**

Dig the junk out of your Linux disk, like a mole.

Mogura (もぐら, Japanese for "mole") is an interactive disk cleaner and analyzer built natively for Linux. A single static binary with no runtime or library dependencies — at runtime it only uses standard system tools (sh, coreutils), and package managers like dpkg / snap / flatpak / uv are used when present, skipped when not.

The UI follows your locale: English by default, Traditional Chinese when `LANG` starts with `zh` (override with `MOGURA_LANG=en|zh`).

![mogura clean](demo/clean.gif)

![mogura analyze](demo/analyze.gif)

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/cluion/Mogura/main/install.sh | sh
```

Other options:

- **Debian / Ubuntu**: grab the `.deb` from [Releases](https://github.com/cluion/Mogura/releases), then `sudo dpkg -i mogura_*.deb`
- **Fedora / openSUSE**: grab the `.rpm`, then `sudo rpm -i mogura_*.rpm`
- **Arch (AUR)**: `yay -S mogura-bin`
- **From source**: `CGO_ENABLED=0 go build -o mogura ./cmd/mogura`

Cleaning rules adapt to your distro automatically (apt / pacman / dnf / zypper / snap / flatpak — rules for tools you don't have simply stay hidden).

## Usage

```bash
mogura              # scan + interactive select + clean
mogura clean --list # list reclaimable space only, clean nothing
mogura analyze [path] # disk usage analyzer, browse interactively
mogura dev [path]     # scan build artifacts (node_modules, target, vendor...)
mogura orphan        # find configs left behind by uninstalled software
mogura monitor       # live system monitor (CPU, memory, disk, network)
mogura mem           # top memory consumers; --drop-caches / --swap-reset
mogura config        # open settings; or press , inside any TUI
mogura completion bash|zsh|fish  # print shell completion script
```

Every list command also takes `--json` (stable ids, sizes in bytes, locale-independent keys) for scripting:

```bash
mogura clean --json | jq '[.[] | select(.size_known)] | map(.size_bytes) | add'
```

- Scans first and shows the size of every item; nothing is deleted until you select and confirm
- Optional trash mode in settings: deletions go to the system trash (gio trash / XDG Trash) so you can undo
- User-level items (caches, trash) never need root; items marked 🔒 ask for sudo per item
- Sizes are honest `du` semantics: real disk usage (`st_blocks`), hardlinks counted once
- Cleaning rules are declarative YAML (`internal/rules/data/`) — add a rule without touching code

## Configuration

`~/.config/mogura/config.yaml` (the first three are editable via `mogura config` or `,` inside any TUI):

```yaml
language: auto     # auto | zh | en
delete: direct     # direct | trash (deletions go to the system trash, restorable)
journal_days: 7    # journal log retention in days
exclude:           # paths skipped by clean and dev scans (~ supported)
  - ~/.cache/huggingface
```

## Development

```bash
CGO_ENABLED=0 go build -o mogura ./cmd/mogura
go test -race ./...
```

## License

[MIT](LICENSE)
