package memory

import "testing"

func TestRank(t *testing.T) {
	procs := []Proc{
		{PID: 1, Name: "small", RSS: 100},
		{PID: 2, Name: "big", RSS: 9000},
		{PID: 3, Name: "mid", RSS: 500},
	}
	top := Rank(procs, 2)
	if len(top) != 2 {
		t.Fatalf("長度 = %d, 預期 2", len(top))
	}
	if top[0].Name != "big" || top[1].Name != "mid" {
		t.Errorf("排序錯誤: %+v", top)
	}
}

func TestRead(t *testing.T) {
	s, err := Read()
	if err != nil {
		t.Fatal(err)
	}
	if s.Total == 0 || s.Available == 0 {
		t.Errorf("記憶體統計不應為 0: %+v", s)
	}
}
