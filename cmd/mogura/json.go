package main

import (
	"encoding/json"
	"os"
	"time"

	"mogura/internal/clean"
	"mogura/internal/devjunk"
	"mogura/internal/memory"
	"mogura/internal/orphan"
)

// JSON 輸出契約:鍵永遠是英文、大小一律 bytes、id 穩定不隨介面語言變動

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

type cleanItemJSON struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	SizeBytes   int64    `json:"size_bytes"`
	SizeKnown   bool     `json:"size_known"`
	Risk        string   `json:"risk"`
	NeedsRoot   bool     `json:"needs_root"`
	Targets     []string `json:"targets,omitempty"`
}

func printCleanJSON(results []clean.Result) error {
	items := make([]cleanItemJSON, len(results))
	for i, r := range results {
		items[i] = cleanItemJSON{
			ID: r.Rule.ID, Name: r.Rule.Name, Description: r.Rule.Description,
			SizeBytes: r.Size, SizeKnown: r.Known,
			Risk: r.Rule.Risk, NeedsRoot: r.Rule.Root, Targets: r.Targets,
		}
	}
	return printJSON(items)
}

type devItemJSON struct {
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	SizeBytes int64  `json:"size_bytes"`
	IdleDays  int    `json:"idle_days"`
	ModTime   string `json:"mtime"`
	Risk      string `json:"risk"`
}

func printDevJSON(junks []devjunk.Junk) error {
	items := make([]devItemJSON, len(junks))
	for i, j := range junks {
		items[i] = devItemJSON{
			Path: j.Path, Kind: j.Kind.Label, SizeBytes: j.Size,
			IdleDays: j.IdleDays(), ModTime: j.ModTime.Format(time.RFC3339), Risk: j.Kind.Risk,
		}
	}
	return printJSON(items)
}

type orphanCandJSON struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	IdleDays  int    `json:"idle_days"`
	ModTime   string `json:"mtime"`
}

type orphanJSON struct {
	RemovedConfigs []string         `json:"removed_configs"`
	Candidates     []orphanCandJSON `json:"candidates"`
}

func printOrphanJSON(cands []orphan.Candidate, rc []string) error {
	out := orphanJSON{RemovedConfigs: rc, Candidates: make([]orphanCandJSON, len(cands))}
	if out.RemovedConfigs == nil {
		out.RemovedConfigs = []string{}
	}
	for i, c := range cands {
		out.Candidates[i] = orphanCandJSON{
			Path: c.Path, SizeBytes: c.Size,
			IdleDays: c.IdleDays(), ModTime: c.ModTime.Format(time.RFC3339),
		}
	}
	return printJSON(out)
}

type memProcJSON struct {
	PID      int32  `json:"pid"`
	Name     string `json:"name"`
	RSSBytes uint64 `json:"rss_bytes"`
}

type memJSON struct {
	TotalBytes     uint64        `json:"total_bytes"`
	UsedBytes      uint64        `json:"used_bytes"`
	AvailableBytes uint64        `json:"available_bytes"`
	CachedBytes    uint64        `json:"cached_bytes"`
	SwapTotalBytes uint64        `json:"swap_total_bytes"`
	SwapUsedBytes  uint64        `json:"swap_used_bytes"`
	Top            []memProcJSON `json:"top"`
}

func printMemJSON(s memory.Stats, procs []memory.Proc) error {
	out := memJSON{
		TotalBytes: s.Total, UsedBytes: s.Used, AvailableBytes: s.Available,
		CachedBytes: s.Cached, SwapTotalBytes: s.SwapTotal, SwapUsedBytes: s.SwapUsed,
		Top: make([]memProcJSON, len(procs)),
	}
	for i, p := range procs {
		out.Top[i] = memProcJSON{PID: p.PID, Name: p.Name, RSSBytes: p.RSS}
	}
	return printJSON(out)
}
