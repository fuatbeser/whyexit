package main

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

func itoa(n int) string { return strconv.Itoa(n) }

func runCLI(args ...string) (stdout, stderr string, status int) {
	var out, errb bytes.Buffer
	status = run(args, &out, &errb)
	return out.String(), errb.String(), status
}

func TestRun(t *testing.T) {
	tests := []struct {
		args   []string
		status int
		want   string
	}{
		{[]string{"127"}, 0, "127: command not found\n"},
		{[]string{"1"}, 0, "1: general error"},
		{[]string{"42"}, 0, "42: no standard meaning"},
		{[]string{"137"}, 0, "137 = 128 + 9 (SIGKILL): killed by force, often by the OOM killer\n"},
		{[]string{"130"}, 0, "130 = 128 + 2 (SIGINT): interrupted with Ctrl+C\n"},
		{[]string{"--os", "linux", "138"}, 0, "138 = 128 + 10 (SIGUSR1)"},
		{[]string{"--os", "macos", "138"}, 0, "138 = 128 + 10 (SIGBUS)"},
		{[]string{"--os", "windows", "1"}, 2, ""},
		{[]string{"200"}, 0, "200: no standard meaning"},
		{[]string{"64"}, 0, "64 (EX_USAGE): command line usage error\n"},
		{[]string{"78"}, 0, "78 (EX_CONFIG): configuration error\n"},
		{[]string{"256"}, 0, "0: success\n  Note: exit statuses are 8 bits, so 256 is seen as 0.\n"},
		{[]string{"--", "-1"}, 0, "255: exit status out of range"},
		{[]string{"127", "--os", "macos"}, 0, "127: command not found"},
		{[]string{"1", "2"}, 2, ""},
		{[]string{"--context", "docker", "137"}, 0, "docker: 137 = 128 + 9 (SIGKILL): the container was killed"},
		{[]string{"--context", "docker", "--os", "macos", "143"}, 0, "docker: 143 = 128 + 15 (SIGTERM)"},
		{[]string{"--context", "docker", "2"}, 0, "2: misuse"},
		{[]string{"--context", "k8s", "137"}, 0, "k8s: 137 = 128 + 9 (SIGKILL)"},
		{[]string{"--context", "nope", "1"}, 2, ""},
		{[]string{"abc"}, 2, ""},
		{[]string{}, 2, ""},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			out, _, status := runCLI(tt.args...)
			if status != tt.status {
				t.Errorf("status = %d, want %d", status, tt.status)
			}
			if !strings.HasPrefix(out, tt.want) {
				t.Errorf("output = %q, want prefix %q", out, tt.want)
			}
		})
	}
}

func TestRunJSON(t *testing.T) {
	out, _, status := runCLI("137", "--json", "--os", "linux")
	if status != 0 {
		t.Fatalf("status = %d", status)
	}
	var x Explanation
	if err := json.Unmarshal([]byte(out), &x); err != nil {
		t.Fatal(err)
	}
	if x.Code != 137 || x.Source != "signal" || x.OS != "linux" || x.Signal == nil || x.Signal.Name != "SIGKILL" || x.Signal.Number != 9 {
		t.Errorf("got %+v", x)
	}
	if len(x.Hints) == 0 {
		t.Error("no hints")
	}
}
