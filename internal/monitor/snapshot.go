// Package monitor 提供系統即時監控:CPU、記憶體、磁碟、網路
package monitor

import (
	"os"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	gnet "github.com/shirou/gopsutil/v4/net"
)

// DiskUsage 是一個掛載點的使用狀況
type DiskUsage struct {
	Mount   string
	Used    uint64
	Total   uint64
	Percent float64
}

// Snapshot 是一次系統狀態取樣
type Snapshot struct {
	Hostname string
	Uptime   time.Duration
	Load1    float64
	Load5    float64
	Load15   float64

	CPUTotal float64
	CPUCores []float64

	MemUsed      uint64
	MemTotal     uint64
	MemAvailable uint64
	MemPercent   float64
	SwapUsed     uint64
	SwapTotal    uint64

	Disks []DiskUsage

	RxRate float64 // bytes/sec,與上次取樣的差值
	TxRate float64

	rxBytes uint64
	txBytes uint64
	takenAt time.Time
}

// 只顯示真實檔案系統,略過 tmpfs、squashfs(snap)等虛擬掛載
var realFS = map[string]bool{
	"ext4": true, "ext3": true, "ext2": true, "xfs": true, "btrfs": true,
	"zfs": true, "f2fs": true, "vfat": true, "ntfs": true, "exfat": true,
}

// Take 取樣目前系統狀態;prev 用於計算網路速率(首次可傳 nil)
func Take(prev *Snapshot) Snapshot {
	s := Snapshot{takenAt: time.Now()}
	s.Hostname, _ = os.Hostname()

	if up, err := host.Uptime(); err == nil {
		s.Uptime = time.Duration(up) * time.Second
	}
	if avg, err := load.Avg(); err == nil {
		s.Load1, s.Load5, s.Load15 = avg.Load1, avg.Load5, avg.Load15
	}

	// cpu.Percent(0, ...) 回傳「距上次呼叫」的使用率,由 gopsutil 內部維護狀態
	if pcts, err := cpu.Percent(0, false); err == nil && len(pcts) > 0 {
		s.CPUTotal = pcts[0]
	}
	if cores, err := cpu.Percent(0, true); err == nil {
		s.CPUCores = cores
	}

	if vm, err := mem.VirtualMemory(); err == nil {
		s.MemUsed, s.MemTotal = vm.Used, vm.Total
		s.MemAvailable, s.MemPercent = vm.Available, vm.UsedPercent
	}
	if sw, err := mem.SwapMemory(); err == nil {
		s.SwapUsed, s.SwapTotal = sw.Used, sw.Total
	}

	if parts, err := disk.Partitions(false); err == nil {
		seen := map[string]bool{}
		for _, p := range parts {
			if !realFS[p.Fstype] || seen[p.Device] {
				continue
			}
			seen[p.Device] = true
			if u, err := disk.Usage(p.Mountpoint); err == nil && u.Total > 0 {
				s.Disks = append(s.Disks, DiskUsage{
					Mount: p.Mountpoint, Used: u.Used, Total: u.Total, Percent: u.UsedPercent,
				})
			}
		}
	}

	if counters, err := gnet.IOCounters(false); err == nil && len(counters) > 0 {
		s.rxBytes, s.txBytes = counters[0].BytesRecv, counters[0].BytesSent
		if prev != nil && !prev.takenAt.IsZero() {
			secs := s.takenAt.Sub(prev.takenAt).Seconds()
			if secs > 0 && s.rxBytes >= prev.rxBytes && s.txBytes >= prev.txBytes {
				s.RxRate = float64(s.rxBytes-prev.rxBytes) / secs
				s.TxRate = float64(s.txBytes-prev.txBytes) / secs
			}
		}
	}
	return s
}
