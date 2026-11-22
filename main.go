// Command whyexit explains process exit codes in plain English.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const usage = `usage: whyexit [flags] CODE

Explains a process exit code in plain English.

flags:
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	data, err := loadData()
	if err != nil {
		fmt.Fprintln(stderr, "whyexit:", err)
		return 1
	}

	fs := flag.NewFlagSet("whyexit", flag.ContinueOnError)
	fs.SetOutput(stderr)
	osName := fs.String("os", defaultOS(), "signal numbering to use: linux or macos")
	context := fs.String("context", "", "tool the code came from: "+strings.Join(data.ContextIDs(), ", "))
	asJSON := fs.Bool("json", false, "print JSON")
	fs.Usage = func() {
		fmt.Fprint(stderr, usage)
		fs.PrintDefaults()
	}
	pos, err := parseInterspersed(fs, args)
	if err != nil {
		return 2
	}
	if *osName != "linux" && *osName != "macos" {
		fmt.Fprintf(stderr, "whyexit: unknown --os %q (want linux or macos)\n", *osName)
		return 2
	}
	if *context != "" && data.Context(*context) == nil {
		fmt.Fprintf(stderr, "whyexit: unknown --context %q (want %s)\n", *context, strings.Join(data.ContextIDs(), ", "))
		return 2
	}
	if len(pos) != 1 {
		fs.Usage()
		return 2
	}
	code, err := strconv.Atoi(pos[0])
	if err != nil {
		fmt.Fprintf(stderr, "whyexit: %q is not a number\n", pos[0])
		return 2
	}
	x := data.Explain(code&0xff, *osName, *context)
	if code < 0 || code > 255 {
		x.Note = fmt.Sprintf("exit statuses are 8 bits, so %d is seen as %d", code, code&0xff)
	}
	if *asJSON {
		return writeJSON(stdout, stderr, x)
	}
	fmt.Fprint(stdout, format(x))
	return 0
}

// parseInterspersed lets flags appear after the code, as in "whyexit 137 --json".
func parseInterspersed(fs *flag.FlagSet, args []string) ([]string, error) {
	var pos []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		if fs.NArg() == 0 {
			return pos, nil
		}
		pos = append(pos, fs.Arg(0))
		args = fs.Args()[1:]
	}
}

func writeJSON(stdout, stderr io.Writer, v any) int {
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintln(stderr, "whyexit:", err)
		return 1
	}
	return 0
}

func headline(x Explanation) string {
	var b strings.Builder
	if x.Context != "" && x.Source == x.Context {
		b.WriteString(x.Context + ": ")
	}
	b.WriteString(strconv.Itoa(x.Code))
	if x.Signal != nil {
		fmt.Fprintf(&b, " = 128 + %d (%s)", x.Signal.Number, x.Signal.Name)
	} else if x.Name != "" {
		fmt.Fprintf(&b, " (%s)", x.Name)
	}
	b.WriteString(": ")
	b.WriteString(x.Meaning)
	return b.String()
}

func format(x Explanation) string {
	var b strings.Builder
	b.WriteString(headline(x))
	b.WriteByte('\n')
	if x.Note != "" {
		b.WriteString("  Note: " + x.Note + ".\n")
	}
	for _, h := range x.Hints {
		b.WriteString("  ")
		b.WriteString(h)
		b.WriteByte('\n')
	}
	return b.String()
}
