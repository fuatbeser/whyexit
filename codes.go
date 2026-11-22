package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"runtime"
)

//go:embed codes.json
var codesJSON []byte

// Entry is one exit code in a table.
type Entry struct {
	Code    int      `json:"code"`
	Name    string   `json:"name,omitempty"`
	Meaning string   `json:"meaning"`
	Hints   []string `json:"hints"`
	// Signal marks codes that follow the 128+N convention, so the
	// signal name is shown next to the table's own meaning.
	Signal bool `json:"signal,omitempty"`
}

// Table groups the codes of one source, such as the shell or a tool.
type Table struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	// OS fixes the signal numbering, e.g. containers always run Linux.
	OS    string  `json:"os,omitempty"`
	Codes []Entry `json:"codes"`
}

// Signal describes one signal and its number on each platform.
// A zero number means the platform does not have it.
type Signal struct {
	Name    string   `json:"name"`
	Linux   int      `json:"linux,omitempty"`
	MacOS   int      `json:"macos,omitempty"`
	Meaning string   `json:"meaning"`
	Hints   []string `json:"hints"`
}

// Data is the parsed contents of codes.json.
type Data struct {
	Tables   []Table  `json:"tables"`
	Contexts []Table  `json:"contexts"`
	Signals  []Signal `json:"signals"`
}

// SignalRef names the signal behind a 128+N code.
type SignalRef struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
}

// Explanation is the answer for a single exit code.
type Explanation struct {
	Code    int        `json:"code"`
	Source  string     `json:"source"`
	Context string     `json:"context,omitempty"`
	OS      string     `json:"os"`
	Name    string     `json:"name,omitempty"`
	Signal  *SignalRef `json:"signal,omitempty"`
	Meaning string     `json:"meaning"`
	Hints   []string   `json:"hints"`
	Note    string     `json:"note,omitempty"`
}

func loadData() (*Data, error) {
	var d Data
	if err := json.Unmarshal(codesJSON, &d); err != nil {
		return nil, fmt.Errorf("codes.json: %w", err)
	}
	return &d, nil
}

// defaultOS picks the signal numbering for the running system.
// The BSDs share macOS numbering; everything else uses Linux's.
func defaultOS() string {
	switch runtime.GOOS {
	case "darwin", "ios", "freebsd", "openbsd", "netbsd", "dragonfly":
		return "macos"
	}
	return "linux"
}

func (s *Signal) number(osName string) int {
	if osName == "macos" {
		return s.MacOS
	}
	return s.Linux
}

func (d *Data) signal(n int, osName string) *Signal {
	for i := range d.Signals {
		if n > 0 && d.Signals[i].number(osName) == n {
			return &d.Signals[i]
		}
	}
	return nil
}

func (t *Table) find(code int) *Entry {
	for i := range t.Codes {
		if t.Codes[i].Code == code {
			return &t.Codes[i]
		}
	}
	return nil
}

// Context returns the context table with the given id, or nil.
func (d *Data) Context(id string) *Table {
	for i := range d.Contexts {
		if d.Contexts[i].ID == id {
			return &d.Contexts[i]
		}
	}
	return nil
}

// ContextIDs lists the valid --context values.
func (d *Data) ContextIDs() []string {
	ids := make([]string, len(d.Contexts))
	for i, t := range d.Contexts {
		ids[i] = t.ID
	}
	return ids
}

// Explain looks code up in the context table (if any), then in the
// general tables, then as 128+N.
func (d *Data) Explain(code int, osName, context string) Explanation {
	tables := d.Tables
	if ctx := d.Context(context); ctx != nil {
		tables = append([]Table{*ctx}, tables...)
		if ctx.OS != "" {
			osName = ctx.OS
		}
	}
	x := Explanation{Code: code, Context: context, OS: osName}
	sig := d.signal(code-128, osName)
	for _, t := range tables {
		if e := t.find(code); e != nil {
			x.Source, x.Name, x.Meaning, x.Hints = t.ID, e.Name, e.Meaning, e.Hints
			if e.Signal && sig != nil {
				x.Signal = &SignalRef{Number: code - 128, Name: sig.Name}
			}
			return x
		}
	}
	if sig != nil {
		x.Source, x.Meaning, x.Hints = "signal", sig.Meaning, sig.Hints
		x.Signal = &SignalRef{Number: code - 128, Name: sig.Name}
		return x
	}
	x.Source = "unknown"
	x.Meaning = "no standard meaning; the program defines it"
	x.Hints = []string{"Check the program's documentation or search its source for the code."}
	if code > 128 {
		x.Hints = append(x.Hints, fmt.Sprintf("If it came from a signal, %d is not a known signal number on %s.", code-128, osName))
	}
	return x
}
