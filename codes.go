package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed codes.json
var codesJSON []byte

// Entry is one exit code in a table.
type Entry struct {
	Code    int      `json:"code"`
	Name    string   `json:"name,omitempty"`
	Meaning string   `json:"meaning"`
	Hints   []string `json:"hints"`
}

// Table groups the codes of one source, such as the shell or a tool.
type Table struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Codes []Entry `json:"codes"`
}

// Data is the parsed contents of codes.json.
type Data struct {
	Tables []Table `json:"tables"`
}

// Explanation is the answer for a single exit code.
type Explanation struct {
	Code    int      `json:"code"`
	Source  string   `json:"source"`
	Name    string   `json:"name,omitempty"`
	Meaning string   `json:"meaning"`
	Hints   []string `json:"hints"`
}

func loadData() (*Data, error) {
	var d Data
	if err := json.Unmarshal(codesJSON, &d); err != nil {
		return nil, fmt.Errorf("codes.json: %w", err)
	}
	return &d, nil
}

func (t *Table) find(code int) *Entry {
	for i := range t.Codes {
		if t.Codes[i].Code == code {
			return &t.Codes[i]
		}
	}
	return nil
}

// Explain looks code up in the general tables.
func (d *Data) Explain(code int) Explanation {
	for i := range d.Tables {
		if e := d.Tables[i].find(code); e != nil {
			return Explanation{Code: code, Source: d.Tables[i].ID, Name: e.Name, Meaning: e.Meaning, Hints: e.Hints}
		}
	}
	return Explanation{
		Code:    code,
		Source:  "unknown",
		Meaning: "no standard meaning; the program defines it",
		Hints:   []string{"Check the program's documentation or search its source for the code."},
	}
}
