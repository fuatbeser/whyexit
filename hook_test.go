package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain lets the test binary stand in for whyexit when the hook
// scripts under test call it.
func TestMain(m *testing.M) {
	if os.Getenv("WHYEXIT_TEST_MAIN") == "1" {
		main()
		return
	}
	dir, err := os.MkdirTemp("", "whyexit-test")
	if err != nil {
		panic(err)
	}
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	if err := os.Symlink(exe, filepath.Join(dir, "whyexit")); err != nil {
		panic(err)
	}
	os.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestHookScript(t *testing.T) {
	d := mustLoad(t)
	tests := []struct {
		shell, osName string
		interrupt     bool
		want          string
	}{
		{"bash", "linux", false, "0|130|148)"},
		{"bash", "linux", true, "0|148)"},
		{"bash", "macos", false, "0|130|146)"},
		{"zsh", "linux", false, "0|130|148)"},
		{"zsh", "macos", true, "0|146)"},
		{"fish", "linux", false, "case 0 130 148\n"},
		{"fish", "macos", true, "case 0 146\n"},
	}
	for _, tt := range tests {
		script, err := hookScript(d, tt.shell, tt.osName, tt.interrupt)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(script, tt.want) {
			t.Errorf("%s/%s interrupt=%v: script lacks %q", tt.shell, tt.osName, tt.interrupt, tt.want)
		}
	}
	if _, err := hookScript(d, "tcsh", "linux", false); err == nil {
		t.Error("tcsh: want error")
	}
}

func TestHookSyntax(t *testing.T) {
	d := mustLoad(t)
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			bin := lookShell(t, shell)
			script, err := hookScript(d, shell, "linux", false)
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(bin, "-n")
			cmd.Stdin = strings.NewReader(script)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("%v\n%s", err, out)
			}
		})
	}
}

// Each case loads the hook, fakes one command that exits with a status,
// calls the hook function the way the shell would before the next
// prompt, and prints the status the prompt would see.
func TestHookBehavior(t *testing.T) {
	run := map[string]string{
		"bash": `eval "$(whyexit init bash %s)"; history -s %q; _whyexit_ran=1; (exit %d); _whyexit_hook; echo "status=$?"`,
		"zsh":  `eval "$(whyexit init zsh %s)"; _whyexit_preexec %q; (exit %d); _whyexit_precmd; echo "status=$?"`,
		"fish": `whyexit init fish %s | source; set -l cmd %q; sh -c 'exit %d'; __whyexit_postexec $cmd; echo "status=$status"`,
	}
	tests := []struct {
		name    string
		flag    string
		cmdline string
		code    int
		want    string
	}{
		{"success is silent", "", "true", 0, ""},
		{"failure", "", "false", 1, "exit 1: general error"},
		{"not found", "", "nosuchcmd", 127, "exit 127: command not found"},
		{"signal", "", "sleep 100", 137, "exit 137 = 128 + 9 (SIGKILL)"},
		{"context from command", "", "curl -s localhost:1", 7, "curl exit 7 (CURLE_COULDNT_CONNECT)"},
		{"context after sudo", "", "sudo FOO=1 docker run x", 125, "docker exit 125:"},
		{"ctrl-c is silent", "", "sleep 100", 130, ""},
		{"ctrl-c with flag", "--show-interrupt", "sleep 100", 130, "exit 130 = 128 + 2 (SIGINT)"},
		{"ctrl-z is silent", "--show-interrupt", "vim", 128 + mustLoad(t).signalNumber("SIGTSTP", defaultOS()), ""},
	}
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			bin := lookShell(t, shell)
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					src := sprintf(run[shell], tt.flag, tt.cmdline, tt.code)
					cmd := exec.Command(bin, "-c", src)
					// GORACE keeps -race builds from sleeping a second on exit.
					cmd.Env = append(os.Environ(), "WHYEXIT_TEST_MAIN=1", "NO_COLOR=1", "GORACE=atexit_sleep_ms=0")
					var stderr strings.Builder
					cmd.Stderr = &stderr
					out, err := cmd.Output()
					if err != nil {
						t.Fatalf("%v\n%s", err, stderr.String())
					}
					if got, want := strings.TrimSpace(string(out)), "status="+itoa(tt.code); got != want {
						t.Errorf("prompt sees %q, want %q", got, want)
					}
					hint := strings.TrimSpace(stderr.String())
					if tt.want == "" && hint != "" {
						t.Errorf("want no hint, got %q", hint)
					}
					if tt.want != "" && (!strings.HasPrefix(hint, tt.want) || strings.Contains(hint, "\n")) {
						t.Errorf("hint = %q, want one line starting with %q", hint, tt.want)
					}
				})
			}
		})
	}
}

func TestContextFor(t *testing.T) {
	d := mustLoad(t)
	tests := map[string]string{
		"docker run --rm x":         "docker",
		"/usr/local/bin/curl -sf x": "curl",
		"sudo -E kubectl apply":     "",
		"time npx jest":             "npm",
		"LANG=C git push":           "git",
		"ssh host true":             "ssh",
		"make test":                 "",
		"":                          "",
	}
	for cmdline, want := range tests {
		if got := d.contextFor(cmdline); got != want {
			t.Errorf("contextFor(%q) = %q, want %q", cmdline, got, want)
		}
	}
}

func lookShell(t *testing.T, name string) string {
	t.Helper()
	bin, err := exec.LookPath(name)
	if err != nil {
		t.Skipf("%s not installed", name)
	}
	return bin
}
