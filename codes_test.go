package main

import (
	"strings"
	"testing"
)

func mustLoad(t *testing.T) *Data {
	t.Helper()
	d, err := loadData()
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func checkEntry(t *testing.T, e Entry) {
	t.Helper()
	if e.Code < 0 || e.Code > 255 {
		t.Errorf("code %d out of range 0-255", e.Code)
	}
	if strings.TrimSpace(e.Meaning) == "" {
		t.Error("empty meaning")
	}
	if n := len(e.Hints); n < 1 || n > 2 {
		t.Errorf("want 1 or 2 hints, got %d", n)
	}
	for _, h := range e.Hints {
		if strings.TrimSpace(h) == "" {
			t.Error("empty hint")
		}
	}
}

func TestTables(t *testing.T) {
	d := mustLoad(t)
	for _, tbl := range d.Tables {
		seen := map[int]bool{}
		for _, e := range tbl.Codes {
			t.Run(tbl.ID+"/"+itoa(e.Code), func(t *testing.T) {
				checkEntry(t, e)
				if seen[e.Code] {
					t.Fatalf("duplicate code %d", e.Code)
				}
				seen[e.Code] = true
				x := d.Explain(e.Code)
				if x.Source != tbl.ID || x.Meaning != e.Meaning {
					t.Errorf("Explain(%d) = %s %q, want %s %q", e.Code, x.Source, x.Meaning, tbl.ID, e.Meaning)
				}
			})
		}
	}
}

func TestUnknown(t *testing.T) {
	x := mustLoad(t).Explain(42)
	if x.Source != "unknown" || len(x.Hints) == 0 {
		t.Errorf("got %+v", x)
	}
}
