// Command whyexit explains process exit codes in plain English.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("whyexit", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: whyexit CODE")
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return 2
	}
	code, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "whyexit: %q is not a number\n", fs.Arg(0))
		return 2
	}

	data, err := loadData()
	if err != nil {
		fmt.Fprintln(stderr, "whyexit:", err)
		return 1
	}
	fmt.Fprint(stdout, format(data.Explain(code)))
	return 0
}

func headline(x Explanation) string {
	var b strings.Builder
	b.WriteString(strconv.Itoa(x.Code))
	if x.Name != "" {
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
	for _, h := range x.Hints {
		b.WriteString("  ")
		b.WriteString(h)
		b.WriteByte('\n')
	}
	return b.String()
}
