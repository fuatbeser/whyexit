package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
)

//go:embed hooks
var hooks embed.FS

var shells = []string{"bash", "fish", "zsh"}

// runInit prints the hook script for a shell.
func runInit(data *Data, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("whyexit init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showInterrupt := fs.Bool("show-interrupt", false, "also explain commands stopped with Ctrl+C")
	fs.Usage = func() {
		fmt.Fprintf(stderr, "usage: whyexit init [--show-interrupt] %s\n", strings.Join(shells, "|"))
	}
	pos, err := parseInterspersed(fs, args)
	if err != nil {
		return 2
	}
	if len(pos) != 1 {
		fs.Usage()
		return 2
	}
	script, err := hookScript(data, pos[0], defaultOS(), *showInterrupt)
	if err != nil {
		fmt.Fprintln(stderr, "whyexit:", err)
		return 2
	}
	fmt.Fprint(stdout, script)
	return 0
}

func hookScript(data *Data, shell, osName string, showInterrupt bool) (string, error) {
	src, err := hooks.ReadFile(path.Join("hooks", shell))
	if err != nil {
		return "", fmt.Errorf("unsupported shell %q (want %s)", shell, strings.Join(shells, ", "))
	}
	tmpl, err := template.New(shell).Parse(string(src))
	if err != nil {
		return "", err
	}
	// A job stopped with Ctrl+Z is not a failure, so it never gets a hint.
	skip := []int{128 + data.signalNumber("SIGTSTP", osName)}
	if !showInterrupt {
		skip = append([]int{128 + data.signalNumber("SIGINT", osName)}, skip...)
	}
	var b strings.Builder
	if err := tmpl.Execute(&b, struct{ Skip []int }{skip}); err != nil {
		return "", err
	}
	return b.String(), nil
}

// runHint prints the one-line hint the shell hooks show after a failure.
func runHint(data *Data, args []string, stderr io.Writer) int {
	if len(args) == 0 {
		return 2
	}
	code, err := strconv.Atoi(args[0])
	if err != nil {
		return 2
	}
	cmdline := strings.Join(args[1:], " ")
	x := data.Explain(code&0xff, defaultOS(), data.contextFor(cmdline))
	line := hintLine(x)
	if useColor(os.Stderr) {
		line = "\x1b[2m" + line + "\x1b[0m"
	}
	fmt.Fprintln(stderr, line)
	return 0
}

func hintLine(x Explanation) string {
	if x.Context != "" && x.Source == x.Context {
		return x.Context + " exit " + describe(x)
	}
	return "exit " + describe(x)
}

func useColor(f *os.File) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// contextFor guesses the context from the command line that failed,
// skipping variable assignments and wrappers like sudo.
func (d *Data) contextFor(cmdline string) string {
	for _, word := range strings.Fields(cmdline) {
		switch {
		case strings.Contains(word, "="):
			continue
		case word == "sudo" || word == "env" || word == "time" || word == "command" || word == "exec" || word == "nohup":
			continue
		}
		name := path.Base(word)
		for _, t := range d.Contexts {
			for _, c := range t.Commands {
				if c == name {
					return t.ID
				}
			}
		}
		return ""
	}
	return ""
}
