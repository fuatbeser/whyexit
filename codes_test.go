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
				x := d.Explain(e.Code, "linux", "")
				if x.Source != tbl.ID || x.Meaning != e.Meaning {
					t.Errorf("Explain(%d) = %s %q, want %s %q", e.Code, x.Source, x.Meaning, tbl.ID, e.Meaning)
				}
			})
		}
	}
}

func TestUnknown(t *testing.T) {
	x := mustLoad(t).Explain(42, "linux", "")
	if x.Source != "unknown" || len(x.Hints) == 0 {
		t.Errorf("got %+v", x)
	}
}

func TestSignals(t *testing.T) {
	d := mustLoad(t)
	names := map[string]bool{}
	for _, s := range d.Signals {
		t.Run(s.Name, func(t *testing.T) {
			if names[s.Name] {
				t.Fatal("duplicate signal")
			}
			names[s.Name] = true
			checkEntry(t, Entry{Code: 128 + s.Linux + s.MacOS, Meaning: s.Meaning, Hints: s.Hints})
			for _, osName := range []string{"linux", "macos"} {
				n := s.number(osName)
				if n == 0 {
					continue
				}
				if n < 1 || n > 31 {
					t.Fatalf("%s number %d out of range", osName, n)
				}
				if got := d.signal(n, osName); got == nil || got.Name != s.Name {
					t.Errorf("%s: signal %d resolves to %v", osName, n, got)
				}
				x := d.Explain(128+n, osName, "")
				if x.Signal == nil || x.Signal.Name != s.Name {
					t.Errorf("%s: Explain(%d) signal = %v, want %s", osName, 128+n, x.Signal, s.Name)
				}
			}
		})
	}
}

func TestSignalNumbersDiffer(t *testing.T) {
	d := mustLoad(t)
	tests := []struct {
		code       int
		linux, mac string
	}{
		{135, "SIGBUS", "SIGEMT"},
		{138, "SIGUSR1", "SIGBUS"},
		{140, "SIGUSR2", "SIGSYS"},
		{146, "SIGCONT", "SIGTSTP"},
		{148, "SIGTSTP", "SIGCHLD"},
		{159, "SIGSYS", "SIGUSR2"},
	}
	for _, tt := range tests {
		for osName, want := range map[string]string{"linux": tt.linux, "macos": tt.mac} {
			x := d.Explain(tt.code, osName, "")
			if x.Signal == nil || x.Signal.Name != want {
				t.Errorf("%s %d: got %v, want %s", osName, tt.code, x.Signal, want)
			}
		}
	}
}

func TestContexts(t *testing.T) {
	d := mustLoad(t)
	ids := map[string]bool{}
	for _, ctx := range d.Contexts {
		if ids[ctx.ID] {
			t.Fatalf("duplicate context %s", ctx.ID)
		}
		ids[ctx.ID] = true
		if ctx.OS != "" && ctx.OS != "linux" && ctx.OS != "macos" {
			t.Errorf("%s: bad os %q", ctx.ID, ctx.OS)
		}
		seen := map[int]bool{}
		for _, e := range ctx.Codes {
			t.Run(ctx.ID+"/"+itoa(e.Code), func(t *testing.T) {
				checkEntry(t, e)
				if seen[e.Code] {
					t.Fatalf("duplicate code %d", e.Code)
				}
				seen[e.Code] = true
				for _, osName := range []string{"linux", "macos"} {
					x := d.Explain(e.Code, osName, ctx.ID)
					if x.Source != ctx.ID || x.Meaning != e.Meaning {
						t.Errorf("Explain(%d, %s) = %s %q, want %q", e.Code, ctx.ID, x.Source, x.Meaning, e.Meaning)
					}
					if e.Signal && x.Signal == nil {
						t.Errorf("code %d is marked as a signal but %d is not a %s signal", e.Code, e.Code-128, x.OS)
					}
				}
			})
		}
	}
}
