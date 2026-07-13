package monitor

import (
	"testing"
	"time"
)

func TestTake(t *testing.T) {
	first := Take(nil)
	if first.MemTotal == 0 {
		t.Error("MemTotal 不應為 0")
	}
	if first.Hostname == "" {
		t.Error("Hostname 不應為空")
	}
	for _, d := range first.Disks {
		if d.Total == 0 {
			t.Errorf("磁碟 %s 的 Total 不應為 0", d.Mount)
		}
	}

	time.Sleep(50 * time.Millisecond)
	second := Take(&first)
	if second.RxRate < 0 || second.TxRate < 0 {
		t.Errorf("網路速率不應為負: rx=%f tx=%f", second.RxRate, second.TxRate)
	}
}
